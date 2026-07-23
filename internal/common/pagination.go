package common

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

type Pagination struct {
	Page     int    `form:"page"`
	PageSize int    `form:"pageSize"`
	Search   string `form:"search"`
	SortBy   string `form:"sortBy"`
	SortDir  string `form:"sortDir"`
}

func GetPagination(c *gin.Context) Pagination {
	p := Pagination{
		Page:     1,
		PageSize: 20,
		SortDir:  "DESC",
	}

	if v, err := strconv.Atoi(c.Query("page")); err == nil && v > 0 {
		p.Page = v
	}
	if v, err := strconv.Atoi(c.Query("pageSize")); err == nil && v > 0 {
		p.PageSize = v
		if p.PageSize > 200 {
			p.PageSize = 200
		}
	}
	p.Search = c.Query("search")
	p.SortBy = c.Query("sortBy")
	if dir := c.Query("sortDir"); dir == "ASC" || dir == "DESC" {
		p.SortDir = dir
	}

	return p
}

func (p Pagination) Offset() int {
	return (p.Page - 1) * p.PageSize
}

func (p Pagination) Limit() int {
	return p.PageSize
}
