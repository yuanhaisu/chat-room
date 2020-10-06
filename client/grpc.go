package client

import (
	"chat_room/proto"
	"context"
	"google.golang.org/grpc"
)

type Communicator interface {
	Launch(ctx context.Context, req *proto.Request) error
	Send(*proto.Request)
	Close()
	GetRecvStream() proto.Send_UserSendStreamClient
	GetSendStream() proto.Recv_UserRecvStreamClient
	GetChatMsgSendChan() <-chan *proto.Request
}

func NewCommunicator(sendServerAddr, recvServerAddr string) Communicator {
	return &grpcCommunicator{
		SendServerAddr:  sendServerAddr,
		RecvServerAddr:  recvServerAddr,
		chatMsgSendChan: make(chan *proto.Request),
	}
}

type grpcCommunicator struct {
	SendServerAddr  string
	RecvServerAddr  string
	sendConn        *grpc.ClientConn
	recvConn        *grpc.ClientConn
	recvClient      proto.SendClient
	sendClient      proto.RecvClient
	recvStream      proto.Send_UserSendStreamClient
	sendStream      proto.Recv_UserRecvStreamClient
	chatMsgSendChan chan *proto.Request
	Name            string
}

func (gc *grpcCommunicator) GetChatMsgSendChan() <-chan *proto.Request {
	return gc.chatMsgSendChan
}

func (gc *grpcCommunicator) Close() {
	gc.sendConn.Close()
	gc.recvConn.Close()
}

func (gc *grpcCommunicator) Launch(c context.Context, req *proto.Request) (err error) {
	if err = gc.connectSendServer(c); err != nil {
		return
	}
	if err = gc.connectRecvServer(c, req); err != nil {
		return err
	}
	return
}

func (gc *grpcCommunicator) GetRecvStream() proto.Send_UserSendStreamClient {
	return gc.recvStream
}

func (gc *grpcCommunicator) GetSendStream() proto.Recv_UserRecvStreamClient {
	return gc.sendStream
}

func (gc *grpcCommunicator) connectSendServer(c context.Context) (err error) {
	defer func() {
		if err != nil {
			if gc.sendConn != nil {
				gc.sendConn.Close()
			}
		}
	}()
	gc.sendConn, err = grpc.Dial(gc.RecvServerAddr, grpc.WithInsecure(), grpc.WithInsecure())
	if err != nil {
		return
	}
	gc.sendClient = proto.NewRecvClient(gc.sendConn)
	//发送流给服务端
	gc.sendStream, err = gc.sendClient.UserRecvStream(c)
	if err != nil {
		return
	}
	return
}

func (gc *grpcCommunicator) connectRecvServer(c context.Context, req *proto.Request) (err error) {
	defer func() {
		if err != nil {
			if gc.recvConn != nil {
				gc.recvConn.Close()
			}
		}
	}()
	gc.recvConn, err = grpc.Dial(gc.SendServerAddr, grpc.WithInsecure())
	if err != nil {
		return
	}
	gc.recvClient = proto.NewSendClient(gc.recvConn)
	//发送流给服务端
	gc.recvStream, err = gc.recvClient.UserSendStream(c, req)
	if err != nil {
		return
	}
	return
}

func (gc *grpcCommunicator) Send(req *proto.Request) {
	gc.chatMsgSendChan <- req
}
