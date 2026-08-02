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

export function createPerson(data: any) {
  return request.post('/persons', data)
}

export function updatePerson(id: number, data: any) {
  return request.put(`/persons/${id}`, data)
}

export function updatePersonProfile(id: number, data: any) {
  return request.put(`/persons/${id}/profile`, data)
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

export function addPersonPhone(personId: number, data: any) {
  return request.post(`/persons/${personId}/phones`, data)
}

export function updatePersonPhone(phoneId: number, data: any) {
  return request.put(`/persons/0/phones/${phoneId}`, data)
}

export function deletePersonPhone(phoneId: number) {
  return request.delete(`/persons/0/phones/${phoneId}`)
}

export function addPersonEmail(personId: number, data: any) {
  return request.post(`/persons/${personId}/emails`, data)
}

export function updatePersonEmail(emailId: number, data: any) {
  return request.put(`/persons/0/emails/${emailId}`, data)
}

export function deletePersonEmail(emailId: number) {
  return request.delete(`/persons/0/emails/${emailId}`)
}

export function addPersonBankCard(personId: number, data: any) {
  return request.post(`/persons/${personId}/bank-cards`, data)
}

export function updatePersonBankCard(cardId: number, data: any) {
  return request.put(`/persons/0/bank-cards/${cardId}`, data)
}

export function deletePersonBankCard(cardId: number) {
  return request.delete(`/persons/0/bank-cards/${cardId}`)
}

export function addPersonEmergencyContact(personId: number, data: any) {
  return request.post(`/persons/${personId}/emergency-contacts`, data)
}

export function updatePersonEmergencyContact(contactId: number, data: any) {
  return request.put(`/persons/0/emergency-contacts/${contactId}`, data)
}

export function deletePersonEmergencyContact(contactId: number) {
  return request.delete(`/persons/0/emergency-contacts/${contactId}`)
}
