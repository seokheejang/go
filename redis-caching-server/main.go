package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"golang.org/x/net/context"
)

var (
	rdb    *redis.Client
	ctx    = context.Background()
	expire = 30 * time.Second
)

func initRedis() {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}
	rdb = redis.NewClient(&redis.Options{
		Addr: addr,
	})
}

func saveLog(w http.ResponseWriter, r *http.Request) {
	// Parse the query parameters
	query := r.URL.Query()
	// Get the value of the msg parameter
	msg := query.Get("msg")
	if msg == "" {
		http.Error(w, "msg parameter is required", http.StatusBadRequest)
		return
	}

	key := fmt.Sprintf("log:%d", time.Now().UnixNano())
	// Save the log message with the key
	err := rdb.Set(ctx, key, msg, expire).Err()
	if err != nil {
		http.Error(w, "Failed to save log", http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "Log saved: %s\n", msg)
}

func getLogs(w http.ResponseWriter, r *http.Request) {
	// Get all keys that start with "log:"
	keys, err := rdb.Keys(ctx, "log:*").Result()
	if err != nil {
		http.Error(w, "Failed to get logs", http.StatusInternalServerError)
		return
	}

	if len(keys) == 0 {
		fmt.Fprintln(w, "No logs found")
		return
	}

	for _, key := range keys {
		log, _ := rdb.Get(ctx, key).Result()
		fmt.Fprintf(w, "%s: %s\n", key, log)
	}
}

func main() {
	initRedis()

	// Create a new router
	r := mux.NewRouter()
	// Register the saveLog handler
	r.HandleFunc("/log", saveLog).Methods("GET")
	r.HandleFunc("/logs", getLogs).Methods("GET")

	fmt.Println("Server is running on port 800")
	// Start the server
	log.Fatal(http.ListenAndServe(":8080", r))

}
