package redis

import (
	"chat-room/common"
	"chat-room/glog"
	"context"
	"github.com/garyburd/redigo/redis"
	"time"
)

//初始化一个pool
func newPool(server, password string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     30,
		MaxActive:   1000,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", server)

			if err != nil {
				return nil, err
			}
			if _, err := c.Do("AUTH", password); err != nil {
				c.Close()
				return nil, err
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}
}

type Redis struct {
	ctx  context.Context
	pool *redis.Pool
}

var RedisSource Redis

func InitRedis(ctx context.Context, host, port, passwd string) {
	RedisSource.pool = newPool(host+":"+port, passwd)
	glog.Infoln("连接池初始化成功")
}

func (r *Redis) Do(commandName string, args ...interface{}) (reply interface{}, err error) {
	glog.Infoln("redis command:", commandName, common.InterfaceSlice2String(args, " "))
	conn := r.pool.Get()
	defer func() {
		if e := conn.Close(); e != nil {
			glog.Error(e)
		}
	}()
	return conn.Do(commandName, args...)
}

func (r *Redis) Close() {
	for {
		select {
		case <-r.ctx.Done():
			if e := r.pool.Close(); e != nil {
				glog.Error("redis - Pool Close Failed: ", e)
			}
		}
	}

}
