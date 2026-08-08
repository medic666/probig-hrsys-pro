import request from '@/utils/request'

// ============ 结构授权点（跨模块「人员×时间」主轴底座） ============
// 后端对应 middleware.StructureOnly：任何登录用户可用，不参与模块×动作权限判断。
// 业务数据请保持在各模块 API 文件（需模块×动作权限）。

export interface PersonOption {
  id: number
  name: string
  company_id?: number
  company_name?: string
  attendance_group?: string
  is_active?: boolean
  entry_date?: string | null
  leave_date?: string | null
}

export function getPersonOptions(): Promise<PersonOption[]> {
  return request.get('/persons/all') as Promise<PersonOption[]>
}

// searchPersonOptions 人员名称模糊搜索：人员选择组件（NameSelect/ProTable person-select）统一默认数据源
export async function searchPersonOptions(keyword?: string): Promise<PersonOption[]> {
  const list = (await getPersonOptions()) || []
  if (!keyword) return list
  return list.filter((p) => p.name.includes(keyword))
}

export function getPersonCards() {
  return request.get('/persons/cards')
}

export function getCompanyOptions(): Promise<{ id: number; name: string }[]> {
  return request.get('/companies/all') as Promise<{ id: number; name: string }[]>
}
