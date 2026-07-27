import request from '@/utils/request'

export function getCompanies(params: any) {
  return request.get('/companies', { params })
}

export function getAllCompanies() {
  return request.get('/companies/all')
}

export function getCompany(id: number) {
  return request.get(`/companies/${id}`)
}

export function createCompany(data: any) {
  return request.post('/companies', data)
}

export function updateCompany(id: number, data: any) {
  return request.put(`/companies/${id}`, data)
}

export function deleteCompany(id: number) {
  return request.delete(`/companies/${id}`)
}

export function restoreCompany(id: number) {
  return request.post(`/companies/${id}/restore`)
}

export function getDeletedCompanies(params: any) {
  return request.get('/companies/trash', { params })
}

export function exportCompanies(params: any) {
  return request.get('/companies/export', { params, responseType: 'blob' })
}
