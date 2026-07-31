package handler

import (
	"encoding/json"
	"strconv"

	"probig/server/internal/service"
	"probig/server/internal/utils"

	"github.com/gin-gonic/gin"
)

func GetSalarySummaries(c *gin.Context) {
	pageReq := utils.BindPage(c)
	personID, _ := strconv.ParseUint(c.Query("person_id"), 10, 64)
	list, total, err := service.GetSalarySummaries(c.Query("month"), uint(personID), pageReq.PageNum, pageReq.PageSize)
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	var result []map[string]interface{}
	for i := range list {
		data, _ := json.Marshal(list[i])
		var item map[string]interface{}
		json.Unmarshal(data, &item)
		item["status"] = service.IsSalarySummaryStale(&list[i])
		result = append(result, item)
	}
	utils.Success(c, utils.NewPageResult(result, total, pageReq))
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
	success, fail, skip, err := service.CalculateSalaryBatch(c.Request.Context(), req.Month, req.PersonIDs, userID, userName)
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.Success(c, map[string]interface{}{
		"success": success, "fail": fail, "skip": skip,
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

func ExportSalarySummaries(c *gin.Context) {
	list, _, _ := service.GetSalarySummaries(c.Query("month"), 0, 1, 10000)

	var rows [][]interface{}
	for _, s := range list {
		rows = append(rows, []interface{}{
			s.BelongMonth, service.PersonName(s.PersonID),
			s.AttendanceSalary, s.OvertimeWorkdaySalary, s.OvertimeHolidaySalary,
			s.AnnualLeaveCarryoverSalary, s.AttendanceBonus, s.PerformanceSalary,
			s.PostAllowance, s.MealAllowance, s.HousingAllowance, s.TransportAllowance,
			s.HighTempAllowance, s.InsuranceCompensation, s.FundCompensation,
			s.SalesCommission, s.RewardPunishment, s.BorrowingRepayment,
			s.SocialSecurityDeduct, s.HousingFundDeduct, s.TaxDeduct, s.FinalSalary,
		})
	}
	writeExcel(c, "工资汇总", "salary_summaries",
		[]string{"月份", "人员", "出勤工资", "工作日加班工资", "节假日加班工资", "年假结转工资", "全勤奖",
			"绩效工资", "职位津贴", "餐补", "房补", "交通补贴", "高温补贴", "保险补偿", "公积金补偿",
			"提成", "奖惩", "借款还款", "社保代扣", "公积金代扣", "个税代扣", "实发工资"}, rows)
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
