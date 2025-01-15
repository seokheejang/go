package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	pb "grpc-wasd-server/proto"

	"google.golang.org/grpc"
)

func main() {
	serverAddress := "localhost:50051"

	// gRPC 연결 생성 (TLS를 사용하지 않음)
	conn, err := grpc.DialContext(context.Background(), serverAddress, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	client := pb.NewInputServiceClient(conn)

	fmt.Println("Enter 'w', 'a', 's', or 'd' to send to the server. Type 'exit' to quit.")
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("Input: ")
		if !scanner.Scan() {
			log.Println("Failed to read input")
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "exit" {
			fmt.Println("Exiting...")
			break
		}

		if input != "w" && input != "a" && input != "s" && input != "d" {
			fmt.Println("Invalid input. Please enter 'w', 'a', 's', 'd', or 'exit'.")
			continue
		}

		req := &pb.KeyRequest{Key: input}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		res, err := client.SendKey(ctx, req)
		if err != nil {
			log.Printf("Failed to send key '%s': %v", input, err)
			continue
		}

		fmt.Printf("Server response: %s\n", res.Message)
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error reading input: %v", err)
	}
}