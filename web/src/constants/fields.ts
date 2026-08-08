import { hoursToDays } from '@/utils'

export interface FieldDef {
  key: string
  label: string
  width?: string | number
  formatter?: (row: any) => any
}

// 月度考勤核算统一字段（列表/导出/详情/追溯共用口径，工时按天展示）
export const ATTENDANCE_CALC_FIELDS: FieldDef[] = [
  { key: 'belong_month', label: '月份', width: 90 },
  { key: 'person_name', label: '人员', width: 80 },
  { key: 'salary_days', label: '计薪天数', width: 80 },
  { key: 'total_work_hours', label: '记出勤', width: 90, formatter: (r) => hoursToDays(r.total_work_hours || 0).toFixed(2) },
  { key: 'weighted_base_salary', label: '加权基本工资', width: 100 },
  { key: 'attendance_salary', label: '出勤工资', width: 90 },
  { key: 'total_overtime_workday_hours', label: '工作日加班', width: 110, formatter: (r) => hoursToDays(r.total_overtime_workday_hours || 0).toFixed(2) },
  { key: 'overtime_workday_salary', label: '工作日加班工资', width: 110 },
  { key: 'total_overtime_holiday_hours', label: '节假日加班', width: 110, formatter: (r) => hoursToDays(r.total_overtime_holiday_hours || 0).toFixed(2) },
  { key: 'overtime_holiday_salary', label: '节假日加班工资', width: 110 },
  { key: 'has_personal_leave_month', label: '有事假', width: 70, formatter: (r) => (r.has_personal_leave_month ? '是' : '否') },
  { key: 'total_violation_count', label: '违纪次数', width: 80 },
  { key: 'attendance_bonus', label: '全勤奖', width: 80 },
]

// 月度工资汇总统一字段（22 项）
export const SALARY_SUMMARY_FIELDS: FieldDef[] = [
  { key: 'belong_month', label: '月份', width: 90 },
  { key: 'person_name', label: '人员', width: 80 },
  { key: 'attendance_salary', label: '出勤工资', width: 90 },
  { key: 'overtime_workday_salary', label: '工作日加班工资', width: 110 },
  { key: 'overtime_holiday_salary', label: '节假日加班工资', width: 110 },
  { key: 'annual_leave_carryover_salary', label: '年假结转工资', width: 100 },
  { key: 'attendance_bonus', label: '全勤奖', width: 80 },
  { key: 'performance_salary', label: '绩效工资', width: 90 },
  { key: 'post_allowance', label: '职位津贴', width: 90 },
  { key: 'meal_allowance', label: '餐补', width: 70 },
  { key: 'housing_allowance', label: '房补', width: 70 },
  { key: 'transport_allowance', label: '交通补贴', width: 90 },
  { key: 'high_temp_allowance', label: '高温补贴', width: 90 },
  { key: 'insurance_compensation', label: '保险补偿', width: 90 },
  { key: 'fund_compensation', label: '公积金补偿', width: 90 },
  { key: 'sales_commission', label: '提成', width: 80 },
  { key: 'reward_punishment', label: '奖惩', width: 80 },
  { key: 'borrowing_repayment', label: '预支还款', width: 90 },
  { key: 'social_security_deduct', label: '社保代扣', width: 90 },
  { key: 'housing_fund_deduct', label: '公积金代扣', width: 90 },
  { key: 'tax_deduct', label: '个税代扣', width: 80 },
  { key: 'final_salary', label: '实发工资', width: 100 },
]

// 字段表 → ProTable 列（列表视图复用，tail 追加状态/核算时间等模块差异列）
export function fieldsToColumns(fields: FieldDef[], tail: any[] = []) {
  return [...fields.map((f) => ({ prop: f.key, label: f.label, width: f.width, formatter: f.formatter })), ...tail]
}
