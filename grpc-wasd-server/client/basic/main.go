package main

import (
	"context"
	"log"
	"time"

	pb "grpc-wasd-server/proto"

	"google.golang.org/grpc"
)

func main() {
	serverAddress := "localhost:50051"

	conn, err := grpc.Dial(serverAddress, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	client := pb.NewInputServiceClient(conn)

	keys := []string{"w", "a", "s", "d"}

	for _, key := range keys {
		req := &pb.KeyRequest{Key: key}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		res, err := client.SendKey(ctx, req)
		if err != nil {
			log.Printf("Failed to send key '%s': %v", key, err)
			continue
		}

		log.Printf("Server response: %s", res.Message)
	}
}