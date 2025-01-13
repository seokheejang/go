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
	// 서버 주소
	serverAddress := "localhost:50051"

	// gRPC 연결 생성 (TLS를 사용하지 않음)
	conn, err := grpc.DialContext(context.Background(), serverAddress, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	// 클라이언트 생성
	client := pb.NewInputServiceClient(conn)

	// 사용자 입력 대기
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

		// 유효한 키인지 확인
		if input != "w" && input != "a" && input != "s" && input != "d" {
			fmt.Println("Invalid input. Please enter 'w', 'a', 's', 'd', or 'exit'.")
			continue
		}

		// 서버에 키 전송
		req := &pb.KeyRequest{Key: input}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		res, err := client.SendKey(ctx, req)
		if err != nil {
			log.Printf("Failed to send key '%s': %v", input, err)
			continue
		}

		// 서버 응답 출력
		fmt.Printf("Server response: %s\n", res.Message)
	}

	// 사용자 입력 종료 처리
	if err := scanner.Err(); err != nil {
		log.Printf("Error reading input: %v", err)
	}
}