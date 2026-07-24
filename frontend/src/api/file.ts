import http from './request'

export function getFileList(params: any) {
  return http.get('/files', { params })
}

export function getFileDownloadUrl(id: number) {
  return `/api/files/${id}`
}

export function uploadFile(formData: FormData) {
  return http.post('/files/upload', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
}

export function deleteFile(id: number) {
  return http.delete(`/files/${id}`)
}

export function restoreFile(id: number) {
  return http.post(`/files/${id}/restore`)
}

export function getFileRelations(params: any) {
  return http.get('/files/relations', { params })
}

export function createFileRelation(data: any) {
  return http.post('/files/relations', data)
}

export function deleteFileRelation(id: number) {
  return http.delete(`/files/relations/${id}`)
}
