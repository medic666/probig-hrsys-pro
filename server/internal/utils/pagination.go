package utils

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

type PageRequest struct {
	PageNum  int `json:"page_num" form:"pageNum"`
	PageSize int `json:"page_size" form:"pageSize"`
}

type PageResult struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	PageNum  int         `json:"page_num"`
	PageSize int         `json:"page_size"`
}

const (
	DefaultPageNum  = 1
	DefaultPageSize = 20
	MaxPageSize     = 100
)

func BindPage(c *gin.Context) PageRequest {
	pageNum, _ := strconv.Atoi(c.DefaultQuery("pageNum", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	if pageNum < 1 {
		pageNum = DefaultPageNum
	}
	if pageSize < 1 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}

	return PageRequest{
		PageNum:  pageNum,
		PageSize: pageSize,
	}
}

func NewPageResult(list interface{}, total int64, req PageRequest) *PageResult {
	return &PageResult{
		List:     list,
		Total:    total,
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
	}
}

func (p *PageRequest) Offset() int {
	return (p.PageNum - 1) * p.PageSize
}
