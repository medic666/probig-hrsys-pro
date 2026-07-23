import api from './client'

export const getAttendanceEvents = (params: any) => api.get('/attendance/events', { params })
export const createAttendanceEvent = (data: any) => api.post('/attendance/events', data)
export const updateAttendanceEvent = (id: number, data: any) => api.put(`/attendance/events/${id}`, data)
export const deleteAttendanceEvent = (id: number) => api.delete(`/attendance/events/${id}`)
export const getLeaveBalance = (personId: number) => api.get('/attendance/leave-balance', { params: { personId } })
export const grantAnnualLeave = (personId: number) => api.post('/attendance/grant-annual-leave', { personId })
export const closeMonth = (personId: number, yearMonth: string) => api.post('/attendance/close-month', { personId, yearMonth })
export const getMonthlyEvents = (personId: number, yearMonth: string) => api.get('/attendance/monthly', { params: { personId, yearMonth } })
