import request from '@/utils/request'

export function getAnnualLeaveEvents(params: any) {
  return request.get('/annual-leave-events', { params })
}
// getAnnualLeaveEventBadges 年假事件徽章：周年月且上月未结转 → orange，否则 green
export function getAnnualLeaveEventBadges() {
  return request.get('/annual-leave-events/badges')
}
export function getAnnualLeaveEvent(id: number) {
  return request.get(`/annual-leave-events/${id}`)
}
export function exportAnnualLeaveEvents(params: any) {
  return request.get('/annual-leave-events/export', { params, responseType: 'blob' })
}
export function createAnnualLeaveEvent(data: any) {
  return request.post('/annual-leave-events', data)
}
export function updateAnnualLeaveEvent(id: number, data: any) {
  return request.put(`/annual-leave-events/${id}`, data)
}
export function deleteAnnualLeaveEvent(id: number) {
  return request.delete(`/annual-leave-events/${id}`)
}
export function restoreAnnualLeaveEvent(id: number) {
  return request.post(`/annual-leave-events/${id}/restore`)
}
export function getDeletedAnnualLeaveEvents(params: any) {
  return request.get('/annual-leave-events/trash', { params })
}
export function getPersonALBalance(personId: number) {
  return request.get(`/persons/${personId}/annual-leave-balance`)
}
export function getPersonALHistory(personId: number) {
  return request.get(`/persons/${personId}/annual-leave-balance-history`)
}
export function getLILEvents(params: any) {
  return request.get('/lil-events', { params })
}
export function getPersonLILBalance(personId: number) {
  return request.get(`/persons/${personId}/lil-balance`)
}
export function getPersonLILHistory(personId: number) {
  return request.get(`/persons/${personId}/lil-balance-history`)
}
export function executeCarryover(month: string) {
  return request.post('/annual-leave-carryover', { month })
}
export function cancelCarryover(batchId: number) {
  return request.post(`/annual-leave-carryover/${batchId}/cancel`)
}
export function getCarryoverBatches() {
  return request.get('/annual-leave-carryover/batches')
}
