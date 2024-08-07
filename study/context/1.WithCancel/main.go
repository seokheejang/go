package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

var wg sync.WaitGroup

func main() {
	wg.Add(1)
	ctx, cancel := context.WithCancel(context.Background())

	go PrintTick(ctx, &wg)

	time.Sleep(5 * time.Second)
	cancel()

	wg.Wait()
}

func PrintTick(ctx context.Context, wg *sync.WaitGroup) {
	tick := time.Tick(time.Second)
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Done:", ctx.Err())
			wg.Done()
			return
		case <-tick:
			fmt.Println("tick")
		}
	}
}
