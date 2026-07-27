package handler

import (
	"strconv"
	"strings"
	"time"

	"probig/server/internal/utils"

	"github.com/gin-gonic/gin"
)

type MockItem struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Category  string    `json:"category"`
	Status    int       `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type MockNameItem struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type FormData struct {
	Name     string `json:"name" binding:"required"`
	Category string `json:"category"`
	Number   int    `json:"number"`
	Remark   string `json:"remark"`
}

func TestPagination(c *gin.Context) {
	pageReq := utils.BindPage(c)

	var mockData []MockItem
	for i := 1; i <= 55; i++ {
		mockData = append(mockData, MockItem{
			ID:        uint(i),
			Name:      "测试记录" + strconv.Itoa(i),
			Category:  "分类" + strconv.Itoa(i%3+1),
			Status:    i % 3,
			CreatedAt: time.Now().AddDate(0, 0, -i),
		})
	}

	nameFilter := c.Query("name")
	categoryFilter := c.Query("category")
	var filtered []MockItem
	for _, item := range mockData {
		if nameFilter != "" && !strings.Contains(item.Name, nameFilter) {
			continue
		}
		if categoryFilter != "" && item.Category != categoryFilter {
			continue
		}
		filtered = append(filtered, item)
	}

	total := int64(len(filtered))
	start := pageReq.Offset()
	end := start + pageReq.PageSize
	if start > len(filtered) {
		start = len(filtered)
	}
	if end > len(filtered) {
		end = len(filtered)
	}

	utils.Success(c, utils.NewPageResult(filtered[start:end], total, pageReq))
}

func TestNames(c *gin.Context) {
	keyword := c.Query("keyword")
	var mockNames []MockNameItem
	names := []string{"张三", "张明", "李四", "李华", "王五", "王芳", "赵六", "孙七", "周八", "吴九"}
	for i, name := range names {
		if keyword == "" || strings.Contains(name, keyword) {
			mockNames = append(mockNames, MockNameItem{ID: uint(i + 1), Name: name})
		}
	}
	utils.Success(c, mockNames)
}

func TestFormSubmit(c *gin.Context) {
	var form FormData
	if err := c.ShouldBindJSON(&form); err != nil {
		utils.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	utils.SuccessWithMsg(c, "提交成功", form)
}

type TrashItem struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	DeletedAt time.Time `json:"deleted_at"`
}

func TestTrashList(c *gin.Context) {
	pageReq := utils.BindPage(c)

	var items []TrashItem
	for i := 1; i <= 12; i++ {
		items = append(items, TrashItem{
			ID:        uint(i),
			Name:      "已删除记录" + strconv.Itoa(i),
			DeletedAt: time.Now().AddDate(0, 0, -i),
		})
	}

	total := int64(len(items))
	start := pageReq.Offset()
	end := start + pageReq.PageSize
	if start > len(items) {
		start = len(items)
	}
	if end > len(items) {
		end = len(items)
	}

	utils.Success(c, utils.NewPageResult(items[start:end], total, pageReq))
}

func TestTrashRestore(c *gin.Context) {
	var req struct {
		ID uint `json:"id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}
	utils.SuccessWithMsg(c, "恢复成功", nil)
}
