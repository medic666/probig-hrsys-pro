export function downloadFile(blob: Blob, fileName: string) {
  const url = window.URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = fileName
  link.click()
  window.URL.revokeObjectURL(url)
}

export function exportCSV(rows: any[][], headers: string[], fileName: string) {
  const content = [headers.join(','), ...rows.map(r => r.join(','))].join('\n')
  const blob = new Blob(['\uFEFF' + content], { type: 'text/csv;charset=utf-8;' })
  downloadFile(blob, fileName)
}
