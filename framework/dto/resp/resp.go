package resp

import (
	"github.com/gin-gonic/gin"
	"strings"
)

type RespData struct {
	Code    string      `json:"code"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
	TraceId string      `json:"traceId"`
}
type RespJson map[string]interface{}

func Error(ginContext *gin.Context, message ...string) {
	msg := "交易失败"
	if len(message) > 0 {
		msg = strings.Join(message, ",")
	}
	respData := RespData{
		Code:    "100000",
		Data:    nil,
		Message: msg,
	}
	ginContext.JSON(200, respData)
}

func Success(ginContext *gin.Context, data interface{}, message ...string) {
	msg := "交易成功"
	if len(message) > 0 {
		msg = strings.Join(message, ",")
	}
	respData := RespData{
		Code:    "000000",
		Data:    data,
		Message: msg,
	}
	ginContext.JSON(200, respData)
}

type PageResult struct {
	Data  interface{} `json:"data"`
	Total int64       `json:"total"`
}
