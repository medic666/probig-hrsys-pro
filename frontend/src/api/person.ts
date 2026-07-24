import http from './request'

export function getPersonList(params: any) {
  return http.get('/persons', { params })
}

export function getPerson(id: number) {
  return http.get(`/persons/${id}`)
}

export function createPerson(data: any) {
  return http.post('/persons', data)
}

export function updatePerson(id: number, data: any) {
  return http.put(`/persons/${id}`, data)
}

export function deletePerson(id: number) {
  return http.delete(`/persons/${id}`)
}

export function restorePerson(id: number) {
  return http.post(`/persons/${id}/restore`)
}
