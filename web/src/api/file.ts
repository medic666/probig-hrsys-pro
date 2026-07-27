import request from '@/utils/request'

export function uploadFile(file: File) {
  const formData = new FormData()
  formData.append('file', file)
  return request.post('/files/upload', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
}

export function getFilesByTarget(target_type: string, target_id: number) {
  return request.get('/files/by-target', { params: { target_type, target_id } })
}

export function associateFile(file_id: number, target_type: string, target_id: number) {
  return request.post('/files/associate', { file_id, target_type, target_id })
}

export function disassociateFile(id: number) {
  return request.post('/files/disassociate', { id })
}

export function downloadFile(fileId: number) {
  return request.get(`/files/${fileId}/download`, { responseType: 'blob' })
}
