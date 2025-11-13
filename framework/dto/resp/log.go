package resp

import "github.com/gin-gonic/gin"

func GetTraceIDFromContext(ctx *gin.Context) string {
	id, ok := ctx.Get("traceID")
	if ok {
		return id.(string)
	}
	return ""
}
