package main

import (
	"context"
	"log"
	"net"

	pb "grpc-wasd-server/proto"

	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedInputServiceServer
}

func (s *server) SendKey(ctx context.Context, req *pb.KeyRequest) (*pb.KeyResponse, error) {
	log.Printf("Received key: %s", req.Key)
	return &pb.KeyResponse{Message: "Key received: " + req.Key}, nil
}

func main() {
	// 기본 포트로 서버를 리스닝
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// gRPC 서버를 생성
	grpcServer := grpc.NewServer()

	// 서비스 등록
	pb.RegisterInputServiceServer(grpcServer, &server{})

	log.Println("gRPC server is running on port 50051")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}