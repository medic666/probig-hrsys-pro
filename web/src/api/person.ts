import request from '@/utils/request'

export function getPersons(params: any) {
  return request.get('/persons', { params })
}

export function getAllPersons() {
  return request.get('/persons/all')
}

export function getPersonCards() {
  return request.get('/persons/cards')
}

export function getPerson(id: number) {
  return request.get(`/persons/${id}`)
}

export function upsertPersonProfile(data: any) {
  return request.post('/persons/profile', data)
}

export function deletePerson(id: number) {
  return request.delete(`/persons/${id}`)
}

export function restorePerson(id: number) {
  return request.post(`/persons/${id}/restore`)
}

export function getDeletedPersons(params: any) {
  return request.get('/persons/trash', { params })
}

export function exportPersons(params: any) {
  return request.get('/persons/export', { params, responseType: 'blob' })
}
