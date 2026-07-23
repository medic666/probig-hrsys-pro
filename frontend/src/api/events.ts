import api from './client'

export const getEvents = (params: any) => api.get('/events', { params })
export const getEntityEvents = (entityType: string, entityId: number) => api.get(`/events/entity/${entityType}/${entityId}`)
export const updateEventRemark = (id: number, remark: string) => api.put(`/events/${id}/remark`, { remark })
export const deleteEvent = (id: number) => api.delete(`/events/${id}`)
