package person

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
	"probig/internal/position"
	"strconv"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup) {
	auth := r.Group("").Use(middleware.Auth, middleware.Permission("person:read"))
	{
		auth.GET("", List)
		auth.GET("/:id", GetByID)
	}

	write := r.Group("").Use(middleware.Auth, middleware.Permission("person:write"))
	{
		write.POST("", Create)
		write.PUT("/:id", Update)
		write.POST("/phone", AddPhone)
		write.PUT("/phone/:id", UpdatePhone)
		write.DELETE("/phone/:id", DeletePhone)
		write.POST("/email", AddEmail)
		write.PUT("/email/:id", UpdateEmail)
		write.DELETE("/email/:id", DeleteEmail)
		write.POST("/bank-card", AddBankCard)
		write.PUT("/bank-card/:id", UpdateBankCard)
		write.DELETE("/bank-card/:id", DeleteBankCard)
	}

	del := r.Group("").Use(middleware.Auth, middleware.Permission("person:delete"))
	{
		del.DELETE("/:id", Delete)
		del.POST("/:id/restore", Restore)
		del.GET("/deleted", GetDeletedList)
	}

	r.GET("/export", middleware.Auth, middleware.Permission("person:export"), Export)
}

func List(c *gin.Context) {
	pageNum, _ := strconv.Atoi(c.DefaultQuery("pageNum", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	svc := NewService()
	list, total, err := svc.List(pageNum, pageSize,
		c.Query("name"), c.Query("id_card"),
		c.Query("attendance_group"), c.Query("status"))
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
		response.Error(c, err.Error())
		return
	}

	entryDate := getString(req, "effective_date")
	if entryDate == "" {
		entryDate = time.Now().Format("2006-01-02")
	}
	posReq := map[string]interface{}{
		"person_id":      float64(id),
		"event_name":     "入职",
		"effective_date": entryDate,
	}
	posSvc := position.NewService()
	posSvc.CreateEvent(posReq)

	userID, _ := c.Get("user_id")
	uname, _ := c.Get("username")
	audit.Write(c, userID.(uint), uname.(string), "人员", id, getString(req, "name"), "新增", nil, req)
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
	audit.Write(c, userID.(uint), uname.(string), "人员", id, getString(req, "name"), "修改", nil, req)
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
	audit.Write(c, userID.(uint), uname.(string), "人员", id, "", "删除", nil, nil)
	response.SuccessMsg(c, "删除成功")
}

func Restore(c *gin.Context) {
	id := utils.ParseUint(c.Param("id"))
	svc := NewService()
	if err := svc.Restore(id); err != nil {
		response.Error(c, "恢复失败")
		return
	}
	userID, _ := c.Get("user_id")
	uname, _ := c.Get("username")
	audit.Write(c, userID.(uint), uname.(string), "人员", id, "", "恢复", nil, nil)
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

func AddPhone(c *gin.Context) {
	var req struct {
		PersonID uint   `json:"person_id" binding:"required"`
		Phone    string `json:"phone" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "参数错误")
		return
	}
	svc := NewService()
	if err := svc.AddPhone(req.PersonID, req.Phone); err != nil {
		response.Error(c, "添加失败")
		return
	}
	response.SuccessMsg(c, "添加成功")
}

func UpdatePhone(c *gin.Context) {
	id := utils.ParseUint(c.Param("id"))
	var req struct {
		Phone string `json:"phone" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "参数错误")
		return
	}
	svc := NewService()
	if err := svc.UpdatePhone(id, req.Phone); err != nil {
		response.Error(c, "更新失败")
		return
	}
	response.SuccessMsg(c, "更新成功")
}

func DeletePhone(c *gin.Context) {
	id := utils.ParseUint(c.Param("id"))
	svc := NewService()
	if err := svc.DeletePhone(id); err != nil {
		response.Error(c, "删除失败")
		return
	}
	response.SuccessMsg(c, "删除成功")
}

func AddEmail(c *gin.Context) {
	var req struct {
		PersonID uint   `json:"person_id" binding:"required"`
		Email    string `json:"email" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "参数错误")
		return
	}
	svc := NewService()
	if err := svc.AddEmail(req.PersonID, req.Email); err != nil {
		response.Error(c, "添加失败")
		return
	}
	response.SuccessMsg(c, "添加成功")
}

func UpdateEmail(c *gin.Context) {
	id := utils.ParseUint(c.Param("id"))
	var req struct {
		Email string `json:"email" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "参数错误")
		return
	}
	svc := NewService()
	if err := svc.UpdateEmail(id, req.Email); err != nil {
		response.Error(c, "更新失败")
		return
	}
	response.SuccessMsg(c, "更新成功")
}

func DeleteEmail(c *gin.Context) {
	id := utils.ParseUint(c.Param("id"))
	svc := NewService()
	if err := svc.DeleteEmail(id); err != nil {
		response.Error(c, "删除失败")
		return
	}
	response.SuccessMsg(c, "删除成功")
}

func AddBankCard(c *gin.Context) {
	var req struct {
		PersonID uint   `json:"person_id" binding:"required"`
		CardNo   string `json:"card_no" binding:"required"`
		BankName string `json:"bank_name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "参数错误")
		return
	}
	svc := NewService()
	if err := svc.AddBankCard(req.PersonID, req.CardNo, req.BankName); err != nil {
		response.Error(c, "添加失败")
		return
	}
	response.SuccessMsg(c, "添加成功")
}

func UpdateBankCard(c *gin.Context) {
	id := utils.ParseUint(c.Param("id"))
	var req struct {
		CardNo   string `json:"card_no"`
		BankName string `json:"bank_name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "参数错误")
		return
	}
	svc := NewService()
	if err := svc.UpdateBankCard(id, req.CardNo, req.BankName); err != nil {
		response.Error(c, "更新失败")
		return
	}
	response.SuccessMsg(c, "更新成功")
}

func DeleteBankCard(c *gin.Context) {
	id := utils.ParseUint(c.Param("id"))
	svc := NewService()
	if err := svc.DeleteBankCard(id); err != nil {
		response.Error(c, "删除失败")
		return
	}
	response.SuccessMsg(c, "删除成功")
}

func Export(c *gin.Context) {
	svc := NewService()
	list, _, err := svc.List(1, 100000, c.Query("name"), c.Query("id_card"), c.Query("attendance_group"), c.Query("status"))
	if err != nil {
		response.Error(c, "导出失败")
		return
	}
	exp := excel.NewExporter([]string{"姓名", "身份证号", "性别", "生日", "民族", "籍贯", "住址", "政治面貌", "婚姻状态", "别名"})
	for _, item := range list {
		exp.AddRow([]string{
			fmt.Sprint(item["name"]),
			fmt.Sprint(item["id_card"]),
			fmt.Sprint(item["gender"]),
			fmt.Sprint(item["birthday"]),
			fmt.Sprint(item["nation"]),
			fmt.Sprint(item["native_place"]),
			fmt.Sprint(item["address"]),
			fmt.Sprint(item["political_status"]),
			fmt.Sprint(item["marital_status"]),
			fmt.Sprint(item["alias"]),
		})
	}
	filePath := filepath.Join(os.TempDir(), fmt.Sprintf("person_export_%d.csv", time.Now().Unix()))
	if err := exp.WriteCSV(filePath); err != nil {
		response.Error(c, "导出失败")
		return
	}
	c.Header("Content-Disposition", "attachment; filename=person_export.csv")
	c.Header("Content-Type", "text/csv")
	c.File(filePath)
}
