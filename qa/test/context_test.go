package test

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestValue(t *testing.T) {
    ctx := context.Background()
    process(ctx)

    ctx = context.WithValue(ctx, "traceId", "qcrao-2019")
    process(ctx)
}

func process(ctx context.Context) {
    traceId, ok := ctx.Value("traceId").(string)
    if ok {
        fmt.Printf("process over. trace_id=%s\n", traceId)
    } else {
        fmt.Printf("process over. no trace_id\n")
    }
}

func TestCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	go Perform(ctx)
	time.Sleep(10 * time.Second)
	cancel()
}

func Perform(ctx context.Context) {
    for {
		println("do something")
        select {
        case <-ctx.Done():
            // 被取消，直接返回
            return
        case <-time.After(time.Second):
            // block 1 秒钟 
        }
    }
}