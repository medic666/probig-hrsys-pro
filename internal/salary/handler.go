package salary

import (
	"strconv"

	"probig/internal/pkg/excel"
	"probig/internal/pkg/response"
	"probig/internal/pkg/utils"

	"github.com/gin-gonic/gin"
)

func ListEventsHandler(c *gin.Context) {
	var filter SalaryEventFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		response.Error(c, response.ParamError, "参数错误: "+err.Error())
		return
	}

	events, total, err := ListEvents(filter)
	if err != nil {
		response.Error(c, response.InternalError, "查询失败: "+err.Error())
		return
	}

	response.Success(c, response.PageResult{List: events, Total: total})
}

func GetEventHandler(c *gin.Context) {
	id, err := utils.ParseID(c, "id")
	if err != nil {
		response.Error(c, response.ParamError, "参数错误")
		return
	}

	event, err := GetSalaryEventByID(id)
	if err != nil {
		response.Error(c, response.NotFound, "工资事件不存在")
		return
	}

	response.Success(c, event)
}

func CreateEventHandler(c *gin.Context) {
	var req SalaryEventCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.ParamError, "参数错误: "+err.Error())
		return
	}

	event := &SalaryEvent{
		PersonID:    req.PersonID,
		BelongMonth: req.BelongMonth,
		EventType:   req.EventType,
		Amount:      req.Amount,
		EventName:   req.EventName,
		Remark:      req.Remark,
	}

	operatorID := utils.GetUserID(c)
	operatorName := utils.GetUsername(c)
	clientIP := c.ClientIP()

	if err := CreateEvent(event, operatorID, operatorName, clientIP); err != nil {
		response.Error(c, response.InternalError, "创建失败: "+err.Error())
		return
	}

	response.SuccessWithMsg(c, "创建成功", event)
}

func UpdateEventHandler(c *gin.Context) {
	id, err := utils.ParseID(c, "id")
	if err != nil {
		response.Error(c, response.ParamError, "参数错误")
		return
	}

	var req SalaryEventUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.ParamError, "参数错误: "+err.Error())
		return
	}

	operatorID := utils.GetUserID(c)
	operatorName := utils.GetUsername(c)
	clientIP := c.ClientIP()

	if err := UpdateEvent(id, &req, operatorID, operatorName, clientIP); err != nil {
		response.Error(c, response.InternalError, "更新失败: "+err.Error())
		return
	}

	event, _ := GetSalaryEventByID(id)
	response.SuccessWithMsg(c, "更新成功", event)
}

func DeleteEventHandler(c *gin.Context) {
	id, err := utils.ParseID(c, "id")
	if err != nil {
		response.Error(c, response.ParamError, "参数错误")
		return
	}

	operatorID := utils.GetUserID(c)
	operatorName := utils.GetUsername(c)
	clientIP := c.ClientIP()

	if err := DeleteEvent(id, operatorID, operatorName, clientIP); err != nil {
		response.Error(c, response.InternalError, "删除失败: "+err.Error())
		return
	}

	response.SuccessWithMsg(c, "删除成功", nil)
}

func RestoreEventHandler(c *gin.Context) {
	id, err := utils.ParseID(c, "id")
	if err != nil {
		response.Error(c, response.ParamError, "参数错误")
		return
	}

	operatorID := utils.GetUserID(c)
	operatorName := utils.GetUsername(c)
	clientIP := c.ClientIP()

	if err := RestoreEvent(id, operatorID, operatorName, clientIP); err != nil {
		response.Error(c, response.InternalError, "恢复失败: "+err.Error())
		return
	}

	response.SuccessWithMsg(c, "恢复成功", nil)
}

func ListTrashHandler(c *gin.Context) {
	var filter SalaryEventFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		response.Error(c, response.ParamError, "参数错误: "+err.Error())
		return
	}

	events, total, err := ListTrashEvents(filter)
	if err != nil {
		response.Error(c, response.InternalError, "查询失败: "+err.Error())
		return
	}

	response.Success(c, response.PageResult{List: events, Total: total})
}

func ListSummariesHandler(c *gin.Context) {
	var filter SalarySummaryFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		response.Error(c, response.ParamError, "参数错误: "+err.Error())
		return
	}

	summaries, total, err := ListSummaries(filter)
	if err != nil {
		response.Error(c, response.InternalError, "查询失败: "+err.Error())
		return
	}

	response.Success(c, response.PageResult{List: summaries, Total: total})
}

func CalcSummaryHandler(c *gin.Context) {
	var req CalcRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.ParamError, "参数错误: "+err.Error())
		return
	}

	operatorID := utils.GetUserID(c)
	operatorName := utils.GetUsername(c)

	succeeded, failed, errs := CalcSalarySummary(req.BelongMonth, req.PersonIDs, operatorID, operatorName)

	msg := "核算完成"
	if failed > 0 {
		msg = "核算完成，成功" + strconv.Itoa(succeeded) + "条，失败" + strconv.Itoa(failed) + "条"
	}

	response.SuccessWithMsg(c, msg, gin.H{
		"succeeded": succeeded,
		"failed":    failed,
		"errors":    errs,
	})
}

func GetSummaryDetailHandler(c *gin.Context) {
	id, err := utils.ParseID(c, "id")
	if err != nil {
		response.Error(c, response.ParamError, "参数错误")
		return
	}

	summary, err := GetSummaryDetail(id)
	if err != nil {
		response.Error(c, response.NotFound, "工资汇总不存在")
		return
	}

	response.Success(c, summary)
}

func ExportSummariesHandler(c *gin.Context) {
	var filter SalarySummaryFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		response.Error(c, response.ParamError, "参数错误: "+err.Error())
		return
	}

	summaries, err := GetAllSummaries(filter)
	if err != nil {
		response.Error(c, response.InternalError, "查询失败: "+err.Error())
		return
	}

	headers := []string{"月份", "人员姓名", "计薪天数", "加权基本工资", "记出勤工时", "工作日加班工时", "节假日加班工时",
		"出勤工资", "工作日加班工资", "节假日加班工资", "未休年假折算", "全勤奖", "绩效工资",
		"职位津贴", "餐补", "房补", "交通补贴", "高温补贴", "保险补偿", "公积金补偿",
		"调整项合计", "社保代扣", "公积金代扣", "个税代扣", "实发工资", "状态"}

	var rows [][]interface{}
	for _, s := range summaries {
		rows = append(rows, []interface{}{
			s.BelongMonth, s.PersonName, s.SalaryDays, s.WeightedBaseSalary,
			s.TotalWorkHours, s.TotalOvertimeWorkdayHours, s.TotalOvertimeHolidayHours,
			s.AttendanceSalary, s.OvertimeWorkdaySalary, s.OvertimeHolidaySalary,
			s.AnnualLeaveCarryoverSalary, s.AttendanceBonus, s.PerformanceSalary,
			s.PostAllowance, s.MealAllowance, s.HousingAllowance, s.TransportAllowance,
			s.HighTempAllowance, s.InsuranceCompensation, s.FundCompensation,
			s.TotalAdjustment, s.SocialSecurityDeduct, s.HousingFundDeduct, s.TaxDeduct,
			s.FinalSalary, s.Status,
		})
	}

	f, err := excel.ExportExcel([]excel.SheetData{
		{SheetName: "工资汇总", Headers: headers, Rows: rows},
	})
	if err != nil {
		response.Error(c, response.InternalError, "导出失败: "+err.Error())
		return
	}

	excel.WriteToResponse(c, f, "salary_summaries.xlsx")
}
