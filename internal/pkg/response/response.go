package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	CodeSuccess      = 0
	CodeError        = 1
	CodeUnauthorized = 401
	CodeForbidden    = 403
	CodeNotFound     = 404
	CodeParamError   = 400
	CodeServerError  = 500
)

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

type PageData struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"pageSize"`
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
		Code: CodeError,
		Msg:  msg,
		Data: nil,
	})
}

func Page(c *gin.Context, list interface{}, total int64, page, pageSize int) {
	c.JSON(http.StatusOK, Response{
		Code: CodeSuccess,
		Msg:  "success",
		Data: PageData{
			List:     list,
			Total:    total,
			Page:     page,
			PageSize: pageSize,
		},
	})
}

func ParamError(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, Response{
		Code: CodeParamError,
		Msg:  msg,
		Data: nil,
	})
}

func Unauthorized(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, Response{
		Code: CodeUnauthorized,
		Msg:  msg,
		Data: nil,
	})
}

func Forbidden(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, Response{
		Code: CodeForbidden,
		Msg:  msg,
		Data: nil,
	})
}

func NotFound(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, Response{
		Code: CodeNotFound,
		Msg:  msg,
		Data: nil,
	})
}

func ServerError(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, Response{
		Code: CodeServerError,
		Msg:  msg,
		Data: nil,
	})
}
