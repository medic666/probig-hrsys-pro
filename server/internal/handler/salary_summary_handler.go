package handler

import (
	"strconv"
	"strings"

	"probig/server/internal/service"
	"probig/server/internal/utils"

	"github.com/gin-gonic/gin"
)

// salarySummaryListQuery 工资汇总列表筛选解析（列表与导出共用）
func salarySummaryListQuery(c *gin.Context) service.SalarySummaryListQuery {
	pageReq := utils.BindPage(c)
	personID, _ := strconv.ParseUint(c.Query("person_id"), 10, 64)
	monthsStr := c.Query("months")
	var months []string
	if monthsStr != "" {
		months = strings.Split(monthsStr, ",")
	}
	return service.SalarySummaryListQuery{
		PageNum:  pageReq.PageNum,
		PageSize: pageReq.PageSize,
		Month:    c.Query("month"),
		Months:   months,
		PersonID: uint(personID),
		Status:   c.Query("status"),
	}
}

func GetSalarySummaries(c *gin.Context) {
	q := salarySummaryListQuery(c)
	list, total, err := service.GetSalarySummaries(q)
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	successPage(c, list, total, q.PageNum, q.PageSize)
}

type calcSalaryReq struct {
	Month     string `json:"month" binding:"required"`
	PersonIDs []uint `json:"person_ids"`
}

func CalculateSalarySummaries(c *gin.Context) {
	var req calcSalaryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数错误")
		return
	}
	if len(req.PersonIDs) == 0 {
		req.PersonIDs = service.GetActivePersonIDsInMonth(req.Month)
	}
	if len(req.PersonIDs) == 0 {
		utils.Error(c, "当月无在职人员")
		return
	}

	userID := c.GetUint("userID")
	userName := c.GetString("username")
	hasValue, empty, fail, err := service.CalculateSalaryBatch(c.Request.Context(), req.Month, req.PersonIDs, userID, userName)
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.Success(c, map[string]interface{}{
		"has_value": hasValue, "empty": empty, "fail": fail,
	})
}

func GetSalaryVersions(c *gin.Context) {
	personID, _ := strconv.ParseUint(c.Param("personId"), 10, 64)
	month := c.Param("month")
	versions, err := service.GetSalaryVersions(uint(personID), month)
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.Success(c, versions)
}

func GetSalaryVersionDetail(c *gin.Context) {
	versionID, _ := strconv.ParseUint(c.Param("vid"), 10, 64)
	version, err := service.GetSalaryVersionByID(uint(versionID))
	if err != nil {
		utils.Error(c, "版本不存在")
		return
	}
	utils.Success(c, version)
}

// salarySummaryExportFilters 工资汇总导出文件名筛选摘要
func salarySummaryExportFilters(q service.SalarySummaryListQuery) []string {
	var parts []string
	if q.Month != "" {
		parts = append(parts, "月份="+q.Month)
	}
	if q.PersonID > 0 {
		parts = append(parts, "人员="+service.PersonName(q.PersonID))
	}
	return parts
}

// salarySummaryExportFields 月度工资汇总导出字段表（与列表/明细/版本/对比/追溯统一口径）
var salarySummaryExportFields = []exportField{
	{"belong_month", "月份", nil},
	{"person_name", "人员", nil},
	{"attendance_salary", "出勤工资", nil},
	{"overtime_workday_salary", "工作日加班工资", nil},
	{"overtime_holiday_salary", "节假日加班工资", nil},
	{"annual_leave_carryover_salary", "年假结转工资", nil},
	{"attendance_bonus", "全勤奖", nil},
	{"performance_salary", "绩效工资", nil},
	{"post_allowance", "职位津贴", nil},
	{"meal_allowance", "餐补", nil},
	{"housing_allowance", "房补", nil},
	{"transport_allowance", "交通补贴", nil},
	{"high_temp_allowance", "高温补贴", nil},
	{"insurance_compensation", "保险补偿", nil},
	{"fund_compensation", "公积金补偿", nil},
	{"sales_commission", "提成", nil},
	{"reward_punishment", "奖惩", nil},
	{"borrowing_repayment", "预支还款", nil},
	{"social_security_deduct", "社保代扣", nil},
	{"housing_fund_deduct", "公积金代扣", nil},
	{"tax_deduct", "个税代扣", nil},
	{"final_salary", "实发工资", nil},
	{"status", "状态", exportStatusFmt},
	{"last_calc_at", "核算时间", exportTimeStrFmt},
}

func ExportSalarySummaries(c *gin.Context) {
	// 导出严格关联列表视图的当前筛选
	q := salarySummaryListQuery(c)
	q.PageNum, q.PageSize = 1, 10000
	list, _, _ := service.GetSalarySummaries(q)

	rows, headers := buildExportRows(list, salarySummaryExportFields)
	writeExcel(c, "月度工资汇总", headers, rows, salarySummaryExportFilters(q)...)
}

func GetSalaryTrace(c *gin.Context) {
	personID, _ := strconv.ParseUint(c.Param("personId"), 10, 64)
	month := c.Param("month")
	trace, err := service.GetSalaryTrace(uint(personID), month)
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.Success(c, trace)
}
