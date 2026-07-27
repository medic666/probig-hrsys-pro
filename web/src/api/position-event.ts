import request from '@/utils/request'

export function getPositionEvents(params: any) {
  return request.get('/position-events', { params })
}

export function getPositionEvent(id: number) {
  return request.get(`/position-events/${id}`)
}

export function createPositionEvent(data: any) {
  const body: any = {}
  for (const key of Object.keys(data)) {
    if (data[key] !== null && data[key] !== undefined && data[key] !== '') {
      body[key] = data[key]
    }
  }
  return request.post('/position-events', body)
}

export function updatePositionEvent(id: number, data: any) {
  return request.put(`/position-events/${id}`, data)
}

export function deletePositionEvent(id: number) {
  return request.delete(`/position-events/${id}`)
}

export function restorePositionEvent(id: number) {
  return request.post(`/position-events/${id}/restore`)
}

export function getDeletedPositionEvents(params: any) {
  return request.get('/position-events/trash', { params })
}
