import request from '@/utils/request'

export function getSalaryEvents(params: any) { return request.get('/salary-events', { params }) }
// getSalaryAdvanceBalances 预支待还：person_id → balance（累计工资预支 + 累计预支还款，还款为负）
export function getSalaryAdvanceBalances() { return request.get('/salary-events/advance-balances') }
// getSalarySummariesBadges 工资汇总徽章（默认上月）：person_id → gray/green/orange
export function getSalarySummariesBadges(month?: string) {
  return request.get('/salary-summaries/badges', { params: month ? { month } : {} })
}
export function getSalaryEvent(id: number) { return request.get(`/salary-events/${id}`) }
export function exportSalaryEvents(params: any) { return request.get('/salary-events/export', { params, responseType: 'blob' }) }
export function createSalaryEvent(data: any) { return request.post('/salary-events', data) }
export function updateSalaryEvent(id: number, data: any) { return request.put(`/salary-events/${id}`, data) }
export function deleteSalaryEvent(id: number) { return request.delete(`/salary-events/${id}`) }
export function restoreSalaryEvent(id: number) { return request.post(`/salary-events/${id}/restore`) }
export function getDeletedSalaryEvents(params: any) { return request.get('/salary-events/trash', { params }) }
export function getSalarySummaries(params: any) { return request.get('/salary-summaries', { params }) }
export function exportSalarySummaries(params: any) { return request.get('/salary-summaries/export', { params, responseType: 'blob' }) }
export function calculateSalaries(data: any) { return request.post('/salary-summaries/calculate', data) }
export function getSalaryVersions(personId: number, month: string) { return request.get(`/salary-summaries/${personId}/${month}/versions`) }
export function getSalaryTrace(personId: number, month: string) { return request.get(`/salary-summaries/${personId}/${month}/trace`) }
export function getSalaryVersionDetail(vid: number) { return request.get(`/salary-summaries/versions/${vid}`) }
