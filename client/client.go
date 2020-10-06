package client

import (
	"bufio"
	"chat-room/common"
	"chat-room/proto"
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/spf13/viper"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

var (
	UserName       string
	inputReader    *bufio.Reader
	confFilePath   = flag.String("confFilePath", "./", "define config file info")
	confFileName   = flag.String("confFileName", "config", "define config file info")
	confFileFormat = flag.String("confFileFormat", "yaml", "define config file info")
)

func Execute() {
	ctx, cancel := context.WithCancel(context.Background())

	InitConfig(*confFileName, *confFilePath, *confFileFormat)

	inputReader = bufio.NewReader(os.Stdin)
	joinMsg := WhatYourName()
	//此处启动

	sendAdddr := viper.GetString("send_grpc.ip") + ":" + viper.GetString("send_grpc.port")
	recvAdddr := viper.GetString("recv_grpc.ip") + ":" + viper.GetString("recv_grpc.port")
	communicator := NewCommunicator(sendAdddr, recvAdddr)
	err := communicator.Launch(ctx, joinMsg)
	if err != nil {
		panic(err)
	}
	defer communicator.Close()

	go ReadyForSend(ctx, communicator)
	go InputWindow(ctx, communicator)
	go InitShowWindow(ctx, communicator)

	WaitShutdown(cancel)
}

func WaitShutdown(cancel context.CancelFunc) (ch chan os.Signal) {

	ch = make(chan os.Signal, 1)
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
		cancel()
		time.Sleep(time.Second * 5)
		//if UserName != "" {
		//	communicator.Send(&proto.Request{
		//		From:    UserName,
		//		Content: UserName + "下线了",
		//		To:      "",
		//		Time:    time.Now().Format("2006-01-02 15:04:05"),
		//	})
		//}
	}
	return
}

func WhatYourName() (joinMsg *proto.Request) {
	fmt.Println("Please input your name:")
	for {
		input, err := inputReader.ReadString('\n')
		if err != nil {
			fmt.Printf("An error occurred :%s\n", err)
			os.Exit(1)
		} else {
			UserName = input[:len(input)-1]
			joinMsg, err = NewMsgRequest(UserName, true)
			if err != nil {
				continue
			}
			break
		}
	}
	return
}

func InputWindow(ctx context.Context, communicator Communicator) {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("input window closed!")
			return
		default:
			input, err := inputReader.ReadString('\n')
			if err != nil {
				fmt.Printf("An error occurred :%s\n", err)
				continue
			}
			input = input[:len(input)-1]
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
}

func InitShowWindow(c context.Context, gc Communicator) {
	for {
		select {
		case <-c.Done():
			fmt.Println("show window closed!")
			return
		case <-gc.GetRecvStream().Context().Done():
			fmt.Println("recv stream closed by server!")
			return
		default:
			data, e := gc.GetRecvStream().Recv()
			if e != nil {
				if e == io.EOF {
					break
				}
				if e == common.ErrNameIsExisted {
					continue
				}
				panic(e)
			}
			//glog.Infoln("收到消息：", data)
			MsgFormat(data)
		}
	}
}

func MsgFormat(data *proto.Request) {
	if data.Action == common.Aite {
		if data.From == UserName {
			fmt.Println(data.Time + ": 发送至" + data.From + "的个人消息：" + data.Content)
		} else {
			fmt.Println(data.Time + ": 来自" + data.From + "的个人消息：" + data.Content)
		}
	} else {
		fmt.Println(data.Time + ": " + data.From + "：" + data.Content)
	}
}

func ReadyForSend(c context.Context, gc Communicator) {
	msgCh := gc.GetChatMsgSendChan()
	for {
		select {
		case req := <-msgCh:
			if e := gc.GetSendStream().Send(req); e != nil {
				if e == io.EOF {
					fmt.Println("与房间连接已断开")
					return
				}
				fmt.Println("发送失败：", e)
			}
		case <-c.Done():
			fmt.Println("send stream closed!")
			return
		}
	}
}

func NewMsgRequest(input string, isJoin ...bool) (req *proto.Request, err error) {
	//TODO::使用消息id和ack来处理消息发送失败流程
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

func InitConfig(fileName, filePath, format string) {
	viper.SetConfigName(fileName)
	viper.SetConfigType(format)
	viper.AddConfigPath(filePath)
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
}
