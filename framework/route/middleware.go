package route

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/hawthorntrees/cronframework/framework/dto/resp"
	"github.com/hawthorntrees/cronframework/framework/logger"
	"github.com/hawthorntrees/cronframework/framework/utils"
	"runtime/debug"
	"strings"
)

func JwtAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, ok := whitelist[c.Request.URL.Path]
		if ok {
			c.Next()
			return
		}
		tokenHeader := c.GetHeader("Authorization")
		if tokenHeader == "" || !strings.HasPrefix(tokenHeader, "Bearer ") {
			resp.Error(c, "缺少授权令牌")
			c.Abort()
			return
		}
		tokenString := tokenHeader[7:] // 去掉 "Bearer "

		claims, err := utils.ParseToken(tokenString)
		if err != nil {
			resp.Error(c, "令牌无效或已过期")
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Next()
	}
}

func traceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := utils.GenerateTraceID()
		if err != nil {
			id = "error"
		}
		c.Set("traceID", id)
		c.Next()
	}
}
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				stack := debug.Stack()
				log := logger.GetLogger(c)
				log.Sugar().Errorf(
					"[PANIC] 请求路径：%s | 方法：%s | 错误：%v | 堆栈：%s",
					c.Request.URL.Path,
					c.Request.Method,
					err,
					string(stack),
				)
				resp.Error(c, fmt.Sprintf("%v", err))
				c.Abort()
			}
		}()
		c.Next()
	}
}
