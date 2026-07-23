import { useState, useEffect } from 'react'
import { Table, Button, Space, Input, Drawer, Form, Select, message, Popconfirm, Modal, Tag } from 'antd'
import { PlusOutlined, SearchOutlined, ReloadOutlined, EditOutlined, DeleteOutlined, HistoryOutlined } from '@ant-design/icons'
import { getAssets, createAsset, updateAsset, deleteAsset, getAssetVersions, getAssetTimeline } from '../api/assets'
import EventTimeline from '../components/EventTimeline'
import dayjs from 'dayjs'

const assetTypes = [
  { label: '公司资料', value: 'company_info' },
  { label: '实体物产', value: 'physical' },
  { label: '虚拟物产', value: 'virtual' },
]

export default function Assets() {
  const [data, setData] = useState<any[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(false)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const [search, setSearch] = useState('')
  const [assetType, setAssetType] = useState('')
  const [drawerOpen, setDrawerOpen] = useState(false)
  const [editingRecord, setEditingRecord] = useState<any>(null)
  const [versionsOpen, setVersionsOpen] = useState(false)
  const [versions, setVersions] = useState<any[]>([])
  const [timelineOpen, setTimelineOpen] = useState(false)
  const [timelineEvents, setTimelineEvents] = useState<any[]>([])
  const [form] = Form.useForm()

  const fetchData = async () => {
    setLoading(true)
    try {
      const res: any = await getAssets({ page, pageSize, search, assetType })
      setData(res.data.list || [])
      setTotal(res.data.total || 0)
    } finally { setLoading(false) }
  }

  useEffect(() => { fetchData() }, [page, pageSize])

  const handleCreate = () => { setEditingRecord(null); form.resetFields(); setDrawerOpen(true) }

  const handleEdit = (record: any) => {
    setEditingRecord(record)
    form.setFieldsValue(record)
    setDrawerOpen(true)
  }

  const handleSave = async () => {
    const values = await form.validateFields()
    try {
      if (editingRecord) {
        await updateAsset(editingRecord.id, values)
      } else {
        await createAsset(values)
      }
      message.success('保存成功')
      setDrawerOpen(false)
      fetchData()
    } catch (e: any) { message.error(e.message || '保存失败') }
  }

  const handleDelete = async (id: number) => { await deleteAsset(id); message.success('删除成功'); fetchData() }

  const handleVersions = async (id: number) => {
    const res: any = await getAssetVersions(id)
    setVersions(res.data || [])
    setVersionsOpen(true)
  }

  const handleTimeline = async (id: number) => {
    const res: any = await getAssetTimeline(id)
    setTimelineEvents(res.data || [])
    setTimelineOpen(true)
  }

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 60 },
    { title: '名称', dataIndex: 'name', width: 200 },
    { title: '类型', dataIndex: 'assetType', width: 100, render: (v: string) => <Tag>{assetTypes.find(t => t.value === v)?.label || v}</Tag> },
    { title: '版本', dataIndex: 'version', width: 60 },
    { title: '描述', dataIndex: 'description', ellipsis: true },
    { title: '状态', dataIndex: 'status', width: 80, render: (v: number) => v === 1 ? <Tag color="green">正常</Tag> : <Tag color="red">已删除</Tag> },
    { title: '创建时间', dataIndex: 'createdAt', width: 160, render: (v: string) => dayjs(v).format('YYYY-MM-DD HH:mm') },
    {
      title: '操作', key: 'actions', width: 280,
      render: (_: any, record: any) => (
        <Space>
          <Button size="small" icon={<EditOutlined />} onClick={() => handleEdit(record)}>编辑</Button>
          <Button size="small" onClick={() => handleVersions(record.id)}>版本</Button>
          <Button size="small" icon={<HistoryOutlined />} onClick={() => handleTimeline(record.id)}>历史</Button>
          <Popconfirm title="确定删除?" onConfirm={() => handleDelete(record.id)}>
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
          <Input placeholder="搜索名称/描述" value={search} onChange={(e) => setSearch(e.target.value)} onPressEnter={() => { setPage(1); fetchData() }} style={{ width: 250 }} prefix={<SearchOutlined />} />
          <Select placeholder="资产类型" value={assetType || undefined} onChange={(v) => { setAssetType(v || ''); setPage(1); setTimeout(fetchData, 0) }} allowClear style={{ width: 140 }} options={assetTypes} />
          <Button type="primary" onClick={() => { setPage(1); fetchData() }}>搜索</Button>
          <Button icon={<ReloadOutlined />} onClick={() => { setSearch(''); setAssetType(''); setPage(1); setTimeout(fetchData, 0) }}>刷新</Button>
        </Space>
        <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>新增资产</Button>
      </div>

      <Table rowKey="id" columns={columns} dataSource={data} loading={loading}
        pagination={{ current: page, pageSize, total, showSizeChanger: true, onChange: (p, ps) => { setPage(p); setPageSize(ps) } }} />

      <Drawer title={editingRecord ? '编辑资产' : '新增资产'} open={drawerOpen} onClose={() => setDrawerOpen(false)} width={600}
        extra={<Button type="primary" onClick={handleSave}>保存</Button>}>
        <Form form={form} layout="vertical">
          <Form.Item name="name" label="名称" rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="assetType" label="类型" rules={[{ required: true }]}>
            <Select options={assetTypes} />
          </Form.Item>
          <Form.Item name="description" label="描述"><Input.TextArea rows={3} /></Form.Item>
          <Form.Item name="content" label="内容(JSON)"><Input.TextArea rows={6} placeholder='{"key": "value"}' /></Form.Item>
        </Form>
      </Drawer>

      <Modal title="版本历史" open={versionsOpen} onCancel={() => setVersionsOpen(false)} footer={null} width={700}>
        <Table rowKey="id" dataSource={versions} pagination={false} size="small"
          columns={[
            { title: '版本', dataIndex: 'version', width: 60 },
            { title: '名称', dataIndex: 'name' },
            { title: '当前', dataIndex: 'isCurrent', width: 80, render: (v: number) => v ? <Tag color="green">当前</Tag> : <Tag>历史</Tag> },
            { title: '时间', dataIndex: 'createdAt', width: 160, render: (v: string) => dayjs(v).format('YYYY-MM-DD HH:mm') },
          ]} />
      </Modal>

      <Modal title="事件时间线" open={timelineOpen} onCancel={() => setTimelineOpen(false)} footer={null} width={600}>
        <EventTimeline events={timelineEvents} />
      </Modal>
    </div>
  )
}
