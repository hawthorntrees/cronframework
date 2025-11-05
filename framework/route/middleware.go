package route

import (
	"github.com/gin-gonic/gin"
	"github.com/hawthorntrees/cronframework/framework/dto/resp"
	"github.com/hawthorntrees/cronframework/framework/utils"
	"strings"
)

func JwtAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/api/login" {
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
