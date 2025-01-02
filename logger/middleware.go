package logger

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sing3demons/http-kafka-ms/constant"
	"go.uber.org/zap"
)

func LoggingMiddleware(logger ILogger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		reqId := ctx.Writer.Header().Get(constant.XSession)
		if reqId == "" {
			reqId = uuid.NewString()
			ctx.Writer.Header().Set(constant.XSession, reqId)
		}
		l := logParentID(ctx, logger)

		ctx.Set(constant.Key, l)
		ctx.Next()
		ctx.Next()
	}
}

func logParentID(c *gin.Context, logger ILogger) ILogger {
	xParent := c.Request.Header.Get("X-Parent-ID")
	if xParent == "" {
		xParent = uuid.NewString()
	}
	xSpan := uuid.NewString()

	reqId := c.Writer.Header().Get(constant.XSession)
	return logger.With(zap.String("parent-id", xParent),
		zap.String("span-id", xSpan), zap.String("session", reqId))
}
