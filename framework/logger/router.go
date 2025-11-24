package logger

import (
	"context"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func GetLogger(ctx *gin.Context) *zap.Logger {
	id, ok := ctx.Get("traceID")
	if ok {
		traceID, o := id.(string)
		if o {
			return GetDefaultLogger().With(zap.String("traceID", traceID))
		}
	}
	return GetDefaultLogger().With(zap.String("traceID", ""))
}
func GetContextFromGin(ctx *gin.Context) context.Context {
	id, ok := ctx.Get("traceID")
	if ok {
		traceID, o := id.(string)
		if o {
			return context.WithValue(context.Background(), "traceID", traceID)
		}
	}
	return context.WithValue(context.Background(), "traceID", "notGetTraceID")
}
