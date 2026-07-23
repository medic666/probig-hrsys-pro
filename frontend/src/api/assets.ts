import api from './client'

export const getAssets = (params: any) => api.get('/assets', { params })
export const getAsset = (id: number) => api.get(`/assets/${id}`)
export const createAsset = (data: any) => api.post('/assets', data)
export const updateAsset = (id: number, data: any) => api.put(`/assets/${id}`, data)
export const deleteAsset = (id: number) => api.delete(`/assets/${id}`)
export const getAssetVersions = (id: number) => api.get(`/assets/${id}/versions`)
export const getAssetTimeline = (id: number) => api.get(`/assets/${id}/timeline`)
