package handler

import (
	"strconv"
	"time"

	"probig/server/internal/service"
	"probig/server/internal/utils"

	"github.com/gin-gonic/gin"
)

// monthlyListQuery 月度考勤核算列表筛选解析（列表与导出共用）
func monthlyListQuery(c *gin.Context) service.MonthlyListQuery {
	pageReq := utils.BindPage(c)
	personID, _ := strconv.ParseUint(c.Query("person_id"), 10, 64)
	return service.MonthlyListQuery{
		PageNum:  pageReq.PageNum,
		PageSize: pageReq.PageSize,
		Month:    c.Query("month"),
		PersonID: uint(personID),
	}
}

func GetMonthlyList(c *gin.Context) {
	q := monthlyListQuery(c)
	list, total, err := service.GetMonthlyList(q)
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
		req.PersonIDs = service.GetActivePersonIDsInMonth(req.Month)
		if len(req.PersonIDs) == 0 {
			utils.Error(c, "当月无在职人员")
			return
		}
	}

	success, fail, err := service.CalculateMonthlyBatch(c.Request.Context(), req.Month, req.PersonIDs)
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.SuccessWithMsg(c, "核算完成", gin.H{"success": success, "fail": fail})
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

func ExportAttendanceMonthly(c *gin.Context) {
	// 导出严格关联列表视图的当前筛选
	q := monthlyListQuery(c)
	q.PageNum, q.PageSize = 1, 10000
	list, _, err := service.GetMonthlyList(q)
	if err != nil {
		utils.Error(c, "导出失败")
		return
	}

	var rows [][]interface{}
	for _, s := range list {
		rows = append(rows, []interface{}{
			s["belong_month"], s["person_name"], s["salary_days"], s["weighted_base_salary"],
			s["total_work_hours"], s["total_overtime_workday_hours"], s["total_overtime_holiday_hours"],
			s["attendance_salary"], s["overtime_workday_salary"], s["overtime_holiday_salary"],
			s["attendance_bonus"], s["total_violation_count"], exportBool(s["has_personal_leave_month"].(bool)),
			s["last_calc_at"].(time.Time).Format("2006-01-02 15:04:05"),
			map[string]string{"calculated": "已核算", "data_changed": "数据已变动"}[s["status"].(string)],
		})
	}
	writeExcel(c, "月度考勤核算",
		[]string{"月份", "人员", "计薪天数", "加权基本工资", "记出勤工时", "工作日加班工时", "节假日加班工时",
			"出勤工资", "工作日加班工资", "节假日加班工资", "全勤奖", "违纪次数", "有事假", "核算时间", "状态"}, rows,
		monthlyExportFilters(q)...)
}
