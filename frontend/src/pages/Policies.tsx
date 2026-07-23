import { useState, useEffect } from 'react'
import { Table, Button, Space, Input, Drawer, Form, Select, message, Popconfirm, Modal, Tag } from 'antd'
import { PlusOutlined, SearchOutlined, ReloadOutlined, EditOutlined, DeleteOutlined, HistoryOutlined } from '@ant-design/icons'
import { getPolicies, createPolicy, updatePolicy, deletePolicy, getPolicyVersions, getPolicyTimeline } from '../api/policies'
import EventTimeline from '../components/EventTimeline'
import dayjs from 'dayjs'

const policyTypes = [
  { label: '考勤制度', value: 'attendance' },
  { label: '工资制度', value: 'salary' },
  { label: '年假制度', value: 'annual_leave' },
  { label: '行政制度', value: 'admin' },
  { label: '其他', value: 'other' },
]

export default function Policies() {
  const [data, setData] = useState<any[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(false)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const [search, setSearch] = useState('')
  const [policyType, setPolicyType] = useState('')
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
      const res: any = await getPolicies({ page, pageSize, search, policyType })
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
        await updatePolicy(editingRecord.id, values)
      } else {
        await createPolicy(values)
      }
      message.success('保存成功')
      setDrawerOpen(false)
      fetchData()
    } catch (e: any) { message.error(e.message || '保存失败') }
  }

  const handleDelete = async (id: number) => { await deletePolicy(id); message.success('删除成功'); fetchData() }

  const handleVersions = async (id: number) => {
    const res: any = await getPolicyVersions(id)
    setVersions(res.data || [])
    setVersionsOpen(true)
  }

  const handleTimeline = async (id: number) => {
    const res: any = await getPolicyTimeline(id)
    setTimelineEvents(res.data || [])
    setTimelineOpen(true)
  }

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 60 },
    { title: '标题', dataIndex: 'title', width: 200 },
    { title: '类型', dataIndex: 'policyType', width: 100, render: (v: string) => <Tag>{policyTypes.find(t => t.value === v)?.label || v}</Tag> },
    { title: '版本', dataIndex: 'version', width: 60 },
    { title: '内容', dataIndex: 'content', ellipsis: true, render: (v: string) => v?.substring(0, 100) },
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

  const versionColumns = [
    { title: '版本', dataIndex: 'version', width: 60 },
    { title: '标题', dataIndex: 'title' },
    { title: '状态', dataIndex: 'isCurrent', width: 80, render: (v: number) => v ? <Tag color="green">当前</Tag> : <Tag>历史</Tag> },
    { title: '时间', dataIndex: 'createdAt', width: 160, render: (v: string) => dayjs(v).format('YYYY-MM-DD HH:mm') },
  ]

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
        <Space>
          <Input placeholder="搜索标题/内容" value={search} onChange={(e) => setSearch(e.target.value)} onPressEnter={() => { setPage(1); fetchData() }} style={{ width: 250 }} prefix={<SearchOutlined />} />
          <Select placeholder="制度类型" value={policyType || undefined} onChange={(v) => { setPolicyType(v || ''); setPage(1); setTimeout(fetchData, 0) }} allowClear style={{ width: 140 }} options={policyTypes} />
          <Button type="primary" onClick={() => { setPage(1); fetchData() }}>搜索</Button>
          <Button icon={<ReloadOutlined />} onClick={() => { setSearch(''); setPolicyType(''); setPage(1); setTimeout(fetchData, 0) }}>刷新</Button>
        </Space>
        <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>新增制度</Button>
      </div>

      <Table rowKey="id" columns={columns} dataSource={data} loading={loading}
        pagination={{ current: page, pageSize, total, showSizeChanger: true, onChange: (p, ps) => { setPage(p); setPageSize(ps) } }} />

      <Drawer title={editingRecord ? '编辑制度' : '新增制度'} open={drawerOpen} onClose={() => setDrawerOpen(false)} width={600}
        extra={<Button type="primary" onClick={handleSave}>保存</Button>}>
        <Form form={form} layout="vertical">
          <Form.Item name="title" label="标题" rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="policyType" label="类型" rules={[{ required: true }]}>
            <Select options={policyTypes} />
          </Form.Item>
          <Form.Item name="content" label="内容" rules={[{ required: true }]}><Input.TextArea rows={10} /></Form.Item>
        </Form>
      </Drawer>

      <Modal title="版本历史" open={versionsOpen} onCancel={() => setVersionsOpen(false)} footer={null} width={700}>
        <Table rowKey="id" columns={versionColumns} dataSource={versions} pagination={false} size="small" />
      </Modal>

      <Modal title="事件时间线" open={timelineOpen} onCancel={() => setTimelineOpen(false)} footer={null} width={600}>
        <EventTimeline events={timelineEvents} />
      </Modal>
    </div>
  )
}
