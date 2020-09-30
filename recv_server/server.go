package recv_server

import (
	"chat_room/glog"
	"chat_room/mq"
	"chat_room/proto"
	"chat_room/redis"
	"flag"
	"fmt"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func Execute() {
	flag.Parse()

	initConfig()

	redis.InitRedis(
		viper.GetString("redis.host"),
		viper.GetString("redis.port"),
		viper.GetString("redis.passwd"),
	)

	rabbitUrl := viper.GetString("rabbit.protocol") + "://" + viper.GetString("rabbit.username") + ":" + viper.GetString("rabbit.password") + "@" + viper.GetString("rabbit.ip") + ":" + viper.GetString("rabbit.port")

	//获取配置，监听端口
	addr := viper.GetString("recv_grpc.ip") + ":" + viper.GetString("recv_grpc.port")
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		panic("failed to listen: " + err.Error())
	}

	s := grpc.NewServer()

	mc, e := mq.NewMsgQueue(rabbitUrl)
	if e != nil {
		panic(e)
	}
	proto.RegisterRecvServer(s, NewRpcServer(mc))
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	go func() {
		coon, e := lis.Accept()
		if e != nil {
			panic(e)
		}
		glog.Infoln(coon.RemoteAddr())
	}()

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
