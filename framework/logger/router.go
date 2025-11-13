package logger

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func GetLogger(ctx *gin.Context) *zap.Logger {
	id, ok := ctx.Get("traceID")
	if ok {
		traceID, o := id.(string)
		if o {
			return GetBaseLogger().With(zap.String("traceID", traceID))
		}
	}
	return GetBaseLogger()
}
