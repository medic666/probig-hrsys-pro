package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code: 0,
		Msg:  "ok",
		Data: data,
	})
}

func SuccessWithMsg(c *gin.Context, msg string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code: 0,
		Msg:  msg,
		Data: data,
	})
}

func Error(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, Response{
		Code: 500,
		Msg:  msg,
		Data: nil,
	})
}

func ErrorWithCode(c *gin.Context, code int, msg string) {
	c.JSON(http.StatusOK, Response{
		Code: code,
		Msg:  msg,
		Data: nil,
	})
}

func BadRequest(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, Response{
		Code: 400,
		Msg:  msg,
		Data: nil,
	})
}

func Unauthorized(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, Response{
		Code: 401,
		Msg:  msg,
		Data: nil,
	})
}

func Forbidden(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, Response{
		Code: 403,
		Msg:  msg,
		Data: nil,
	})
}
