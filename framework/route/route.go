package route

import (
	"github.com/gin-gonic/gin"
	"github.com/hawthorntrees/cronframework/framework/controller"
	"github.com/hawthorntrees/cronframework/framework/logger"
	"go.uber.org/zap"
)

func Init() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(traceMiddleware())
	router.Use(JwtAuth())
	group := router.Group("/api")
	controller.RegisterRouter(group)
	return router
}

func GetLogger(ctx *gin.Context) *zap.Logger {
	id, ok := ctx.Get("id")
	if ok {
		traceID, o := id.(string)
		if o {
			return logger.GetBaseLogger().With(zap.String("traceID", traceID))
		}
	}
	return logger.GetBaseLogger()
}
