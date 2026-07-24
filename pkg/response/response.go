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
	PageSize int         `json:"page_size"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{Code: CodeSuccess, Msg: "success", Data: data})
}

func PageSuccess(c *gin.Context, data PageData) {
	c.JSON(http.StatusOK, Response{Code: CodeSuccess, Msg: "success", Data: data})
}

func Error(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, Response{Code: CodeError, Msg: msg, Data: nil})
	c.Abort()
}

func Unauthorized(c *gin.Context, msg string) {
	c.JSON(http.StatusUnauthorized, Response{Code: CodeUnauthorized, Msg: msg, Data: nil})
	c.Abort()
}

func Forbidden(c *gin.Context) {
	c.JSON(http.StatusForbidden, Response{Code: CodeForbidden, Msg: "权限不足", Data: nil})
	c.Abort()
}
