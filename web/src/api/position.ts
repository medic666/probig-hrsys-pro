import request from './request'

export function getPositionEvents(params: any) {
  return request({ url: '/positions/events', method: 'get', params })
}

export function createPositionEvent(data: any) {
  return request({ url: '/positions/events', method: 'post', data })
}

export function updatePositionEvent(id: number, data: any) {
  return request({ url: `/positions/events/${id}`, method: 'put', data })
}

export function deletePositionEvent(id: number) {
  return request({ url: `/positions/events/${id}`, method: 'delete' })
}

export function getSnapshots(personId: number) {
  return request({ url: `/positions/snapshots/${personId}`, method: 'get' })
}

export function getCurrentSnapshot(personId: number) {
  return request({ url: `/positions/current-snapshot/${personId}`, method: 'get' })
}
