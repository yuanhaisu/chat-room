package client

import (
	"chat_room/proto"
	"context"
	"fmt"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
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
	if !*ctxDone {
		t.Errorf("got [%v] expect [%v]", *ctxDone, true)
	}
	t.Log(*ctxDone)
}

type testCommunicator struct {
	tcCh               chan *proto.Request
	sendUserSendStream chan struct{}
}

func NewTestCommunicator(ch chan *proto.Request, sendUserSendStream chan struct{}) Communicator {
	return &testCommunicator{
		tcCh:               ch,
		sendUserSendStream: sendUserSendStream,
	}
}

func (tc *testCommunicator) Launch(ctx context.Context, req *proto.Request) error {
	return nil
}
func (tc *testCommunicator) Send(req *proto.Request) {
	fmt.Println("body")
}
func (tc *testCommunicator) Close()              {}
func (tc *testCommunicator) GetRecvStream() proto.Send_UserSendStreamClient {
	return &sendUserSendStreamClient{
		ch: tc.sendUserSendStream,
	}
}
func (tc *testCommunicator) GetSendStream() proto.Recv_UserRecvStreamClient {
	return &recvUserRecvStreamClient{}
}
func (tc *testCommunicator) GetChatMsgSendChan() <-chan *proto.Request {
	return tc.tcCh
}

type recvUserRecvStreamClient struct {
	grpc.ClientStream
}

func (x *recvUserRecvStreamClient) Send(m *proto.Request) error {
	//fmt.Println("发送消息成功")
	return nil
}

func (x *recvUserRecvStreamClient) CloseAndRecv() (*proto.Request, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(proto.Request)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

type sendUserSendStreamClient struct {
	grpc.ClientStream
	ch chan struct{}
}

func (x *sendUserSendStreamClient) Recv() (*proto.Request, error) {
	m := &proto.Request{
		From:    "发送者",
		Content: "消息内容",
		To:      "我",
		Time:    "",
		Action:  0,
	}
	<-x.ch
	return m, nil
}

func TestInitShowWindow(t *testing.T) {
	ctx := context.Background()
	ch := make(chan *proto.Request)
	sendUserSendStreamCh := make(chan struct{})
	gc := NewTestCommunicator(ch,sendUserSendStreamCh)
	go InitShowWindow(ctx, gc)
	for i := 0; i < 5; i++ {
		sendUserSendStreamCh <- struct{}{}
	}
}

func TestReadyForSend(t *testing.T) {
	ctx := context.Background()
	ch := make(chan *proto.Request)
	sendUserSendStreamCh := make(chan struct{})
	gc := NewTestCommunicator(ch,sendUserSendStreamCh)
	go ReadyForSend(ctx,gc)
	for i:=0;i<5;i++{
		ch <- &proto.Request{
			From:    "发送者",
			Content: "消息内容",
			To:      "我",
			Time:    "",
			Action:  0,
		}
	}
}

func BenchmarkReadyForSend(b *testing.B) {
	ctx := context.Background()
	ch := make(chan *proto.Request)
	sendUserSendStreamCh := make(chan struct{})
	gc := NewTestCommunicator(ch,sendUserSendStreamCh)
	go ReadyForSend(ctx,gc)
	for i:=0;i<b.N;i++{
		ch <- &proto.Request{
			From:    "发送者",
			Content: "消息内容",
			To:      "我",
			Time:    "",
			Action:  0,
		}
	}
}
