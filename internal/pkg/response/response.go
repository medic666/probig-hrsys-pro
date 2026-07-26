package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	CodeSuccess   = 0
	ParamError    = 1001
	Unauthorized  = 1002
	Forbidden     = 1003
	NotFound      = 1004
	InternalError = 1005
	BusinessError = 1006
)

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

type PageResult struct {
	List  interface{} `json:"list"`
	Total int64       `json:"total"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code: CodeSuccess,
		Msg:  "success",
		Data: data,
	})
}

func SuccessWithMsg(c *gin.Context, msg string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code: CodeSuccess,
		Msg:  msg,
		Data: data,
	})
}

func Error(c *gin.Context, code int, msg string) {
	c.JSON(http.StatusOK, Response{
		Code: code,
		Msg:  msg,
		Data: nil,
	})
}

func ErrorWithMsg(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, Response{
		Code: BusinessError,
		Msg:  msg,
		Data: nil,
	})
}
