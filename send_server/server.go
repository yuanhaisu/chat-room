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

var (
	confFilePath   = flag.String("confFilePath", "./", "define config file info")
	confFileName   = flag.String("confFileName", "config", "define config file info")
	confFileFormat = flag.String("confFileFormat", "yaml", "define config file info")
)

func Execute() {

	ctx, cancel := context.WithCancel(context.Background())

	InitConfig(*confFileName, *confFilePath, *confFileFormat)

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

func InitConfig(fileName, filePath, format string) {
	viper.SetConfigName(fileName)
	viper.SetConfigType(format)
	viper.AddConfigPath(filePath)
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
}
