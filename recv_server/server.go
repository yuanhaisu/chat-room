package recv_server

import (
	"chat-room/mq"
	"chat-room/redis"
	"chat-room/rpc"
	"context"
	"flag"
	"fmt"
	"github.com/spf13/viper"
)

func Execute() {
	flag.Parse()
	ctx, cancle := context.WithCancel(context.Background())

	initConfig()

	redis.InitRedis(
		ctx,
		viper.GetString("redis.host"),
		viper.GetString("redis.port"),
		viper.GetString("redis.passwd"),
	)

	rabbitUrl := viper.GetString("rabbit.protocol") + "://" + viper.GetString("rabbit.username") + ":" + viper.GetString("rabbit.password") + "@" + viper.GetString("rabbit.ip") + ":" + viper.GetString("rabbit.port")

	mc, e := mq.InitMsgQueue(ctx, rabbitUrl)
	if e != nil {
		panic(e)
	}

	addr := viper.GetString("recv_grpc.ip") + ":" + viper.GetString("recv_grpc.port")
	rpc.NewGrpcRecvServer(cancle, addr, NewRpcServer(mc))
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
