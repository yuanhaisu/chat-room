package mq

import (
	"chat_room/glog"
	"github.com/spf13/viper"
	"github.com/streadway/amqp"
)

type MsgSendQueue interface {
	Publish(data []byte) error
	QueuedMsg() <-chan amqp.Delivery
}

type rabbitMq struct {
	conn     *amqp.Connection
	channel  *amqp.Channel
	queue    amqp.Queue
	consumer <-chan amqp.Delivery
	Addr     string
}

func NewMsgQueue(addr string) (MsgSendQueue, error) {

	rm := &rabbitMq{Addr: addr}

	rm.Addr = addr
	e := rm.connect()

	return rm, e
}

func (rm *rabbitMq) connect() (e error) {
	rm.conn, e = amqp.Dial(rm.Addr)
	if e != nil {
		glog.Error("Failed to connect to RabbitMQ，", e)
		return e
	}
	rm.channel, e = rm.conn.Channel()
	if e != nil {
		glog.Error("Failed to open a channel，", e)
		return e
	}
	rm.queue, e = rm.channel.QueueDeclare(
		viper.GetString("rabbit.chat_msg_queue_name"), // name
		false, // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if e != nil {
		glog.Error("Failed to declare a queue，", e)
		return e
	}
	return
}

func (rm *rabbitMq) QueuedMsg() <-chan amqp.Delivery {

	if rm.consumer == nil {
		rm.newConsume()
	} else {
		_, isClose := <-rm.consumer
		if !isClose {
			rm.newConsume()
		}
	}
	return rm.consumer
}

func (rm *rabbitMq) newConsume() {
	var err error
	rm.consumer, err = rm.channel.Consume(
		rm.queue.Name, // queue
		"",            // consumer
		true,          // auto-ack
		false,         // exclusive
		false,         // no-local
		false,         // no-wait
		nil,           // args
	)
	if err != nil {
		if err == amqp.ErrClosed {
			if e := rm.connect(); e != nil {
				panic(e)
			}
			return
		}
		panic("mq consumer init fail:" + err.Error())
	}
}

func (rm *rabbitMq) Publish(data []byte) error {
	if err := rm.channel.Publish(
		"",            // exchange
		rm.queue.Name, // routing key
		false,         // mandatory
		false,         // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        data,
		}); err != nil {
		glog.Error("Failed to publish a message:", err)
		return err
	}
	glog.Infoln("消息发布成功")
	return nil
}
