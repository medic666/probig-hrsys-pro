import api from './client'

export const getPersons = (params: any) => api.get('/persons', { params })
export const getAllPersons = () => api.get('/persons/all')
export const getPerson = (id: number) => api.get(`/persons/${id}`)
export const createPerson = (data: any) => api.post('/persons', data)
export const updatePerson = (id: number, data: any) => api.put(`/persons/${id}`, data)
export const deletePerson = (id: number) => api.delete(`/persons/${id}`)
export const getPersonTimeline = (id: number) => api.get(`/persons/${id}/timeline`)
export const getPersonSnapshot = (id: number, at: string) => api.get(`/persons/${id}/snapshot`, { params: { at } })
