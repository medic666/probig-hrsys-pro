import { useState, useEffect } from 'react'
import { Table, Button, Space, Input, Drawer, Form, Select, DatePicker, InputNumber, message, Popconfirm, Card, Statistic } from 'antd'
import { PlusOutlined, SearchOutlined, ReloadOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons'
import { getAttendanceEvents, createAttendanceEvent, updateAttendanceEvent, deleteAttendanceEvent, getLeaveBalance, grantAnnualLeave, closeMonth } from '../api/attendance'
import { getAllPersons } from '../api/persons'
import dayjs from 'dayjs'

const eventTypes = ['工作日出勤', '节假日出勤', '补班出勤', '事假', '病假', '年假', '法定假', '福利假', '缺卡', '迟到', '早退']

export default function Attendance() {
  const [data, setData] = useState<any[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(false)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const [personId, setPersonId] = useState<number | undefined>()
  const [yearMonth, setYearMonth] = useState('')
  const [eventType, setEventType] = useState('')
  const [drawerOpen, setDrawerOpen] = useState(false)
  const [editingRecord, setEditingRecord] = useState<any>(null)
  const [persons, setPersons] = useState<any[]>([])
  const [leaveBalance, setLeaveBalance] = useState<any>(null)
  const [balancePersonId, setBalancePersonId] = useState<number | undefined>()
  const [form] = Form.useForm()

  useEffect(() => { getAllPersons().then((res: any) => setPersons(res.data || [])).catch(() => {}) }, [])

  const fetchData = async () => {
    setLoading(true)
    try {
      const params: any = { page, pageSize }
      if (personId) params.personId = personId
      if (yearMonth) params.yearMonth = yearMonth
      if (eventType) params.eventType = eventType
      const res: any = await getAttendanceEvents(params)
      setData(res.data.list || [])
      setTotal(res.data.total || 0)
    } finally { setLoading(false) }
  }

  useEffect(() => { fetchData() }, [page, pageSize])

  const handleCreate = () => { setEditingRecord(null); form.resetFields(); setDrawerOpen(true) }

  const handleEdit = (record: any) => {
    setEditingRecord(record)
    form.setFieldsValue({ ...record, eventDate: record.eventDate ? dayjs(record.eventDate) : undefined })
    setDrawerOpen(true)
  }

  const handleSave = async () => {
    const values = await form.validateFields()
    const payload = { ...values, eventDate: values.eventDate?.format('YYYY-MM-DD') }
    try {
      if (editingRecord) {
        await updateAttendanceEvent(editingRecord.id, payload)
      } else {
        await createAttendanceEvent(payload)
      }
      message.success('保存成功')
      setDrawerOpen(false)
      fetchData()
    } catch (e: any) { message.error(e.message || '保存失败') }
  }

  const handleDelete = async (id: number) => { await deleteAttendanceEvent(id); message.success('删除成功'); fetchData() }

  const handleQueryBalance = async () => {
    if (!balancePersonId) return
    try {
      const res: any = await getLeaveBalance(balancePersonId)
      setLeaveBalance(res.data)
    } catch (e: any) { message.error(e.message) }
  }

  const handleGrantLeave = async () => {
    if (!balancePersonId) return
    try {
      await grantAnnualLeave(balancePersonId)
      message.success('年假发放成功')
      handleQueryBalance()
    } catch (e: any) { message.error(e.message || '发放失败') }
  }

  const handleCloseMonth = async () => {
    if (!balancePersonId || !yearMonth) return
    try {
      await closeMonth(balancePersonId, yearMonth)
      message.success('月结完成')
      handleQueryBalance()
    } catch (e: any) { message.error(e.message || '月结失败') }
  }

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 60 },
    { title: '人员ID', dataIndex: 'personId', width: 70 },
    { title: '日期', dataIndex: 'eventDate', width: 110 },
    { title: '类型', dataIndex: 'eventType', width: 110 },
    { title: '开始', dataIndex: 'startTime', width: 80 },
    { title: '结束', dataIndex: 'endTime', width: 80 },
    { title: '时长(h)', dataIndex: 'durationHours', width: 80 },
    { title: '描述', dataIndex: 'description', ellipsis: true },
    { title: '创建时间', dataIndex: 'createdAt', width: 160, render: (v: string) => dayjs(v).format('MM-DD HH:mm') },
    {
      title: '操作', key: 'actions', width: 160,
      render: (_: any, record: any) => (
        <Space>
          <Button size="small" icon={<EditOutlined />} onClick={() => handleEdit(record)}>编辑</Button>
          <Popconfirm title="确定删除?" onConfirm={() => handleDelete(record.id)}>
            <Button size="small" danger icon={<DeleteOutlined />}>删除</Button>
          </Popconfirm>
        </Space>
      ),
    },
  ]

  return (
    <div>
      <Card size="small" style={{ marginBottom: 16 }}>
        <Space wrap>
          <Select placeholder="选择人员" value={balancePersonId} onChange={setBalancePersonId} allowClear showSearch style={{ width: 200 }}
            options={persons.map((p: any) => ({ label: `${p.id} - ${p.name}`, value: p.id }))}
            filterOption={(input, option) => (option?.label as string)?.toLowerCase().includes(input.toLowerCase())} />
          <Input placeholder="月份(如2024-01)" value={yearMonth} onChange={(e) => setYearMonth(e.target.value)} style={{ width: 140 }} />
          <Button type="primary" onClick={handleQueryBalance}>查询年假</Button>
          <Button onClick={handleGrantLeave}>发放年假</Button>
          <Button onClick={handleCloseMonth}>月结</Button>
        </Space>
        {leaveBalance && (
          <div style={{ marginTop: 12 }}>
            <Space size="large">
              <Statistic title="年假总额(天)" value={leaveBalance.totalDays} precision={1} />
              <Statistic title="已用(天)" value={leaveBalance.usedDays} precision={1} />
              <Statistic title="剩余(天)" value={leaveBalance.remaining} precision={1} />
            </Space>
          </div>
        )}
      </Card>

      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
        <Space>
          <Select placeholder="筛选人员" value={personId} onChange={(v) => { setPersonId(v); setPage(1); }} allowClear style={{ width: 180 }}
            options={persons.map((p: any) => ({ label: `${p.id} - ${p.name}`, value: p.id }))}
            filterOption={(input, option) => (option?.label as string)?.toLowerCase().includes(input.toLowerCase())} />
          <Select placeholder="事件类型" value={eventType || undefined} onChange={(v) => { setEventType(v || ''); setPage(1); }} allowClear style={{ width: 130 }}
            options={eventTypes.map(t => ({ label: t, value: t }))} />
          <Button type="primary" icon={<SearchOutlined />} onClick={() => { setPage(1); fetchData() }}>搜索</Button>
          <Button icon={<ReloadOutlined />} onClick={() => { setPersonId(undefined); setEventType(''); setYearMonth(''); setPage(1); setTimeout(fetchData, 0) }}>刷新</Button>
        </Space>
        <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>新增事件</Button>
      </div>

      <Table rowKey="id" columns={columns} dataSource={data} loading={loading}
        pagination={{ current: page, pageSize, total, showSizeChanger: true, onChange: (p, ps) => { setPage(p); setPageSize(ps) } }} />

      <Drawer title={editingRecord ? '编辑考勤事件' : '新增考勤事件'} open={drawerOpen} onClose={() => setDrawerOpen(false)} width={500}
        extra={<Button type="primary" onClick={handleSave}>保存</Button>}>
        <Form form={form} layout="vertical">
          <Form.Item name="personId" label="人员" rules={[{ required: true }]}>
            <Select showSearch options={persons.map((p: any) => ({ label: `${p.id} - ${p.name}`, value: p.id }))}
              filterOption={(input, option) => (option?.label as string)?.toLowerCase().includes(input.toLowerCase())} />
          </Form.Item>
          <Form.Item name="eventDate" label="日期" rules={[{ required: true }]}><DatePicker style={{ width: '100%' }} /></Form.Item>
          <Form.Item name="eventType" label="事件类型" rules={[{ required: true }]}>
            <Select options={eventTypes.map(t => ({ label: t, value: t }))} />
          </Form.Item>
          <Form.Item name="startTime" label="开始时间"><Input placeholder="09:00" /></Form.Item>
          <Form.Item name="endTime" label="结束时间"><Input placeholder="18:00" /></Form.Item>
          <Form.Item name="durationHours" label="时长(小时)"><InputNumber style={{ width: '100%' }} min={0} precision={1} /></Form.Item>
          <Form.Item name="description" label="描述"><Input.TextArea rows={3} /></Form.Item>
        </Form>
      </Drawer>
    </div>
  )
}
