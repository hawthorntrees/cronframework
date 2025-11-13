package route

import (
	"github.com/gin-gonic/gin"
	"github.com/hawthorntrees/cronframework/framework/config"
	"github.com/hawthorntrees/cronframework/framework/controller"
)

var whitelist = make(map[string]struct{})

func Init(cfg *config.ServerConfig) (*gin.Engine, *gin.RouterGroup) {
	gin.SetMode(gin.ReleaseMode)
	initWhitelist()
	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(traceMiddleware())
	engine.Use(JwtAuth())
	routerGroup := engine.Group(cfg.BashPath)
	controller.RegisterRouter(routerGroup)
	return engine, routerGroup
}
func initWhitelist() {
	basePath := config.GetBasePath()
	if basePath == "/" {
		basePath = ""
	}
	login := basePath + "/sys/login"
	whitelist[login] = struct{}{}
}
