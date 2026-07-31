import request from '@/utils/request'
import { downloadBlob, type BlobResult } from '@/utils/download'

export function uploadFile(file: File) {
  const formData = new FormData()
  formData.append('file', file)
  return request.post('/files/upload', formData, { headers: { 'Content-Type': 'multipart/form-data' } })
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
export function downloadFile(fileId: number, fileName?: string) {
  return request.get(`/files/${fileId}/download`, { responseType: 'blob' }).then((res) => {
    downloadBlob(res as unknown as BlobResult, fileName || 'download')
  })
}
export function getFiles(params: any) { return request.get('/files', { params }) }
export function deleteFile(id: number) { return request.delete(`/files/${id}`) }
export function restoreFile(id: number) { return request.post(`/files/${id}/restore`) }
export function getDeletedFiles(params: any) { return request.get('/files/trash', { params }) }
export function getFileAssociations(id: number) { return request.get(`/files/${id}/associations`) }
export function permanentDeleteFile(id: number) { return request.delete(`/files/${id}/permanent`) }
export function cleanOrphanFiles() { return request.post('/files/clean-orphans') }
