import request from './request'

export function getFileList(params: any) {
  return request({ url: '/files', method: 'get', params })
}

export function uploadFile(formData: FormData) {
  return request({ url: '/files/upload', method: 'post', data: formData, headers: { 'Content-Type': 'multipart/form-data' } })
}

export function deleteFile(id: number) {
  return request({ url: `/files/${id}`, method: 'delete' })
}

export function restoreFile(id: number) {
  return request({ url: `/files/${id}/restore`, method: 'post' })
}

export function getFileRelations(id: number) {
  return request({ url: `/files/${id}/relations`, method: 'get' })
}

export function addFileRelation(id: number, data: { target_type: string; target_id: number }) {
  return request({ url: `/files/${id}/relations`, method: 'post', data })
}

export function getTargetFiles(targetType: string, targetId: number) {
  return request({ url: `/files/target/${targetType}/${targetId}`, method: 'get' })
}
