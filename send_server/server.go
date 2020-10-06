package send_server

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
	ctx, cancel := context.WithCancel(context.Background())

	initConfig()

	redis.InitRedis(
		ctx,
		viper.GetString("redis.host"),
		viper.GetString("redis.port"),
		viper.GetString("redis.passwd"),
	)

	rabbitUrl := viper.GetString("rabbit.protocol") + "://" +
		viper.GetString("rabbit.username") + ":" +
		viper.GetString("rabbit.password") + "@" +
		viper.GetString("rabbit.ip") + ":" +
		viper.GetString("rabbit.port")
	rm, e := mq.InitMsgQueue(ctx, rabbitUrl)
	if e != nil {
		panic(e)
	}
	defer rm.Close()
	mc := NewMsgCenter(rm)

	go mc.Delivery(ctx)

	addr := viper.GetString("send_grpc.ip") + ":" + viper.GetString("send_grpc.port")
	rpc.InitGrpcSendServer(cancel, addr, NewRpcServer(mc, redis.RedisSource))
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
