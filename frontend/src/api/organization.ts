import { get, post, put, del } from './client'
import type { PageData, OrganizationSnapshot, OrganizationEvent } from '../types'

export function listOrganizations(params?: { search?: string; page?: number; page_size?: number }) {
  return get<PageData<OrganizationSnapshot>>('/organizations', params)
}

export function getOrganization(id: number) {
  return get<OrganizationSnapshot>(`/organizations/${id}`)
}

export function getOrganizationHistory(id: number) {
  return get<OrganizationSnapshot[]>(`/organizations/${id}/history`)
}

export function listOrganizationEvents(params?: { entity_id?: number; page?: number; page_size?: number }) {
  return get<PageData<OrganizationEvent>>('/organizations/events', params)
}

export function createOrganizationEvent(data: any) {
  return post<OrganizationEvent>('/organizations/events', data)
}

export function updateOrganizationEvent(id: number, data: any) {
  return put<OrganizationEvent>(`/organizations/events/${id}`, data)
}

export function deleteOrganizationEvent(id: number) {
  return del(`/organizations/events/${id}`)
}
