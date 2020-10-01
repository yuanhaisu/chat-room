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
	viper.AddConfigPath("./")
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

func TestNewMsgQueue(t *testing.T) {

}

func TestRabbitMq_Publish(t *testing.T) {
	if e:=Msq.Publish([]byte("ss"));e!=nil{
		t.Errorf("got [%s] expect [%s]",e,nil)
	}
}

func BenchmarkRabbitMq_Publish(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if e:=Msq.Publish([]byte("benchmar"));e!=nil{
			b.Fatal("publish failed！ reason is ",e)
		}
	}
}

func TestRabbitMq_QueuedMsg(t *testing.T) {
	
}

