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
    listener, err := net.Listen("tcp", ":50051")
    if err != nil {
        log.Fatalf("Failed to listen: %v", err)
    }

    grpcServer := grpc.NewServer()
    pb.RegisterInputServiceServer(grpcServer, &server{})

    log.Println("gRPC server is running on port 50051")
    if err := grpcServer.Serve(listener); err != nil {
        log.Fatalf("Failed to serve: %v", err)
    }
}