package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

func main() {
	url := "http://localhost:7878/health"

	limit := 5 * 1024 * 1024
	for _, N := range []int{limit - 1000, limit + 1000} {
		fmt.Printf("\nSending request with %d bytes...\n", N)
		sendRequest(url, N)
	}
}

// HTTP POST 요청을 보내는 함수
func sendRequest(url string, N int) {
	data := make([]byte, N)
	for i := range data {
		data[i] = 'A'
	}

	resp, err := http.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Println("Request Body (first 500 bytes):", string(data[:min(len(data), 500)]))
	fmt.Println("Response Status:", resp.Status)
	fmt.Println("Response Body:", string(body))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
