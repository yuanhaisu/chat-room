package send_server

import (
	"chat_room/common"
	"chat_room/glog"
	"chat_room/mq"
	"chat_room/proto"
	"chat_room/redis"
	"context"
	"encoding/json"
	"sync"
	"time"
)

type MsgCenter interface {
	Subscribe(name string) (ch <-chan *proto.Request, err error)
	Delivery(ctx context.Context)
	NotifyOffline(*proto.Request)
}

type msgCenter struct {
	onlineMsgChan *sync.Map
	offlineNotify chan string
	Mq            mq.MsgSendQueue
}

func NewMsgCenter(mq mq.MsgSendQueue) MsgCenter {
	return &msgCenter{
		offlineNotify: make(chan string),
		Mq:            mq,
		onlineMsgChan: new(sync.Map),
	}
}

func (mc *msgCenter) Subscribe(name string) (ch <-chan *proto.Request, err error) {
	res, isLoad := mc.onlineMsgChan.LoadOrStore(name, make(chan *proto.Request))
	if isLoad {
		return nil, common.ErrNameIsExisted
	}
	req := &proto.Request{
		From:    name,
		Content: name + "上线了",
		To:      "",
		Time:    time.Now().Format("2006-01-02 15:04:05"),
		Action:  common.Join,
	}
	b, _ := json.Marshal(req)
	_ = mc.Mq.Publish(b)
	return res.(chan *proto.Request), nil
}

func (mc *msgCenter) Delivery(ctx context.Context) {
	msgs := mc.Mq.QueuedMsg()
	for {
		select {
		case <-ctx.Done():
			glog.Infoln("consumer is closing...")
			return
		case msg, isClose := <-msgs:
			if !isClose {
				msgs = mc.Mq.QueuedMsg()
				glog.Error("重连消费者")
				return
			}
			glog.Infoln("收到消息队列内容：", string(msg.Body))
			req := &proto.Request{}

			if e := json.Unmarshal(msg.Body, req); e != nil {
				glog.Error("unknow content:", e, string(msg.Body))
				continue
			}
			if req.Action == common.Aite {
				resFrom, ok := mc.onlineMsgChan.Load(req.From)
				if ok {
					resFrom.(chan *proto.Request) <- req
				}
				resTo, ok := mc.onlineMsgChan.Load(req.To)
				if ok {
					resTo.(chan *proto.Request) <- req
					continue
				}
				_, e := redis.RedisSource.Do("LPUSH", req.To, msg.Body)
				if e != nil {
					glog.Error("持久化离线消息失败，失败信息：", e, "，消息内容：", msg.Body)
				}
				continue
			}
			//每个在线用户群发
			mc.onlineMsgChan.Range(func(key, value interface{}) bool {
				c, _ := mc.onlineMsgChan.Load(key)
				c.(chan *proto.Request) <- req
				glog.Infoln("写入消息到发送通道")
				return true
			})
			//用户下线，收到关闭指定通道的通知
		case name := <-mc.offlineNotify:
			res, ok := mc.onlineMsgChan.Load(name)
			if ok {
				close(res.(chan *proto.Request))
				mc.onlineMsgChan.Delete(name)
			}
		}
	}
}

func (mc *msgCenter) NotifyOffline(req *proto.Request) {
	mc.offlineNotify <- req.From
	b, _ := json.Marshal(req)
	mc.Mq.Publish(b)
}
