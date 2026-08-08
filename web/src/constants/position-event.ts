// 职务事件调整字段分组（中文 label 与类型单一来源）：
// 表单调整项、详情页查看态展示共用
export const adjustFieldGroups = [
  {
    label: '岗位信息',
    fields: [
      { key: 'company_id', label: '公司组', type: 'company' },
      { key: 'department', label: '部门', type: 'text' },
      { key: 'position', label: '职位', type: 'text' },
    ],
  },
  {
    label: '考勤/福利',
    fields: [
      { key: 'attendance_group', label: '考勤组', type: 'text' },
      { key: 'has_annual_leave', label: '享有年假', type: 'bool' },
      { key: 'has_attendance_bonus', label: '享有全勤奖', type: 'bool' },
    ],
  },
  {
    label: '薪资参数',
    fields: [
      { key: 'base_salary', label: '基本工资', type: 'number', precision: 2 },
      { key: 'performance_salary', label: '绩效工资基数', type: 'number', precision: 2 },
      { key: 'salary_days', label: '计薪天数', type: 'number', precision: 1 },
    ],
  },
  {
    label: '补贴',
    fields: [
      { key: 'post_allowance', label: '职位津贴', type: 'number', precision: 2 },
      { key: 'meal_allowance', label: '餐补', type: 'number', precision: 2 },
      { key: 'housing_allowance', label: '房补', type: 'number', precision: 2 },
      { key: 'transport_allowance', label: '交通补贴', type: 'number', precision: 2 },
      { key: 'high_temp_allowance', label: '高温补贴', type: 'number', precision: 2 },
    ],
  },
  {
    label: '补偿/代扣',
    fields: [
      { key: 'insurance_compensation', label: '保险补偿', type: 'number', precision: 2 },
      { key: 'fund_compensation', label: '公积金补偿', type: 'number', precision: 2 },
      { key: 'social_security_deduct', label: '社保代扣', type: 'number', precision: 2 },
      { key: 'housing_fund_deduct', label: '公积金代扣', type: 'number', precision: 2 },
    ],
  },
]
