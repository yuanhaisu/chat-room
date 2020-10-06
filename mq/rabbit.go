package mq

import (
	"chat_room/glog"
	"context"
	"github.com/spf13/viper"
	"github.com/streadway/amqp"
	"sync"
)

type MsgSendQueue interface {
	Publish(data []byte) error
	QueuedMsg() <-chan amqp.Delivery
	Close()
	Reconnect()
}

type rabbitMq struct {
	conn          *amqp.Connection
	connNotify    chan *amqp.Error
	channelNotify chan *amqp.Error
	channel       *amqp.Channel
	queue         amqp.Queue
	consumer      <-chan amqp.Delivery
	Addr          string
	ctx           context.Context
	sync.Mutex
}

func InitMsgQueue(ctx context.Context, addr string) (msq MsgSendQueue, err error) {

	rm := &rabbitMq{ctx: ctx, Addr: addr}

	rm.Addr = addr
	if err = rm.connect(); err != nil {
		return nil, err
	}

	return rm, err
}

func (rm *rabbitMq) connect() (err error) {
	rm.conn, err = amqp.Dial(rm.Addr)
	if err != nil {
		glog.Error("Failed to connect to RabbitMQ，", err)
		return err
	}
	rm.connNotify = rm.conn.NotifyClose(make(chan *amqp.Error))
	if err = rm.setChannel(); err != nil {
		return err
	}

	if err = rm.setQueue(); err != nil {
		return err
	}
	return
}

func (rm *rabbitMq) Reconnect() {
	for {
		select {
		case err := <-rm.connNotify:
			if err != nil {
				glog.Error("rabbitmq consumer - connection NotifyClose: ", err)
			}
		case err := <-rm.channelNotify:
			if err != nil {
				glog.Error("rabbitmq consumer - channel NotifyClose: ", err)
			}
		case <-rm.ctx.Done():
			return
		}
		if e := rm.connect(); e != nil {
			panic(e)
		}
	}
}

func (rm *rabbitMq) Close() {
	if e := rm.conn.Close(); e != nil {
		glog.Error("rabbitmq consumer - Close Failed: ", e)
	}
	for err := range rm.channelNotify {
		println(err)
	}
	for err := range rm.connNotify {
		println(err)
	}
}

func (rm *rabbitMq) setChannel() (e error) {
	if !rm.conn.IsClosed() {
		rm.channel, e = rm.conn.Channel()
		if e != nil {
			glog.Error("Failed to open a channel，", e)
			return e
		}
		rm.channelNotify = rm.channel.NotifyClose(make(chan *amqp.Error))
	}

	return
}

func (rm *rabbitMq) setQueue() (e error) {
	rm.queue, e = rm.channel.QueueDeclare(
		viper.GetString("rabbit.chat_msg_queue_name"), // name
		true,  // durable
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
	rm.Lock()
	defer rm.Unlock()
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

func (rm *rabbitMq) GetConn() *amqp.Connection {
	rm.Lock()
	defer rm.Unlock()
	if rm.conn.IsClosed() {
		if e := rm.connect(); e != nil {
			panic(e)
		}
	}
	return rm.conn
}

func (rm *rabbitMq) GetChannel() *amqp.Channel {

	if rm.channel == nil {
		if e := rm.setChannel(); e != nil {
			panic(e)
		}
	}
	return rm.channel
}

func (rm *rabbitMq) newConsume() {
	var err error

	for retry := 3; retry > 0; retry-- {
		rm.consumer, err = rm.GetChannel().Consume(
			rm.queue.Name, // queue
			"",            // consumer
			false,         // auto-ack
			false,         // exclusive
			false,         // no-local
			false,         // no-wait
			nil,           // args
		)
		if err != nil {
			if err == amqp.ErrClosed {
				continue
			}
			panic("mq consumer init fail:" + err.Error())
		}
		return
	}
	panic("mq consumer init fail:" + err.Error())
}

func (rm *rabbitMq) Publish(data []byte) error {
	if err := rm.GetChannel().Publish(
		"",            // exchange
		rm.queue.Name, // routing key
		false,         // mandatory
		false,         // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         data,
		}); err != nil {
		glog.Error("Failed to publish a message:", err)
		return err
	}
	glog.Infoln("消息发布成功")
	return nil
}
