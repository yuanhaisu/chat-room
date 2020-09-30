package main

import (
	"chat_room/glog"
	"context"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"gorm.io/gorm/utils"
	"time"
)

type Model struct {
	ID        int64      `json:"id" mapstructure:"ID" gorm:"primary_key;"`
	CreatedAt time.Time  `json:"-"`
	UpdatedAt time.Time  `json:"-"`
	DeletedAt *time.Time `json:"-" sql:"index"`
	Revision  int        `json:"-"`
}

type Model64 struct {
	ID        uint64     `json:"id" mapstructure:"ID" gorm:"primary_key;"`
	CreatedAt time.Time  `json:"-"`
	UpdatedAt time.Time  `json:"-"`
	DeletedAt *time.Time `json:"-" sql:"index"`
	Revision  uint       `json:"-"`
}

type HModel struct {
	ID        int64     `json:"id" gorm:"primary_key;"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
	Revision  int       `json:"-"`
}

//初始化数据库连接

var MysqlDB *gorm.DB

func InitMysql(name, passwd, host, port, datebase, prefix string) {
	var sql = name + ":" + passwd + "@(" + host + ":" + port + ")/" + datebase + "?charset=utf8mb4&parseTime=True&loc=Local"
	var err error
	MysqlDB, err = gorm.Open(mysql.Open(sql), &gorm.Config{
		Logger:                                   NewGormLogger(logger.Config{
			SlowThreshold: time.Second * 5,
			LogLevel:      logger.Info,
			Colorful:      false,
		}),
		DisableForeignKeyConstraintWhenMigrating: true,
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   prefix,
			SingularTable: true,
		},
	})
	if err != nil {
		panic(err)
	}
	sdb, e := MysqlDB.DB()
	if e != nil {
		panic(e)
	}
	// 启用gger，显示详细日志
	sdb.SetMaxIdleConns(10)
	sdb.SetMaxOpenConns(1024)
	sdb.SetConnMaxLifetime(time.Second * 600)
}

//获取MySql连接
func GetDB() *gorm.DB {
	return MysqlDB
}

func NewGormLogger(lc logger.Config) logger.Interface {
	return &gormLogger{
		lc,
	}
}

type gormLogger struct {
	logger.Config
}

func (gl *gormLogger) LogMode(level logger.LogLevel) logger.Interface {
	newlogger := *gl
	newlogger.LogLevel = level
	return &newlogger
}

func (gl *gormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	glog.InfolnWithTrace(ctx, msg, append([]interface{}{utils.FileWithLineNum()}, data...))
}

func (gl *gormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	glog.WarninglnWithTrace(ctx, msg, append([]interface{}{utils.FileWithLineNum()}, data...))
}

func (gl *gormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	glog.ErrorWithTrace(ctx, msg, append([]interface{}{utils.FileWithLineNum()}, data...))
}

func (gl *gormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if gl.LogLevel > 0 {
		elapsed := time.Since(begin)
		switch {
		case err != nil && gl.LogLevel >= logger.Error:
			sql, rows := fc()
			glog.ErrorWithTrace(ctx, utils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		case elapsed > gl.SlowThreshold && gl.SlowThreshold != 0 && gl.LogLevel >= logger.Warn:
			sql, rows := fc()
			glog.WarninglnWithTrace(ctx, "SlowThreshold:"+string(gl.SlowThreshold), utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, rows, sql)
		case gl.LogLevel >= logger.Info:
			sql, rows := fc()
			glog.InfolnWithTrace(ctx, utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	}
}

