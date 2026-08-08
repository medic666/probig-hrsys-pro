package handler

import (
	"fmt"
	"strconv"

	"probig/server/internal/model"
	"probig/server/internal/service"
	"probig/server/internal/utils"

	"github.com/gin-gonic/gin"
)

// annualLeaveListQuery 年假事件列表筛选解析（列表与导出共用）
func annualLeaveListQuery(c *gin.Context) service.AnnualLeaveListQuery {
	pageReq := utils.BindPage(c)
	personID, _ := strconv.ParseUint(c.Query("person_id"), 10, 64)
	return service.AnnualLeaveListQuery{
		PageNum:   pageReq.PageNum,
		PageSize:  pageReq.PageSize,
		PersonID:  uint(personID),
		DateStart: c.Query("date_start"),
		DateEnd:   c.Query("date_end"),
		EventType: c.Query("event_type"),
	}
}

func GetAnnualLeaveEvents(c *gin.Context) {
	q := annualLeaveListQuery(c)

	// 服务层已按「已确认」合并考勤年假消费（同日多版本仅最新确认组参与），
	// 此处仅补 account 段来源标记，避免未过滤状态的二次合并导致重复与陈旧 pending 混入
	list, _, err := service.GetAnnualLeaveEventList(c.Request.Context(), q)
	if err != nil {
		utils.Error(c, err.Error())
		return
	}

	if len(list) > 0 {
		for i := range list {
			list[i]["source"] = "account"
		}
	}

	successPage(c, list, int64(len(list)), q.PageNum, q.PageSize)
}

// GetAnnualLeaveEventByID 年假事件完整详情（页面化"编辑=查看"取数）
func GetAnnualLeaveEventByID(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	event, err := service.GetAnnualLeaveEvent(c.Request.Context(), uint(id))
	if err != nil {
		utils.Error(c, "事件不存在")
		return
	}
	utils.Success(c, event)
}

func CreateAnnualLeaveEvent(c *gin.Context) {
	var e model.AnnualLeaveAccountEvent
	if err := c.ShouldBindJSON(&e); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}
	if e.PersonID == 0 || e.EventType == "" {
		utils.BadRequest(c, "人员和事件类型为必填项")
		return
	}
	e.SourceType = "manual"
	if err := service.CreateAnnualLeaveEvent(c.Request.Context(), &e); err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "创建成功", gin.H{"id": e.ID})
}

func UpdateAnnualLeaveEvent(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var e model.AnnualLeaveAccountEvent
	if err := c.ShouldBindJSON(&e); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}
	if err := service.UpdateAnnualLeaveEvent(c.Request.Context(), uint(id), &e); err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "更新成功", nil)
}

func DeleteAnnualLeaveEvent(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.DeleteAnnualLeaveEvent(c.Request.Context(), uint(id)); err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "删除成功", nil)
}

func RestoreAnnualLeaveEvent(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.RestoreAnnualLeaveEvent(c.Request.Context(), uint(id)); err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "恢复成功", nil)
}

func GetDeletedAnnualLeaveEvents(c *gin.Context) {
	pageReq := utils.BindPage(c)
	list, _, err := service.GetDeletedAnnualLeaveEvents(c.Request.Context(), pageReq.PageNum, pageReq.PageSize)
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	var filtered []model.AnnualLeaveAccountEvent
	for _, e := range list {
		if e.SourceType != "system_period" {
			filtered = append(filtered, e)
		}
	}
	successPage(c, filtered, int64(len(filtered)), pageReq.PageNum, pageReq.PageSize)
}

// annualLeaveTypeNames 年假事件类型中文名（与前端展示映射一致；存储值保持英文）
var annualLeaveTypeNames = map[string]string{
	"grant":            "配发",
	"adjust":           "人工调整",
	"carryover_deduct": "结转扣除",
}

func annualLeaveTypeName(v interface{}) interface{} {
	if n, ok := annualLeaveTypeNames[fmt.Sprint(v)]; ok {
		return n
	}
	return v
}

// annualLeaveExportFilters 年假事件导出文件名筛选摘要
func annualLeaveExportFilters(q service.AnnualLeaveListQuery) []string {
	var parts []string
	if q.PersonID > 0 {
		parts = append(parts, "人员="+service.PersonName(q.PersonID))
	}
	if q.EventType != "" {
		parts = append(parts, "类型="+fmt.Sprint(annualLeaveTypeName(q.EventType)))
	}
	if p := dateRangePiece("日期", q.DateStart, q.DateEnd); p != "" {
		parts = append(parts, p)
	}
	return parts
}

func ExportAnnualLeaveEvents(c *gin.Context) {
	// 导出严格关联列表视图的当前筛选
	q := annualLeaveListQuery(c)
	q.PageNum, q.PageSize = 1, 10000
	list, _, _ := service.GetAnnualLeaveEventList(c.Request.Context(), q)

	var rows [][]interface{}
	for _, e := range list {
		rows = append(rows, []interface{}{
			e["person_name"], annualLeaveTypeName(e["event_type"]), e["source_type"], e["hours"], e["effective_date"], e["remark"],
		})
	}
	writeExcel(c, "年假事件",
		[]string{"人员", "类型", "来源", "变动时长(小时)", "生效日期", "备注"}, rows,
		annualLeaveExportFilters(q)...)
}
