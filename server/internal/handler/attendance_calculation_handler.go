package handler

import (
	"strconv"
	"strings"

	"probig/server/internal/service"
	"probig/server/internal/utils"

	"github.com/gin-gonic/gin"
)

// monthlyListQuery 月度考勤核算列表筛选解析（列表与导出共用）
func monthlyListQuery(c *gin.Context) service.MonthlyListQuery {
	pageReq := utils.BindPage(c)
	personID, _ := strconv.ParseUint(c.Query("person_id"), 10, 64)
	monthsStr := c.Query("months")
	var months []string
	if monthsStr != "" {
		months = strings.Split(monthsStr, ",")
	}
	return service.MonthlyListQuery{
		PageNum:  pageReq.PageNum,
		PageSize: pageReq.PageSize,
		Month:    c.Query("month"),
		Months:   months,
		PersonID: uint(personID),
		Status:   c.Query("status"),
	}
}

func GetMonthlyList(c *gin.Context) {
	q := monthlyListQuery(c)
	list, total, err := service.GetMonthlyList(c.Request.Context(), q)
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	successPage(c, list, total, q.PageNum, q.PageSize)
}

type calculateReq struct {
	Month     string `json:"month" binding:"required"`
	PersonIDs []uint `json:"person_ids"`
}

func CalculateMonthly(c *gin.Context) {
	var req calculateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}

	if len(req.PersonIDs) == 0 {
		req.PersonIDs = service.GetActivePersonIDsInMonth(c.Request.Context(), req.Month)
		if len(req.PersonIDs) == 0 {
			utils.Error(c, "当月无在职人员")
			return
		}
	}

	hasValue, empty, fail, err := service.CalculateMonthlyBatch(c.Request.Context(), req.Month, req.PersonIDs)
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "核算完成", gin.H{"has_value": hasValue, "empty": empty, "fail": fail})
}

// monthlyExportFilters 月度考勤核算导出文件名筛选摘要
func monthlyExportFilters(q service.MonthlyListQuery) []string {
	var parts []string
	if q.Month != "" {
		parts = append(parts, "月份="+q.Month)
	}
	if q.PersonID > 0 {
		parts = append(parts, "人员="+service.PersonName(q.PersonID))
	}
	return parts
}

// attendanceCalcExportFields 月度考勤核算导出字段表（与列表/详情/追溯统一口径）
var attendanceCalcExportFields = []exportField{
	{"belong_month", "月份", nil},
	{"person_name", "人员", nil},
	{"salary_days", "计薪天数", nil},
	{"total_work_hours", "记出勤", hoursToDaysFmt},
	{"weighted_base_salary", "加权基本工资", nil},
	{"attendance_salary", "出勤工资", nil},
	{"total_overtime_workday_hours", "工作日加班", hoursToDaysFmt},
	{"overtime_workday_salary", "工作日加班工资", nil},
	{"total_overtime_holiday_hours", "节假日加班", hoursToDaysFmt},
	{"overtime_holiday_salary", "节假日加班工资", nil},
	{"has_personal_leave_month", "有事假", exportBoolFmt},
	{"total_violation_count", "违纪次数", nil},
	{"attendance_bonus", "全勤奖", nil},
	{"status", "状态", exportStatusFmt},
	{"last_calc_at", "核算时间", exportTimeFmt},
}

// hoursToDaysFmt 工时转天（按配置的每日标准工时，与前端 hoursToDays 口径一致）
func hoursToDaysFmt(v interface{}) interface{} {
	return service.HoursToDays(v.(float64))
}

func ExportAttendanceMonthly(c *gin.Context) {
	// 导出严格关联列表视图的当前筛选
	q := monthlyListQuery(c)
	q.PageNum, q.PageSize = 1, 10000
	list, _, err := service.GetMonthlyList(c.Request.Context(), q)
	if err != nil {
		utils.Error(c, "导出失败")
		return
	}

	rows, headers := buildExportRows(list, attendanceCalcExportFields)
	writeExcel(c, "月度考勤核算", headers, rows, monthlyExportFilters(q)...)
}
