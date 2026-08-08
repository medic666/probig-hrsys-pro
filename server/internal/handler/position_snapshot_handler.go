package handler

import (
	"strconv"

	"probig/server/internal/service"
	"probig/server/internal/utils"

	"github.com/gin-gonic/gin"
)

func GetPersonCurrentPosition(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.EnsureOwnPerson(c.Request.Context(), uint(id)); err != nil {
		utils.Error(c, err.Error())
		return
	}
	snapshot, err := service.GetCurrentPositionSnapshot(uint(id))
	if err != nil {
		utils.Error(c, "暂无职务信息")
		return
	}
	companyNameMap := service.CompanyNameMap([]uint{snapshot.CompanyID})
	utils.Success(c, gin.H{
		"id":                   snapshot.ID,
		"person_id":            snapshot.PersonID,
		"effective_start_date": snapshot.EffectiveStartDate,
		"effective_end_date":   snapshot.EffectiveEndDate,
		"is_active":            snapshot.IsActive,
		"entry_date":           snapshot.EntryDate,
		"leave_date":           snapshot.LeaveDate,
		"attendance_group":     snapshot.AttendanceGroup,
		"has_annual_leave":     snapshot.HasAnnualLeave,
		"has_attendance_bonus": snapshot.HasAttendanceBonus,
		"company_id":           snapshot.CompanyID,
		"company_name":         companyNameMap[snapshot.CompanyID],
		"department":           snapshot.Department,
		"position":             snapshot.Position,
		"base_salary":          snapshot.BaseSalary,
		"performance_salary":   snapshot.PerformanceSalary,
		"salary_days":          snapshot.SalaryDays,
		"post_allowance":       snapshot.PostAllowance,
		"meal_allowance":       snapshot.MealAllowance,
		"housing_allowance":    snapshot.HousingAllowance,
		"transport_allowance":  snapshot.TransportAllowance,
		"high_temp_allowance":  snapshot.HighTempAllowance,
		"insurance_compensation": snapshot.InsuranceCompensation,
		"fund_compensation":    snapshot.FundCompensation,
		"social_security_deduct": snapshot.SocialSecurityDeduct,
		"housing_fund_deduct":  snapshot.HousingFundDeduct,
		"last_calc_at":         snapshot.LastCalcAt,
	})
}

func GetPersonPositionHistory(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := service.EnsureOwnPerson(c.Request.Context(), uint(id)); err != nil {
		utils.Error(c, err.Error())
		return
	}
	history, err := service.GetPositionSnapshotHistory(uint(id))
	if err != nil {
		utils.Error(c, err.Error())
		return
	}
	utils.Success(c, history)
}
