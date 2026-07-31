export interface BlobResult {
  blob: Blob
  filename: string
}

export function downloadBlob(result: BlobResult | Blob, fallbackName = 'download.xlsx') {
  const blob = result instanceof Blob ? result : result.blob
  const filename = result instanceof Blob ? fallbackName : result.filename || fallbackName
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  a.click()
  URL.revokeObjectURL(url)
}
