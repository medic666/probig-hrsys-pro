// 年假事件类型中文映射：存储值保持英文（grant/adjust/carryover_deduct），
// 仅展示层映射，与后端导出映射（annualLeaveTypeNames）一致。
export const ANNUAL_LEAVE_TYPES: Record<string, string> = {
  grant: '配发',
  adjust: '人工调整',
  carryover_deduct: '结转扣除',
  休假: '休假',
}

export function annualLeaveTypeText(t?: string): string {
  return (t && ANNUAL_LEAVE_TYPES[t]) || t || '-'
}
