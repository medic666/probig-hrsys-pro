package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/medic666/probig/internal/response"
	"github.com/xuri/excelize/v2"
)

type ExportHandler struct {
	db *sqlx.DB
}

func NewExportHandler(db *sqlx.DB) *ExportHandler {
	return &ExportHandler{db: db}
}

func (h *ExportHandler) ExportAttendance(c *gin.Context) {
	period := c.Query("period")

	rows, err := h.db.Queryx(
		`SELECT as2.*, e.name as person_name
		 FROM attendance_summaries as2
		 JOIN entities e ON e.id = as2.person_id
		 WHERE (? = '' OR as2.period = ?)
		 ORDER BY as2.period DESC, e.name ASC`,
		period, period,
	)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	defer rows.Close()

	f := excelize.NewFile()
	defer f.Close()

	headers := []string{"人员", "期间", "普通出勤", "补班出勤", "补休", "事假", "病假", "年假", "法定假", "福利假", "工作日加班", "节假日加班", "缺卡", "迟到", "早退", "年假配发", "年假结转", "违纪次数"}
	for i, h2 := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue("Sheet1", cell, h2)
	}

	rowIdx := 2
	for rows.Next() {
		row := map[string]interface{}{}
		rows.MapScan(row)
		cols := []string{
			"person_name", "period",
			"normal_attendance_days", "supplementary_attendance_days",
			"compensatory_leave_days", "personal_leave_days", "sick_leave_days",
			"annual_leave_days", "statutory_leave_days", "welfare_leave_days",
			"workday_overtime_days", "holiday_overtime_days",
			"missing_clock_count", "late_count", "early_leave_count",
			"annual_leave_allot", "annual_leave_carryover", "violation_count",
		}
		for i, col := range cols {
			cell, _ := excelize.CoordinatesToCellName(i+1, rowIdx)
			f.SetCellValue("Sheet1", cell, row[col])
		}
		rowIdx++
	}

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename=attendance_summary.xlsx")
	f.Write(c.Writer)
}

func (h *ExportHandler) ExportSalary(c *gin.Context) {
	period := c.Query("period")

	rows, err := h.db.Queryx(
		`SELECT ss.*, e.name as person_name
		 FROM salary_summaries ss
		 JOIN entities e ON e.id = ss.person_id
		 WHERE (? = '' OR ss.period = ?)
		 ORDER BY ss.period DESC, e.name ASC`,
		period, period,
	)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	defer rows.Close()

	f := excelize.NewFile()
	defer f.Close()

	headers := []string{"人员", "期间", "出勤工资", "全勤奖", "加班工资", "绩效工资", "职位津贴", "餐补", "房补", "交通补贴", "高温补贴", "保险补偿", "公积金补偿", "社保代扣", "公积金代扣", "个税扣除", "借款扣除", "奖惩", "其他调整", "实发合计"}
	for i, h2 := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue("Sheet1", cell, h2)
	}

	rowIdx := 2
	for rows.Next() {
		row := map[string]interface{}{}
		rows.MapScan(row)
		cols := []string{
			"person_name", "period",
			"attendance_salary", "full_attendance_bonus", "overtime_salary",
			"performance_salary", "position_allowance", "meal_subsidy",
			"housing_subsidy", "transport_subsidy", "heat_subsidy",
			"insurance_compensation", "housing_fund_compensation",
			"social_insurance_deduct", "housing_fund_deduct",
			"tax_deduct", "loan_deduct", "reward_punish",
			"other_adjustments", "total_salary",
		}
		for i, col := range cols {
			cell, _ := excelize.CoordinatesToCellName(i+1, rowIdx)
			f.SetCellValue("Sheet1", cell, row[col])
		}
		rowIdx++
	}

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename=salary_summary.xlsx")
	f.Write(c.Writer)
}
