package service

import "gorm.io/gorm"

func seedDefaultConfigs(db *gorm.DB) error {
	defaults := []struct {
		Key         string
		Value       string
		Name        string
		Desc        string
		ValueType   string
		OptionValue string
	}{
		{Key: "system.page_size", Value: "20", Name: "分页默认条数", Desc: "列表分页默认每页条数", ValueType: "number"},
		{Key: "system.export_max", Value: "10000", Name: "导出最大条数", Desc: "导出Excel最大记录条数", ValueType: "number"},
		{Key: "system.upload_max_size", Value: "10", Name: "文件上传大小限制(MB)", Desc: "单个文件上传最大大小", ValueType: "number"},
		{Key: "system.upload_path", Value: "./data/uploads", Name: "文件存储根路径", Desc: "上传文件存储目录", ValueType: "string"},
		{Key: "system.work_hours_per_day", Value: "8", Name: "计薪小时基准", Desc: "每日标准计薪小时数", ValueType: "number"},
		{Key: "attendance.min_leave_unit", Value: "0.5", Name: "最小请假单位", Desc: "请假最小时间单位(小时)", ValueType: "number"},
		{Key: "attendance.sick_leave_ratio", Value: "0.8", Name: "病假系数", Desc: "病假折算记出勤工时的系数", ValueType: "number"},
		{Key: "attendance.overtime_workday_ratio", Value: "1.5", Name: "工作日加班系数", Desc: "工作日加班工资倍数", ValueType: "number"},
		{Key: "attendance.overtime_holiday_ratio", Value: "2.0", Name: "节假日加班系数", Desc: "节假日加班工资倍数", ValueType: "number"},
		{Key: "attendance.full_attendance_bonus", Value: "50", Name: "全勤奖日标准", Desc: "全勤奖每日标准金额", ValueType: "number"},
		{Key: "attendance.high_temp_months", Value: `["06","07","08","09"]`, Name: "高温补贴发放月份", Desc: "高温补贴发放月份列表", ValueType: "select", OptionValue: `["06","07","08","09"]`},
		{Key: "annual_leave.yearly_hours", Value: "40", Name: "年假年度额度", Desc: "每年标准年假小时数", ValueType: "number"},
		{Key: "annual_leave.cycle_type", Value: "entry_anniversary", Name: "年假周期规则", Desc: "年假周期计算规则", ValueType: "select", OptionValue: `["entry_anniversary","calendar_year"]`},
	}

	for _, d := range defaults {
		var count int64
		db.Model(&struct{ ID uint }{}).Table("sys_config").Where("config_key = ?", d.Key).Count(&count)
		if count == 0 {
			db.Table("sys_config").Create(map[string]interface{}{
				"config_key":    d.Key,
				"config_value":  d.Value,
				"config_name":   d.Name,
				"config_desc":   d.Desc,
				"value_type":    d.ValueType,
				"option_values": d.OptionValue,
			})
		}
	}

	return nil
}
