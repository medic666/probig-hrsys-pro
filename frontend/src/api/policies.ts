import api from './client'

export const getPolicies = (params: any) => api.get('/policies', { params })
export const getPolicy = (id: number) => api.get(`/policies/${id}`)
export const createPolicy = (data: any) => api.post('/policies', data)
export const updatePolicy = (id: number, data: any) => api.put(`/policies/${id}`, data)
export const deletePolicy = (id: number) => api.delete(`/policies/${id}`)
export const getPolicyVersions = (id: number) => api.get(`/policies/${id}/versions`)
export const getPolicyTimeline = (id: number) => api.get(`/policies/${id}/timeline`)
