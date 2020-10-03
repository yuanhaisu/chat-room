package mq

import (
	"fmt"
	"github.com/spf13/viper"
	"testing"
)

var Msq MsgSendQueue

func init(){
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("../")
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	rabbitUrl := viper.GetString("rabbit.protocol") + "://" + viper.GetString("rabbit.username") + ":" + viper.GetString("rabbit.password") + "@" + viper.GetString("rabbit.ip") + ":" + viper.GetString("rabbit.port")

	Msq, err = NewMsgQueue(rabbitUrl)
	if err!=nil{
		panic(err)
	}
}


func TestRabbitMq_Publish(t *testing.T) {
	if e:=Msq.Publish([]byte("ss"));e!=nil{
		t.Errorf("got [%v] expect [%v]",e,nil)
	}
}


func TestRabbitMq_QueuedMsg(t *testing.T) {
	rabbitUrl := viper.GetString("rabbit.protocol") + "://" + viper.GetString("rabbit.username") + ":" + viper.GetString("rabbit.password") + "@" + viper.GetString("rabbit.ip") + ":" + viper.GetString("rabbit.port")
	rm := &rabbitMq{Addr: rabbitUrl}
	ch:=rm.QueuedMsg()
	for v:=range ch{
		fmt.Println(v.Body)
	}
}

