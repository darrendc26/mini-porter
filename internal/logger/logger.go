package logger

import "go.uber.org/zap"

var Logger *zap.SugaredLogger

func Init() {
	l, _ := zap.NewProduction()
	Logger = l.Sugar()
}
