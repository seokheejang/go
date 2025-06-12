package config

import (
	"time"
)

type Config struct {
	Database DatabaseConfig
	Redis    RedisConfig
	Cache    CacheConfig
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type CacheConfig struct {
	Type   string
	TTL    time.Duration
	MaxTTL time.Duration
}

func NewDefaultConfig() *Config {
	return &Config{
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     6432,
			User:     "postgres",
			Password: "postgres",
			DBName:   "postgres",
			SSLMode:  "disable",
		},
		Redis: RedisConfig{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		},
		Cache: CacheConfig{
			Type:   "mem",
			TTL:    2 * time.Second,
			MaxTTL: 30 * time.Second,
		},
	}
}
