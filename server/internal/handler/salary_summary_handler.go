package handler

import (
	"strconv"

	"probig/server/internal/service"
	"probig/server/internal/utils"

	"github.com/gin-gonic/gin"
)

// salarySummaryListQuery 工资汇总列表筛选解析（列表与导出共用）
func salarySummaryListQuery(c *gin.Context) service.SalarySummaryListQuery {
	pageReq := utils.BindPage(c)
	personID, _ := strconv.ParseUint(c.Query("person_id"), 10, 64)
	return service.SalarySummaryListQuery{
		PageNum:  pageReq.PageNum,
		PageSize: pageReq.PageSize,
		Month:    c.Query("month"),
		PersonID: uint(personID),
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

func ExportSalarySummaries(c *gin.Context) {
	// 导出严格关联列表视图的当前筛选
	q := salarySummaryListQuery(c)
	q.PageNum, q.PageSize = 1, 10000
	list, _, _ := service.GetSalarySummaries(q)

	var rows [][]interface{}
	for _, s := range list {
		rows = append(rows, []interface{}{
			s["belong_month"], s["person_name"],
			s["attendance_salary"], s["overtime_workday_salary"], s["overtime_holiday_salary"],
			s["annual_leave_carryover_salary"], s["attendance_bonus"], s["performance_salary"],
			s["post_allowance"], s["meal_allowance"], s["housing_allowance"], s["transport_allowance"],
			s["high_temp_allowance"], s["insurance_compensation"], s["fund_compensation"],
			s["sales_commission"], s["reward_punishment"], s["borrowing_repayment"],
			s["social_security_deduct"], s["housing_fund_deduct"], s["tax_deduct"], s["final_salary"],
		})
	}
	writeExcel(c, "月度工资汇总",
		[]string{"月份", "人员", "出勤工资", "工作日加班工资", "节假日加班工资", "年假结转工资", "全勤奖",
			"绩效工资", "职位津贴", "餐补", "房补", "交通补贴", "高温补贴", "保险补偿", "公积金补偿",
			"提成", "奖惩", "借款还款", "社保代扣", "公积金代扣", "个税代扣", "实发工资"}, rows,
		salarySummaryExportFilters(q)...)
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
