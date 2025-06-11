package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/seokheejang/go/cache-layer/internal/domains/cache"
	"github.com/seokheejang/go/cache-layer/internal/domains/user"
	cacheGorm "github.com/seokheejang/go/cache-layer/internal/infrastructure/cache/gorm"
	"github.com/seokheejang/go/cache-layer/internal/infrastructure/cache/memory"
	redisCache "github.com/seokheejang/go/cache-layer/internal/infrastructure/cache/redis"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	if len(os.Args) == 1 {
		fmt.Println("Usage: go run main.go --cache=mem|redis")
		os.Exit(1)
	}
	cacheType := flag.String("cache", "mem", "cache type choice (mem or redis)")
	flag.Parse()

	// Configure GORM logger
	newLogger := logger.New(
		log.New(log.Writer(), "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	dsn := "host=localhost user=postgres password=postgres dbname=postgres port=6432 sslmode=disable"
	log.Printf("Connecting to PostgreSQL with DSN: %s", dsn)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	if err := db.Exec("CREATE SCHEMA IF NOT EXISTS public").Error; err != nil {
		log.Fatal("Failed to create schema:", err)
	}

	if err := db.Exec("SET search_path TO public").Error; err != nil {
		log.Fatal("Failed to set search path:", err)
	}

	var cacheService cache.Service
	switch *cacheType {
	case "mem":
		log.Println("Using in-memory cache")
		cacheService = memory.NewInMemoryCache(2*time.Second, 30*time.Second)
	case "redis":
		log.Println("Using Redis cache")
		rdb := redis.NewClient(&redis.Options{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		})
		ttl := 2 * time.Second
		cacheService = redisCache.NewRedisCache(rdb, &ttl)
	default:
		log.Fatal("Invalid cache type. Use 'mem' or 'redis'")
	}
	defer cacheService.Close()

	// Apply GORM cache plugin
	if err := cacheGorm.WithGormCache(db, cacheService); err != nil {
		log.Fatal(err)
	}

	// Run migrations
	if err := db.AutoMigrate(&user.UserRole{}, &user.User{}); err != nil {
		log.Fatal(err)
	}

	// Clean up existing data
	if err := db.Exec("DELETE FROM users").Error; err != nil {
		log.Fatal(err)
	}
	if err := db.Exec("DELETE FROM user_roles").Error; err != nil {
		log.Fatal(err)
	}

	// Create roles
	adminRole := &user.UserRole{
		Name: "Admin",
	}
	if err := db.Create(adminRole).Error; err != nil {
		log.Fatal(err)
	}

	guestRole := &user.UserRole{
		Name: "Guest",
	}
	if err := db.Create(guestRole).Error; err != nil {
		log.Fatal(err)
	}

	// Create users
	adminUser := &user.User{
		Name:   "admin",
		RoleID: adminRole.ID,
		Role:   adminRole,
	}
	if err := db.Create(adminUser).Error; err != nil {
		log.Fatal(err)
	}

	guestUser := &user.User{
		Name:   "guest",
		RoleID: guestRole.ID,
		Role:   guestRole,
	}
	if err := db.Create(guestUser).Error; err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	fmt.Println("\n=== Starting Cache Test ===")

	fmt.Println("\n[Test 1] First query (expected cache miss)")
	var users1 []user.User
	if err := db.WithContext(ctx).Find(&users1, "name = ?", "admin").Error; err != nil {
		log.Fatal(err)
	}
	if len(users1) > 0 {
		fmt.Printf("Result: %+v\n", users1[0])
	}

	// Second query (cache hit)
	fmt.Println("\n[Test 2] Second query (expected cache hit)")
	var users2 []user.User
	if err := db.WithContext(ctx).Find(&users2, "name = ?", "admin").Error; err != nil {
		log.Fatal(err)
	}
	if len(users2) > 0 {
		fmt.Printf("Result: %+v\n", users2[0])
	}

	fmt.Println("\n=== Cache Test Completed ===")
}
