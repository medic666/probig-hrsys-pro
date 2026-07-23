import { useState, useEffect } from 'react'
import { Table, Button, Space, Select, Input, message, Popconfirm, Tag, Modal } from 'antd'
import { SearchOutlined, ReloadOutlined, DeleteOutlined, EditOutlined } from '@ant-design/icons'
import { getEvents, updateEventRemark, deleteEvent } from '../api/events'
import dayjs from 'dayjs'

const entityTypes = [
  { label: '人员', value: 'person' },
  { label: '制度', value: 'policy' },
  { label: '考勤', value: 'attendance' },
  { label: '工资', value: 'salary' },
  { label: '资产', value: 'asset' },
]

const eventTypeColors: Record<string, string> = {
  create: 'green', update: 'blue', delete: 'red', grant: 'purple', calculate: 'orange', carry_over: 'cyan', close_month: 'gold',
}

export default function Events() {
  const [data, setData] = useState<any[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(false)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const [entityType, setEntityType] = useState('')
  const [eventType, setEventType] = useState('')
  const [remarkModalOpen, setRemarkModalOpen] = useState(false)
  const [remarkId, setRemarkId] = useState<number>(0)
  const [remark, setRemark] = useState('')

  const fetchData = async () => {
    setLoading(true)
    try {
      const params: any = { page, pageSize }
      if (entityType) params.entityType = entityType
      if (eventType) params.eventType = eventType
      const res: any = await getEvents(params)
      setData(res.data.list || [])
      setTotal(res.data.total || 0)
    } finally { setLoading(false) }
  }

  useEffect(() => { fetchData() }, [page, pageSize])

  const handleDelete = async (id: number) => { await deleteEvent(id); message.success('已软删除'); fetchData() }

  const handleEditRemark = (id: number, currentRemark: string) => {
    setRemarkId(id)
    setRemark(currentRemark)
    setRemarkModalOpen(true)
  }

  const handleSaveRemark = async () => {
    try {
      await updateEventRemark(remarkId, remark)
      message.success('备注已更新')
      setRemarkModalOpen(false)
      fetchData()
    } catch (e: any) { message.error(e.message) }
  }

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 60 },
    { title: '实体类型', dataIndex: 'entityType', width: 100, render: (v: string) => <Tag>{entityTypes.find(t => t.value === v)?.label || v}</Tag> },
    { title: '实体ID', dataIndex: 'entityId', width: 70 },
    { title: '事件类型', dataIndex: 'eventType', width: 100, render: (v: string) => <Tag color={eventTypeColors[v] || 'default'}>{v}</Tag> },
    { title: '操作人', dataIndex: 'operatorId', width: 70 },
    { title: '备注', dataIndex: 'remark', width: 150, ellipsis: true },
    { title: '时间', dataIndex: 'createdAt', width: 160, render: (v: string) => dayjs(v).format('YYYY-MM-DD HH:mm:ss') },
    {
      title: '操作', key: 'actions', width: 160,
      render: (_: any, record: any) => (
        <Space>
          <Button size="small" icon={<EditOutlined />} onClick={() => handleEditRemark(record.id, record.remark)}>备注</Button>
          <Popconfirm title="确定软删除?" onConfirm={() => handleDelete(record.id)}>
            <Button size="small" danger icon={<DeleteOutlined />}>删除</Button>
          </Popconfirm>
        </Space>
      ),
    },
  ]

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
        <Space>
          <Select placeholder="实体类型" value={entityType || undefined} onChange={(v) => { setEntityType(v || ''); setPage(1); }} allowClear style={{ width: 140 }} options={entityTypes} />
          <Select placeholder="事件类型" value={eventType || undefined} onChange={(v) => { setEventType(v || ''); setPage(1); }} allowClear style={{ width: 140 }}
            options={['create', 'update', 'delete', 'grant', 'calculate', 'carry_over', 'close_month'].map(t => ({ label: t, value: t }))} />
          <Button type="primary" icon={<SearchOutlined />} onClick={() => { setPage(1); fetchData() }}>搜索</Button>
          <Button icon={<ReloadOutlined />} onClick={() => { setEntityType(''); setEventType(''); setPage(1); setTimeout(fetchData, 0) }}>刷新</Button>
        </Space>
      </div>

      <Table rowKey="id" columns={columns} dataSource={data} loading={loading}
        pagination={{ current: page, pageSize, total, showSizeChanger: true, onChange: (p, ps) => { setPage(p); setPageSize(ps) } }} />

      <Modal title="修改备注" open={remarkModalOpen} onCancel={() => setRemarkModalOpen(false)} onOk={handleSaveRemark}>
        <Input.TextArea rows={3} value={remark} onChange={(e) => setRemark(e.target.value)} />
      </Modal>
    </div>
  )
}
