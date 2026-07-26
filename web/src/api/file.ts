import { get, post, put, del } from './request'

export interface FileRecord {
  id: number
  file_name: string
  file_type: string
  file_size: number
  file_path: string
  created_at: string
  updated_at: string
}

export interface FileListParams {
  pageNum: number
  pageSize: number
  file_name?: string
  file_type?: string
  upload_time_start?: string
  upload_time_end?: string
}

export interface FileListResponse {
  list: FileRecord[]
  total: number
}

export interface FileRelations {
  file_id: number
  relations: { target_type: string; target_id: number; target_name: string }[]
}

export interface UpdateRelationsRequest {
  file_id: number
  relations: { target_type: string; target_id: number }[]
}

export function listFiles(params: FileListParams): Promise<FileListResponse> {
  return get<FileListResponse>('/api/files', params as unknown as Record<string, unknown>)
}

export function uploadFile(formData: FormData): Promise<FileRecord> {
  return post<FileRecord>('/api/files/upload', formData, {
    headers: { 'Content-Type': 'multipart/form-data' }
  })
}

export function deleteFile(id: number): Promise<void> {
  return del<void>('/api/files', { id })
}

export function restoreFile(id: number): Promise<void> {
  return post<void>('/api/files/restore', { id })
}

export function updateFileRelations(data: UpdateRelationsRequest): Promise<void> {
  return put<void>('/api/files/relations', data)
}

export function getFileRelations(fileId: number): Promise<FileRelations> {
  return get<FileRelations>('/api/files/relations', { file_id: fileId })
}

export function downloadFile(id: number): Promise<Blob> {
  return get<Blob>('/api/files/download', { id }, { responseType: 'blob' })
}
