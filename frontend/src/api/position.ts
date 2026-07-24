import http from './request'

export function getPositionEvents(params: any) {
  return http.get('/position-events', { params })
}

export function getPositionEvent(id: number) {
  return http.get(`/position-events/${id}`)
}

export function createPositionEvent(data: any) {
  return http.post('/position-events', data)
}

export function updatePositionEvent(id: number, data: any) {
  return http.put(`/position-events/${id}`, data)
}

export function deletePositionEvent(id: number) {
  return http.delete(`/position-events/${id}`)
}

export function getPositionSnapshots(params: any) {
  return http.get('/position-snapshots', { params })
}

export function rebuildSnapshots(data: { person_id: number }) {
  return http.post('/position-snapshots/rebuild', null, { params: data })
}
