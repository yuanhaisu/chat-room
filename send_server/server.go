package send_server

import (
	"chat_room/mq"
	"chat_room/proto"
	"chat_room/redis"
	"context"
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
	ctx, cancle := context.WithCancel(context.Background())
	defer cancle()

	initConfig()

	redis.InitRedis(
		viper.GetString("redis.host"),
		viper.GetString("redis.port"),
		viper.GetString("redis.passwd"),
	)

	rabbitUrl := viper.GetString("rabbit.protocol") + "://" + viper.GetString("rabbit.username") + ":" + viper.GetString("rabbit.password") + "@" + viper.GetString("rabbit.ip") + ":" + viper.GetString("rabbit.port")

	//获取配置，监听端口
	addr := viper.GetString("send_grpc.ip") + ":" + viper.GetString("send_grpc.port")
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		panic("failed to listen: " + err.Error())
	}

	s := grpc.NewServer()

	rm, e := mq.NewMsgQueue(rabbitUrl)
	if e != nil {
		panic(e)
	}
	mc := NewMsgCenter(rm)
	proto.RegisterSendServer(s, NewRpcServer(mc, redis.RedisSource))

	go mc.Delivery(ctx)

	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
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
	viper.AddConfigPath("./")
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
}
