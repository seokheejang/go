package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"time"
)

func startRPCServer(addr string, portNum int, rpcServerBodyLimit int) (*http.Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", addr, portNum))
	if err != nil {
		return nil, err
	}

	rpcServer := http.NewServeMux()

	rpcServer.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Health check requested")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	srv := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 요청 로그 출력
			fmt.Printf("Received request: %s %s\n", r.Method, r.URL.Path)
			fmt.Printf("Content-Length: %d\n", r.ContentLength)
			fmt.Printf("Headers: %v\n", r.Header)

			// 요청 본문 로깅 (제한 크기 이하만 로깅)
			if r.ContentLength > 0 && r.ContentLength <= int64(rpcServerBodyLimit) {
				body, err := io.ReadAll(r.Body)
				if err == nil {
					if len(body) > 500 {
						fmt.Printf("Body: %s... (truncated)\n", string(body[:500]))
					} else {
						fmt.Printf("Body: %s\n", string(body))
					}
				} else {
					fmt.Println("Error reading body:", err)
				}
				r.Body.Close()
			}

			// 요청 크기 제한 검사
			if r.ContentLength > int64(rpcServerBodyLimit) {
				http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
				return
			}

			// 실제 요청 처리
			rpcServer.ServeHTTP(w, r)
		}),
		ReadTimeout:       30 * time.Second,
		ReadHeaderTimeout: 30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	go func() {
		err := srv.Serve(listener)
		if err != nil && err != http.ErrServerClosed {
			fmt.Println("Server error:", err)
		}
	}()
	return srv, nil
}

func main() {
	rpcServerBodyLimit := 5 * 1024 * 1024 // geth 5MB
	addr := "localhost"
	portNum := 7878

	_, err := startRPCServer(addr, portNum, rpcServerBodyLimit)
	if err != nil {
		fmt.Println("Failed to start server:", err)
		return
	}

	fmt.Println("Server is running on", addr, portNum)

	// 프로그램이 종료되지 않도록 대기
	select {}
}
