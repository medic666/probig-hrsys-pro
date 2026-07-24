import { get, del, upload as uploadFile } from './client'
import type { PageData, FileItem } from '../types'

export function listFiles(params?: { page?: number; page_size?: number }) {
  return get<PageData<FileItem>>('/files', params)
}

export function upload(formData: FormData) {
  return uploadFile<FileItem>('/files/upload', formData)
}

export function deleteFile(id: number) {
  return del(`/files/${id}`)
}

export function associateFile(fileId: number, data: { target_type: string; target_id: number }) {
  return uploadFile(`/files/${fileId}/associate`, data as any) // sic: using POST
}
