package client

import (
	"context"
	"github.com/spf13/viper"
	"testing"
	"time"
)

func TestInitConfig(t *testing.T) {
	InitConfig("config", "../", "yaml")
	port := viper.GetString("recv_grpc.port")
	if port != "50051" {
		t.Errorf("got [%s] expect [%s]", port, "port")
	}
}

func TestWaitShutdown(t *testing.T) {
	ctxDone := new(bool)
	ctx, cancel := context.WithCancel(context.Background())
	go func(ctx context.Context, ctxDone *bool) {
		select {
		case <-ctx.Done():
			*ctxDone = true
		}
	}(ctx, ctxDone)
	time.Sleep(time.Second)
	cancel()
	time.Sleep(time.Second)
	if !*ctxDone{
		t.Errorf("got [%v] expect [%v]", *ctxDone, true)
	}
	t.Log(*ctxDone)
}
