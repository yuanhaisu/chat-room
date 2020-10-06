package recv_server

import (
	"chat-room/glog"
	"chat-room/mq"
	"chat-room/proto"
	"encoding/json"
	"io"
)

func NewRpcServer(mq mq.MsgSendQueue) proto.RecvServer {
	return &Server{
		Mq: mq,
	}
}

type Server struct {
	Mq mq.MsgSendQueue
}

//接收用户消息
func (s *Server) UserRecvStream(urStream proto.Recv_UserRecvStreamServer) error {
	for {
		select {
		case <-urStream.Context().Done():
			return nil
		default:
			data, e := urStream.Recv()
			if e == io.EOF {
				return nil
			}
			if e != nil {
				glog.Error("stream recv error occurred :", e)
				return e
			}
			glog.Infof("收到用户消息：%+v", data)
			b, _ := json.Marshal(data)
			if e := s.Mq.Publish(b); e != nil {
				glog.Error("发布消息失败，失败信息：")
			}
		}
	}
}
