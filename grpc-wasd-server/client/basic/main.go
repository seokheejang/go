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
		// 키 입력 요청 생성
		req := &pb.KeyRequest{Key: key}

		// 서버에 요청 보내기
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		res, err := client.SendKey(ctx, req)
		if err != nil {
			log.Printf("Failed to send key '%s': %v", key, err)
			continue
		}

		// 서버 응답 출력
		log.Printf("Server response: %s", res.Message)
	}
}