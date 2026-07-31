package handler

import (
	"encoding/json"
	"strconv"
	"time"

	"probig/server/internal/dao"
	"probig/server/internal/model"
	"probig/server/internal/service"
	"probig/server/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
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
		monthStart, _ := time.Parse("2006-01", req.Month)
		monthEnd := monthStart.AddDate(0, 1, -1)
		dao.DB.Model(&model.PositionSnapshot{}).
			Select("DISTINCT person_id").
			Where("effective_start_date <= ? AND effective_end_date >= ? AND is_active = true",
				utils.DateOnlyFromTime(monthEnd), utils.DateOnlyFromTime(monthStart)).
			Pluck("person_id", &req.PersonIDs)
	}
	if len(req.PersonIDs) == 0 {
		utils.Error(c, "当月无在职人员")
		return
	}

	userID := c.GetUint("userID")
	userName := c.GetString("username")
	success, fail, skip, err := service.CalculateSalaryBatch(req.Month, req.PersonIDs, userID, userName)
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
	f := excelize.NewFile()
	defer f.Close()
	sheet := "工资汇总"
	f.SetSheetName("Sheet1", sheet)
	headers := []string{"月份", "人员", "出勤工资", "工作日加班工资", "节假日加班工资", "年假结转工资", "全勤奖", "绩效工资", "职位津贴", "餐补", "房补", "交通补贴", "高温补贴", "保险补偿", "公积金补偿", "提成", "奖惩", "借款还款", "社保代扣", "公积金代扣", "个税代扣", "实发工资"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}
	for i, s := range list {
		row := i + 2
		f.SetCellValue(sheet, cellName(1, row), s.BelongMonth)
		f.SetCellValue(sheet, cellName(2, row), s.PersonID)
		f.SetCellValue(sheet, cellName(3, row), s.AttendanceSalary)
		f.SetCellValue(sheet, cellName(4, row), s.OvertimeWorkdaySalary)
		f.SetCellValue(sheet, cellName(5, row), s.OvertimeHolidaySalary)
		f.SetCellValue(sheet, cellName(6, row), s.AnnualLeaveCarryoverSalary)
		f.SetCellValue(sheet, cellName(7, row), s.AttendanceBonus)
		f.SetCellValue(sheet, cellName(8, row), s.PerformanceSalary)
		f.SetCellValue(sheet, cellName(9, row), s.PostAllowance)
		f.SetCellValue(sheet, cellName(10, row), s.MealAllowance)
		f.SetCellValue(sheet, cellName(11, row), s.HousingAllowance)
		f.SetCellValue(sheet, cellName(12, row), s.TransportAllowance)
		f.SetCellValue(sheet, cellName(13, row), s.HighTempAllowance)
		f.SetCellValue(sheet, cellName(14, row), s.InsuranceCompensation)
		f.SetCellValue(sheet, cellName(15, row), s.FundCompensation)
		f.SetCellValue(sheet, cellName(16, row), s.SalesCommission)
		f.SetCellValue(sheet, cellName(17, row), s.RewardPunishment)
		f.SetCellValue(sheet, cellName(18, row), s.BorrowingRepayment)
		f.SetCellValue(sheet, cellName(19, row), s.SocialSecurityDeduct)
		f.SetCellValue(sheet, cellName(20, row), s.HousingFundDeduct)
		f.SetCellValue(sheet, cellName(21, row), s.TaxDeduct)
		f.SetCellValue(sheet, cellName(22, row), s.FinalSalary)
	}
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename=salary_"+time.Now().Format("20060102")+".xlsx")
	f.Write(c.Writer)
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
