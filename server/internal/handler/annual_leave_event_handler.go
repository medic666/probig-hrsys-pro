package handler

import (
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

	list, _, err := service.GetAnnualLeaveEventList(q)
	if err != nil {
		utils.Error(c, err.Error())
		return
	}

	if len(list) > 0 {
		for i := range list {
			list[i]["source"] = "account"
		}
	}

	if q.EventType == "" || q.EventType == "休假" {
		attDailyList, _, _ := service.GetAttendanceDailyList(service.AttendanceDailyListQuery{
			PageNum: q.PageNum, PageSize: q.PageSize,
			PersonID: q.PersonID, DateStart: q.DateStart, DateEnd: q.DateEnd,
		})
		for _, daily := range attDailyList {
			if details, ok := daily["details"].([]map[string]interface{}); ok {
				for _, d := range details {
					if sub, _ := d["sub_type"].(string); sub == "年假" {
						h, _ := d["hours"].(float64)
						d["event_type"] = "休假(年假)"
						d["source_type"] = "attendance"
						d["source"] = "attendance"
						d["person_id"] = daily["person_id"]
						d["person_name"] = daily["person_name"]
						d["event_date"] = daily["event_date"]
						d["hours"] = -h
						list = append(list, d)
					}
				}
			}
		}
	}

	utils.Success(c, utils.NewPageResult(list, int64(len(list)), utils.PageRequest{PageNum: q.PageNum, PageSize: q.PageSize}))
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
	list, _, err := service.GetDeletedAnnualLeaveEvents(pageReq.PageNum, pageReq.PageSize)
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
	utils.Success(c, utils.NewPageResult(filtered, int64(len(filtered)), pageReq))
}

// annualLeaveExportFilters 年假事件导出文件名筛选摘要
func annualLeaveExportFilters(q service.AnnualLeaveListQuery) []string {
	var parts []string
	if q.PersonID > 0 {
		parts = append(parts, "人员="+service.PersonName(q.PersonID))
	}
	if q.EventType != "" {
		parts = append(parts, "类型="+map[string]string{"grant": "配发", "adjust": "人工调整", "carryover_deduct": "结转扣除"}[q.EventType])
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
	list, _, _ := service.GetAnnualLeaveEventList(q)

	var rows [][]interface{}
	for _, e := range list {
		rows = append(rows, []interface{}{
			e["person_name"], e["event_type"], e["source_type"], e["hours"], e["effective_date"], e["remark"],
		})
	}
	writeExcel(c, "年假事件",
		[]string{"人员", "类型", "来源", "变动时长(小时)", "生效日期", "备注"}, rows,
		annualLeaveExportFilters(q)...)
}
