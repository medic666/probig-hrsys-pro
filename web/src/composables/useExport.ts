import { ref } from 'vue'
import { downloadBlob, type BlobResult } from '@/utils/download'

// useExport 列表导出统一封装：exportFn 为导出接口（返回 blob），getParams 在点击时
// 从列表视图取当前筛选（ProTable.getSearchParams）。返回 { exporting, run }，
// run 直接绑定按钮 @click（事件参数不参与调用）。
export function useExport(
  exportFn: (params: any) => Promise<BlobResult | Blob>,
  getParams: () => any,
) {
  const exporting = ref(false)

  async function run() {
    if (exporting.value) return
    exporting.value = true
    try {
      const data = await exportFn(getParams() || {})
      downloadBlob(data)
    } catch {
      /* error handled by request interceptor */
    } finally {
      exporting.value = false
    }
  }

  return { exporting, run }
}
