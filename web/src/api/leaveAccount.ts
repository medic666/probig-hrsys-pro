import request from './request'

export function getLeaveEvents(params: any) {
  return request({ url: '/leave-accounts/events', method: 'get', params })
}

export function createLeaveEvent(data: any) {
  return request({ url: '/leave-accounts/events', method: 'post', data })
}

export function deleteLeaveEvent(id: number) {
  return request({ url: `/leave-accounts/events/${id}`, method: 'delete' })
}

export function getBalanceList(params: any) {
  return request({ url: '/leave-accounts/balances', method: 'get', params })
}

export function getBalanceDetail(personId: number, leaveType: string) {
  return request({ url: `/leave-accounts/balances/detail/${personId}/${leaveType}`, method: 'get' })
}

export function getBatches(params: any) {
  return request({ url: '/leave-accounts/carryover/batches', method: 'get', params })
}

export function getBatchEvents(batchId: number) {
  return request({ url: `/leave-accounts/carryover/batches/${batchId}/events`, method: 'get' })
}

export function executeCarryover(data: { target_month: string }) {
  return request({ url: '/leave-accounts/carryover/execute', method: 'post', data })
}

export function cancelCarryover(id: number) {
  return request({ url: `/leave-accounts/carryover/cancel/${id}`, method: 'post' })
}
