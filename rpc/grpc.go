package rpc

import (
	"chat_room/proto"
	"google.golang.org/grpc"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func NewGrpcRecvServer(addr string,srv proto.RecvServer){
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		panic("failed to listen: " + err.Error())
	}
	s := grpc.NewServer()
	proto.RegisterRecvServer(s, srv)

	server(lis,s)
	return
}


func NewGrpcSendServer(addr string,srv proto.SendServer){
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		panic("failed to listen: " + err.Error())
	}
	s := grpc.NewServer()
	proto.RegisterSendServer(s, srv)

	server(lis,s)
	return
}

func server(lis net.Listener,s *grpc.Server){

	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch,
			// kill -SIGINT XXXX or Ctrl+c
			os.Interrupt,
			syscall.SIGINT, // register that too, it should be ok
			// os.Kill  is equivalent with the syscall.Kill
			os.Kill,
			syscall.SIGKILL, // register that too, it should be ok
			// kill -SIGTERM XXXX
			syscall.SIGTERM,
		)
		select {
		case <-ch:
			println("shutdown...")
			s.GracefulStop()
			lis.Close()
		}
	}()

	if err := s.Serve(lis); err != nil {
		panic("failed to serve: "+ err.Error())
	}
}
