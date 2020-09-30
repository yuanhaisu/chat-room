package send_server

import (
	"chat_room/common"
	"chat_room/glog"
	"chat_room/proto"
	"chat_room/redis"
	"encoding/json"
	"fmt"
	"io"
	"time"
)

func NewRpcServer(mc MsgCenter, redis redis.Redis) proto.SendServer {
	return &Server{
		Mc:    mc,
		redis: redis,
	}
}

type Server struct {
	Mc    MsgCenter
	redis redis.Redis
}

//发消息给用户
func (s *Server) UserSendStream(req *proto.Request, usStream proto.Send_UserSendStreamServer) error {
	var (
		ch <-chan *proto.Request
		e  error
	)
	if req.Action == common.Join {
		//验证登录
		ch, e = s.Mc.Subscribe(req.From)
		if e != nil {
			return e
		}
	}

	s.SendUnreadMsg(usStream, req.From)
	for {
		select {
		case req := <-ch:
			glog.Infof("从发送通道读取到消息：%+v", req)
			s.doSend(usStream, req)
		case <-usStream.Context().Done():
			fmt.Println("收到" + req.From + "正常下线")
			req.Action = common.Quit
			req.Time = time.Now().Format("2006-01-02 15:04:05")
			req.Content = req.From + "正常下线"
			s.Mc.NotifyOffline(req)
			return nil
		}
	}
}

func (s *Server) SendUnreadMsg(usStream proto.Send_UserSendStreamServer, name string) {
	for {
		raw, e := s.redis.Do("rpop", name)
		if e != nil {
			glog.Errorf("获取未读消息失败，失败信息：%v", e, "，当前用户信息：%+v", name)
			return
		}
		if raw == nil {
			return
		}
		var r proto.Request
		bRaw, ok := raw.([]byte)
		if ok {
			if e := json.Unmarshal(bRaw, &r); e != nil {
				glog.Error("未读消息内容异常，错误信息:", e, "当前用户：", name)
				return
			}
		}
		s.doSend(usStream, &r)
	}
}

func (s *Server) doSend(usStream proto.Send_UserSendStreamServer, req *proto.Request) {
	if e := usStream.Send(req); e != nil {
		if req.Action == common.Aite {
			b, _ := json.Marshal(req)
			if _, e := s.redis.Do("LPUSH", req.From, b); e != nil {
				glog.Error("存储离线消息失败，失败信息:", e)
				return
			}
		}
		if e == io.EOF {
			glog.Error("补货用户")
			return
		}
		glog.Errorf("发送消息给用户失败，失败信息：%v", e, "，消息内容：%+v", req)
		return
	}
	glog.Infoln("完成消息发送给用户")
}
