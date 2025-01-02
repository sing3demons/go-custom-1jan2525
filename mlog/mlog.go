package mlog

import (
	"github.com/gin-gonic/gin"
	"github.com/sing3demons/http-kafka-ms/constant"
	"github.com/sing3demons/http-kafka-ms/logger"
)

func L(c gin.Context) logger.ILogger {
	value, _ := c.Get(constant.Key)
	switch log := value.(type) {
	case logger.ILogger:
		return log
	default:
		return logger.NewLogger()
	}
}
