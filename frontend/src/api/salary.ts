import api from './client'

export const getSalaryRecords = (params: any) => api.get('/salary', { params })
export const calculateSalary = (personId: number, yearMonth: string) => api.post('/salary/calculate', { personId, yearMonth })
export const getSalaryRecord = (personId: number, yearMonth: string) => api.get('/salary/record', { params: { personId, yearMonth } })
export const addAdjustment = (data: any) => api.post('/salary/adjustments', data)
export const deleteAdjustment = (id: number) => api.delete(`/salary/adjustments/${id}`)
export const getAdjustments = (personId: number, yearMonth: string) => api.get('/salary/adjustments', { params: { personId, yearMonth } })
export const getSalaryByMonth = (yearMonth: string) => api.get('/salary/by-month', { params: { yearMonth } })
