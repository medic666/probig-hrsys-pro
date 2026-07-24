import http from './request'

export function getCompanyList(params: any) {
  return http.get('/companies', { params })
}

export function getCompany(id: number) {
  return http.get(`/companies/${id}`)
}

export function createCompany(data: any) {
  return http.post('/companies', data)
}

export function updateCompany(id: number, data: any) {
  return http.put(`/companies/${id}`, data)
}

export function deleteCompany(id: number) {
  return http.delete(`/companies/${id}`)
}

export function restoreCompany(id: number) {
  return http.post(`/companies/${id}/restore`)
}
