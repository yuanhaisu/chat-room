package client

import (
	"bufio"
	"chat_room/common"
	"chat_room/glog"
	"chat_room/proto"
	"context"
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

var (
	UserName    string
	inputReader *bufio.Reader
)

func Execute() {

	ctx := context.Background()
	initConfig()
	//初始化grpc
	sendAdddr := viper.GetString("send_grpc.ip") + ":" + viper.GetString("send_grpc.port")
	recvAdddr := viper.GetString("recv_grpc.ip") + ":" + viper.GetString("recv_grpc.port")
	communicator := NewCommunicator(sendAdddr, recvAdddr)

	inputReader = bufio.NewReader(os.Stdin)
	fmt.Println("Please input your name:")
	var (
		joinMsg *proto.Request
	)
	for {
		input, err := inputReader.ReadString('\n')
		if err != nil {
			fmt.Printf("An error occurred :%s\n", err)
			os.Exit(1)
		} else {
			UserName = input[:len(input)-2]
			joinMsg, err = NewMsgRequest(UserName, true)
			if err != nil {
				continue
			}
			break
		}
	}
	//此处启动
	err := communicator.Launch(ctx, joinMsg)
	if err != nil {
		panic(err)
	}
	defer communicator.Close()

	go InputWindow(communicator)

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
		//if UserName != "" {
		//	communicator.Send(&proto.Request{
		//		From:    UserName,
		//		Content: UserName + "下线了",
		//		To:      "",
		//		Time:    time.Now().Format("2006-01-02 15:04:05"),
		//	})
		//}
	}
}

func InputWindow(communicator Communicator) {
	for {
		input, err := inputReader.ReadString('\n')
		if err != nil {
			fmt.Printf("An error occurred :%s\n", err)
			continue
		}
		input = input[:len(input)-2]
		if input == "" {
			continue
		}
		msg, err := NewMsgRequest(input)
		if err != nil {
			fmt.Println(err)
			continue
		}
		communicator.Send(msg)
	}
}

type Communicator interface {
	Launch(ctx context.Context, req *proto.Request) error
	Send(*proto.Request)
	Close()
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

func (gc *grpcCommunicator) Close() {
	gc.sendConn.Close()
	gc.recvConn.Close()
}

func (gc *grpcCommunicator) Launch(c context.Context, req *proto.Request) (err error) {
	defer func(err error) {
		if err != nil {
			if gc.sendConn != nil {
				gc.sendConn.Close()
			}
			if gc.recvConn != nil {
				gc.recvConn.Close()
			}
		}
	}(err)
	gc.sendConn, err = grpc.Dial(gc.RecvServerAddr, grpc.WithInsecure(), grpc.WithInsecure())
	if err != nil {
		return
	}

	gc.recvConn, err = grpc.Dial(gc.SendServerAddr, grpc.WithInsecure(), grpc.WithInsecure())
	if err != nil {
		return
	}

	gc.recvClient = proto.NewSendClient(gc.recvConn)
	gc.sendClient = proto.NewRecvClient(gc.sendConn)

	//接收服务端发过来的流
	gc.recvStream, err = gc.recvClient.UserSendStream(c, req)
	if err != nil {
		return
	}
	//发送流给服务端
	gc.sendStream, err = gc.sendClient.UserRecvStream(c)
	if err != nil {
		return
	}

	go func(c context.Context) {
		for {
			select {
			case <-c.Done():
				return
			default:
				data, e := gc.recvStream.Recv()
				if e != nil {
					if e == io.EOF {
						break
					}
					if e == common.ErrNameIsExisted {
						continue
					}
					panic(e)
				}
				glog.Infoln("收到消息：", data)
				if data.Action == common.Aite {
					fmt.Println("来自" + data.From + "的个人消息：" + data.Content)
				} else {
					fmt.Println(data.From + "：" + data.Content)
				}
			}
		}
	}(c)

	go func(c context.Context) {
		for {
			select {
			case req := <-gc.chatMsgSendChan:
				if e := gc.sendStream.Send(req); e != nil {
					if e == io.EOF {
						fmt.Println("与房间连接已断开")
						return
					}
					fmt.Println("发送失败：", e)
				}
			case <-c.Done():
				return
			}
		}
	}(c)
	return
}

func (gc *grpcCommunicator) Send(req *proto.Request) {
	gc.chatMsgSendChan <- req
}

func NewMsgRequest(input string, isJoin ...bool) (req *proto.Request, err error) {

	req = &proto.Request{
		From: UserName,
		Time: time.Now().Format("2006-01-02 15:04:05"),
	}
	if len(isJoin) > 0 && isJoin[0] {
		req.Action = common.Join
	}
	if input[:1] == "@" {
		ss := strings.Split(input, " ")
		if len(ss) == 1 || len(ss[0]) == 1 {
			err = errors.New("请求格式错误，@昵称+空格+消息内容")
			return
		}

		req.Action = common.Aite
		req.Content = common.StrSlice2Str(ss[1:], " ")
		req.To = ss[0][1:]
	} else {
		req.Content = input
	}
	return
}

func initConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("D:\\yuanhaisu\\goproject\\chat_room")
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
}
