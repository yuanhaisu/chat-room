package send_server

import (
	"chat_room/mq"
	"chat_room/redis"
	"chat_room/rpc"
	"context"
	"flag"
	"fmt"
	"github.com/spf13/viper"
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

	rabbitUrl := viper.GetString("rabbit.protocol") + "://" +
		viper.GetString("rabbit.username") + ":" +
		viper.GetString("rabbit.password") + "@" +
		viper.GetString("rabbit.ip") + ":" +
		viper.GetString("rabbit.port")
	rm, e := mq.NewMsgQueue(rabbitUrl)
	if e != nil {
		panic(e)
	}
	mc := NewMsgCenter(rm)

	go mc.Delivery(ctx)

	addr := viper.GetString("send_grpc.ip") + ":" + viper.GetString("send_grpc.port")
	rpc.InitGrpcSendServer(addr,NewRpcServer(mc, redis.RedisSource))
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
