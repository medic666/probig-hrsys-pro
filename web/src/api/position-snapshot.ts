import request from '@/utils/request'

export function getCurrentPosition(personId: number) {
  return request.get(`/persons/${personId}/current-position`)
}

export function getPositionHistory(personId: number) {
  return request.get(`/persons/${personId}/position-history`)
}
