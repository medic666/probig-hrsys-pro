import request from './request'

export function getCompanyList(params: any) {
  return request({ url: '/companies', method: 'get', params })
}

export function getCompany(id: number) {
  return request({ url: `/companies/${id}`, method: 'get' })
}

export function createCompany(data: any) {
  return request({ url: '/companies', method: 'post', data })
}

export function updateCompany(id: number, data: any) {
  return request({ url: `/companies/${id}`, method: 'put', data })
}

export function deleteCompany(id: number) {
  return request({ url: `/companies/${id}`, method: 'delete' })
}

export function restoreCompany(id: number) {
  return request({ url: `/companies/${id}/restore`, method: 'post' })
}

export function getDeletedCompanies(params: any) {
  return request({ url: '/companies/deleted', method: 'get', params })
}
