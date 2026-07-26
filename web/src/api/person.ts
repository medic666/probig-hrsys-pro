import request from './request'

export function getPersonList(params: any) {
  return request({ url: '/persons', method: 'get', params })
}

export function getPerson(id: number) {
  return request({ url: `/persons/${id}`, method: 'get' })
}

export function createPerson(data: any) {
  return request({ url: '/persons', method: 'post', data })
}

export function updatePerson(id: number, data: any) {
  return request({ url: `/persons/${id}`, method: 'put', data })
}

export function deletePerson(id: number) {
  return request({ url: `/persons/${id}`, method: 'delete' })
}

export function restorePerson(id: number) {
  return request({ url: `/persons/${id}/restore`, method: 'post' })
}

export function getDeletedPersons(params: any) {
  return request({ url: '/persons/deleted', method: 'get', params })
}

export function addPhone(data: { person_id: number; phone: string }) {
  return request({ url: '/persons/phone', method: 'post', data })
}

export function updatePhone(id: number, data: { phone: string }) {
  return request({ url: `/persons/phone/${id}`, method: 'put', data })
}

export function deletePhone(id: number) {
  return request({ url: `/persons/phone/${id}`, method: 'delete' })
}

export function addEmail(data: { person_id: number; email: string }) {
  return request({ url: '/persons/email', method: 'post', data })
}

export function updateEmail(id: number, data: { email: string }) {
  return request({ url: `/persons/email/${id}`, method: 'put', data })
}

export function deleteEmail(id: number) {
  return request({ url: `/persons/email/${id}`, method: 'delete' })
}

export function addBankCard(data: { person_id: number; card_no: string; bank_name: string }) {
  return request({ url: '/persons/bank-card', method: 'post', data })
}

export function updateBankCard(id: number, data: { card_no: string; bank_name: string }) {
  return request({ url: `/persons/bank-card/${id}`, method: 'put', data })
}

export function deleteBankCard(id: number) {
  return request({ url: `/persons/bank-card/${id}`, method: 'delete' })
}
