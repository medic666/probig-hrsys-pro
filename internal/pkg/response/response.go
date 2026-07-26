package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

const (
	CodeSuccess = 0
	CodeError   = 1
	CodeAuthErr = 401
	CodeForbid  = 403
)

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{Code: CodeSuccess, Msg: "success", Data: data})
}

func SuccessMsg(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, Response{Code: CodeSuccess, Msg: msg, Data: nil})
}

func Error(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, Response{Code: CodeError, Msg: msg, Data: nil})
}

func ErrorWithCode(c *gin.Context, code int, msg string) {
	c.JSON(http.StatusOK, Response{Code: code, Msg: msg, Data: nil})
}

func PageResult(c *gin.Context, list interface{}, total int64) {
	c.JSON(http.StatusOK, Response{
		Code: CodeSuccess,
		Msg:  "success",
		Data: gin.H{"list": list, "total": total},
	})
}
