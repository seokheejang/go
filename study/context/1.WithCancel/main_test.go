package main

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestPrintTick(t *testing.T) {
	// 테스트에서 사용되는 WaitGroup
	var wg sync.WaitGroup

	// WaitGroup의 카운트를 1로 설정
	wg.Add(1)

	// 테스트를 위한 context 생성
	ctx, cancel := context.WithCancel(context.Background())

	// PrintTick 함수 실행
	go func() {
		PrintTick(ctx, &wg)
	}()

	// 2초 후 context를 취소하여 PrintTick 함수를 종료
	time.Sleep(2 * time.Second)
	cancel()

	// PrintTick 함수가 종료될 때까지 대기
	wg.Wait()

	// 종료 메시지가 출력되었는지 확인 (이 부분은 수동 확인이 필요)
	t.Log("PrintTick function completed")
}
