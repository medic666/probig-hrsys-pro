package company

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"probig/internal/pkg/audit"
	"probig/internal/pkg/excel"
	"probig/internal/pkg/middleware"
	"probig/internal/pkg/response"
	"probig/internal/pkg/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup) {
	r.Use(middleware.Auth)
	r.GET("", middleware.Permission("company:read"), List)
	r.GET("/:id", middleware.Permission("company:read"), GetByID)
	r.POST("", middleware.Permission("company:write"), Create)
	r.PUT("/:id", middleware.Permission("company:write"), Update)
	r.DELETE("/:id", middleware.Permission("company:delete"), Delete)
	r.POST("/:id/restore", middleware.Permission("company:delete"), Restore)
	r.GET("/deleted", middleware.Permission("company:read"), GetDeletedList)
	r.GET("/export", middleware.Permission("company:export"), Export)
}

func List(c *gin.Context) {
	pageNum, _ := strconv.Atoi(c.DefaultQuery("pageNum", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	svc := NewService()
	list, total, err := svc.List(pageNum, pageSize, c.Query("name"), c.Query("credit_code"))
	if err != nil {
		response.Error(c, "查询失败")
		return
	}
	response.PageResult(c, list, total)
}

func GetByID(c *gin.Context) {
	id := utils.ParseUint(c.Param("id"))
	svc := NewService()
	data, err := svc.GetByID(id)
	if err != nil {
		response.Error(c, "查询失败")
		return
	}
	response.Success(c, data)
}

func Create(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "参数错误")
		return
	}
	svc := NewService()
	id, err := svc.Create(req)
	if err != nil {
		response.Error(c, "创建失败，统一社会信用代码可能已存在")
		return
	}
	userID, _ := c.Get("user_id")
	uname, _ := c.Get("username")
	name, _ := req["name"].(string)
	audit.Write(c, userID.(uint), uname.(string), "公司", id, name, "新增", nil, req)
	response.Success(c, gin.H{"id": id})
}

func Update(c *gin.Context) {
	id := utils.ParseUint(c.Param("id"))
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "参数错误")
		return
	}
	svc := NewService()
	if err := svc.Update(id, req); err != nil {
		response.Error(c, "更新失败")
		return
	}
	userID, _ := c.Get("user_id")
	uname, _ := c.Get("username")
	audit.Write(c, userID.(uint), uname.(string), "公司", id, "", "修改", nil, req)
	response.SuccessMsg(c, "更新成功")
}

func Delete(c *gin.Context) {
	id := utils.ParseUint(c.Param("id"))
	svc := NewService()
	if err := svc.Delete(id); err != nil {
		response.Error(c, "删除失败")
		return
	}
	userID, _ := c.Get("user_id")
	uname, _ := c.Get("username")
	audit.Write(c, userID.(uint), uname.(string), "公司", id, "", "删除", nil, nil)
	response.SuccessMsg(c, "删除成功")
}

func Restore(c *gin.Context) {
	id := utils.ParseUint(c.Param("id"))
	svc := NewService()
	if err := svc.Restore(id); err != nil {
		response.Error(c, "恢复失败")
		return
	}
	response.SuccessMsg(c, "恢复成功")
}

func GetDeletedList(c *gin.Context) {
	pageNum, _ := strconv.Atoi(c.DefaultQuery("pageNum", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	svc := NewService()
	list, total, err := svc.GetDeletedList(pageNum, pageSize)
	if err != nil {
		response.Error(c, "查询失败")
		return
	}
	response.PageResult(c, list, total)
}

func Export(c *gin.Context) {
	svc := NewService()
	list, _, err := svc.List(1, 100000, c.Query("name"), c.Query("credit_code"))
	if err != nil {
		response.Error(c, "导出失败")
		return
	}
	exp := excel.NewExporter([]string{"公司名称", "统一社会信用代码", "地址", "联系电话", "开户行", "银行账号"})
	for _, item := range list {
		exp.AddRow([]string{
			fmt.Sprint(item["name"]),
			fmt.Sprint(item["credit_code"]),
			fmt.Sprint(item["address"]),
			fmt.Sprint(item["contact_phone"]),
			fmt.Sprint(item["bank_name"]),
			fmt.Sprint(item["bank_account"]),
		})
	}
	filePath := filepath.Join(os.TempDir(), fmt.Sprintf("company_export_%d.csv", time.Now().Unix()))
	if err := exp.WriteCSV(filePath); err != nil {
		response.Error(c, "导出失败")
		return
	}
	c.Header("Content-Disposition", "attachment; filename=company_export.csv")
	c.Header("Content-Type", "text/csv")
	c.File(filePath)
}
