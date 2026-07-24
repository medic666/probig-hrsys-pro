import { get, post, put, del } from './client'
import type { PageData, PersonnelSnapshot, PersonnelEvent } from '../types'

export function listPersonnel(params?: { search?: string; page?: number; page_size?: number }) {
  return get<PageData<PersonnelSnapshot>>('/personnel', params)
}

export function getPersonnel(id: number) {
  return get<PersonnelSnapshot>(`/personnel/${id}`)
}

export function getPersonnelHistory(id: number) {
  return get<PersonnelSnapshot[]>(`/personnel/${id}/history`)
}

export function listPersonnelEvents(params?: { entity_id?: number; page?: number; page_size?: number }) {
  return get<PageData<PersonnelEvent>>('/personnel/events', params)
}

export function createPersonnelEvent(data: any) {
  return post<PersonnelEvent>('/personnel/events', data)
}

export function updatePersonnelEvent(id: number, data: any) {
  return put<PersonnelEvent>(`/personnel/events/${id}`, data)
}

export function deletePersonnelEvent(id: number) {
  return del(`/personnel/events/${id}`)
}
