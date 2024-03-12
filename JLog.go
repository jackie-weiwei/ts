package ts

import (
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

var log = logrus.New()

func InitLog(logPath string) {
	log.SetOutput(&lumberjack.Logger{
		Filename:   logPath, // 日志文件路径
		MaxSize:    10,      // 文件最大大小(MB)
		MaxBackups: 3,       // 保留旧文件的最大个数
		MaxAge:     28,      // 保留旧文件的最大天数
		Compress:   true,    // 是否压缩/归档旧文件
	})
}

func Info(args ...interface{}) {
	log.Info(args...)
}

func Error(args ...interface{}) {
	log.Error(args...)
}

func Debug(args ...interface{}) {
	log.Debug(args...)
}

func Print(args ...interface{}) {
	log.Print(args...)
}

func Trace(args ...interface{}) {
	log.Trace(args...)
}

func Warn(args ...interface{}) {
	log.Warn(args...)
}
