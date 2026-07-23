import { useState, useEffect } from 'react'
import { Table, Button, Space, Input, Drawer, Form, InputNumber, Select, DatePicker, Tabs, message, Popconfirm, Modal } from 'antd'
import { PlusOutlined, SearchOutlined, ReloadOutlined, EditOutlined, DeleteOutlined, HistoryOutlined, ExportOutlined } from '@ant-design/icons'
import { getPersons, createPerson, updatePerson, deletePerson, getPersonTimeline, getAllPersons } from '../api/persons'
import EventTimeline from '../components/EventTimeline'
import dayjs from 'dayjs'
import * as XLSX from 'xlsx'

const genderOptions = ['男', '女', '其他']
const maritalOptions = ['未婚', '已婚', '离异', '丧偶']
const politicalOptions = ['群众', '共青团员', '中共党员', '民主党派', '无党派人士']

export default function Persons() {
  const [data, setData] = useState<any[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(false)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const [search, setSearch] = useState('')
  const [drawerOpen, setDrawerOpen] = useState(false)
  const [editingPerson, setEditingPerson] = useState<any>(null)
  const [timelineOpen, setTimelineOpen] = useState(false)
  const [timelineEvents, setTimelineEvents] = useState<any[]>([])
  const [form] = Form.useForm()

  const fetchData = async () => {
    setLoading(true)
    try {
      const res: any = await getPersons({ page, pageSize, search })
      setData(res.data.list || [])
      setTotal(res.data.total || 0)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { fetchData() }, [page, pageSize])

  const handleSearch = () => { setPage(1); fetchData() }

  const handleCreate = () => {
    setEditingPerson(null)
    form.resetFields()
    form.setFieldsValue({ salaryDays: 21.75 })
    setDrawerOpen(true)
  }

  const handleEdit = (record: any) => {
    setEditingPerson(record)
    form.setFieldsValue({
      ...record,
      hireDate: record.hireDate ? dayjs(record.hireDate) : undefined,
      birthday: record.birthday ? dayjs(record.birthday) : undefined,
    })
    setDrawerOpen(true)
  }

  const handleSave = async () => {
    const values = await form.validateFields()
    const payload = {
      ...values,
      hireDate: values.hireDate ? values.hireDate.format('YYYY-MM-DD') : '',
      birthday: values.birthday ? values.birthday.format('YYYY-MM-DD') : '',
      phones: JSON.stringify(values.phones || []),
      emails: JSON.stringify(values.emails || []),
      bankCards: JSON.stringify(values.bankCards || []),
      resume: JSON.stringify(values.resume || []),
    }
    try {
      if (editingPerson) {
        await updatePerson(editingPerson.id, payload)
      } else {
        await createPerson(payload)
      }
      message.success(editingPerson ? '更新成功' : '创建成功')
      setDrawerOpen(false)
      fetchData()
    } catch (e: any) {
      message.error(e.message || '保存失败')
    }
  }

  const handleDelete = async (id: number) => {
    await deletePerson(id)
    message.success('删除成功')
    fetchData()
  }

  const handleTimeline = async (id: number) => {
    const res: any = await getPersonTimeline(id)
    setTimelineEvents(res.data || [])
    setTimelineOpen(true)
  }

  const handleExport = async () => {
    const res: any = await getAllPersons()
    const persons = res.data || []
    const exportData = persons.map((p: any) => ({
      姓名: p.name, 性别: p.gender, 入职日期: p.hireDate,
      基本工资: p.baseSalary, 绩效工资: p.performanceSalary,
      身份证号: p.idNumber, 电话: p.phones, 邮箱: p.emails,
    }))
    const ws = XLSX.utils.json_to_sheet(exportData)
    const wb = XLSX.utils.book_new()
    XLSX.utils.book_append_sheet(wb, ws, '人员列表')
    XLSX.writeFile(wb, 'persons.xlsx')
  }

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 60 },
    { title: '姓名', dataIndex: 'name', width: 100 },
    { title: '性别', dataIndex: 'gender', width: 60 },
    { title: '身份证号', dataIndex: 'idNumber', width: 180 },
    { title: '入职日期', dataIndex: 'hireDate', width: 100 },
    { title: '基本工资', dataIndex: 'baseSalary', width: 100, render: (v: number) => v?.toFixed(2) },
    { title: '考勤组', dataIndex: 'attendanceGroup', width: 100 },
    {
      title: '操作', key: 'actions', width: 260, fixed: 'right' as const,
      render: (_: any, record: any) => (
        <Space>
          <Button size="small" icon={<EditOutlined />} onClick={() => handleEdit(record)}>编辑</Button>
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
          <Input placeholder="搜索姓名/身份证/电话" value={search} onChange={(e) => setSearch(e.target.value)} onPressEnter={handleSearch} style={{ width: 250 }} prefix={<SearchOutlined />} />
          <Button type="primary" onClick={handleSearch}>搜索</Button>
          <Button icon={<ReloadOutlined />} onClick={() => { setSearch(''); setPage(1); setTimeout(fetchData, 0) }}>刷新</Button>
        </Space>
        <Space>
          <Button icon={<ExportOutlined />} onClick={handleExport}>导出</Button>
          <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>新增人员</Button>
        </Space>
      </div>

      <Table rowKey="id" columns={columns} dataSource={data} loading={loading}
        pagination={{ current: page, pageSize, total, showSizeChanger: true, showTotal: (t) => `共 ${t} 条`, onChange: (p, ps) => { setPage(p); setPageSize(ps) } }}
        scroll={{ x: 1000 }} />

      <Drawer title={editingPerson ? '编辑人员' : '新增人员'} open={drawerOpen} onClose={() => setDrawerOpen(false)} width={720}
        extra={<Button type="primary" onClick={handleSave}>保存</Button>}>
        <Form form={form} layout="vertical">
          <Tabs items={[
            { key: 'basic', label: '基础信息', children: (
              <>
                <Form.Item name="name" label="姓名" rules={[{ required: true }]}><Input /></Form.Item>
                <Form.Item name="gender" label="性别"><Select options={genderOptions.map(v => ({ label: v, value: v }))} /></Form.Item>
                <Form.Item name="idNumber" label="身份证号"><Input /></Form.Item>
                <Form.Item name="birthday" label="生日"><DatePicker style={{ width: '100%' }} /></Form.Item>
                <Form.Item name="ethnicity" label="民族"><Input /></Form.Item>
                <Form.Item name="nativePlace" label="籍贯"><Input /></Form.Item>
                <Form.Item name="address" label="住址"><Input /></Form.Item>
                <Form.Item name="alias" label="别名"><Input /></Form.Item>
                <Form.Item name="politicalStatus" label="政治面貌"><Select options={politicalOptions.map(v => ({ label: v, value: v }))} /></Form.Item>
                <Form.Item name="maritalStatus" label="婚姻状态"><Select options={maritalOptions.map(v => ({ label: v, value: v }))} /></Form.Item>
              </>
            )},
            { key: 'job', label: '职务信息', children: (
              <>
                <Form.Item name="attendanceGroup" label="考勤组"><Input /></Form.Item>
                <Form.Item name="hireDate" label="入职日期"><DatePicker style={{ width: '100%' }} /></Form.Item>
                <Form.Item name="baseSalary" label="基本工资"><InputNumber style={{ width: '100%' }} min={0} precision={2} /></Form.Item>
                <Form.Item name="performanceSalary" label="绩效工资"><InputNumber style={{ width: '100%' }} min={0} precision={2} /></Form.Item>
                <Form.Item name="salaryDays" label="计薪天数"><InputNumber style={{ width: '100%' }} min={1} max={31} /></Form.Item>
                <Form.Item name="positionAllowance" label="职位津贴"><InputNumber style={{ width: '100%' }} min={0} precision={2} /></Form.Item>
                <Form.Item name="mealSubsidy" label="餐补"><InputNumber style={{ width: '100%' }} min={0} precision={2} /></Form.Item>
                <Form.Item name="housingSubsidy" label="房补"><InputNumber style={{ width: '100%' }} min={0} precision={2} /></Form.Item>
                <Form.Item name="transportSubsidy" label="交通补贴"><InputNumber style={{ width: '100%' }} min={0} precision={2} /></Form.Item>
                <Form.Item name="heatSubsidy" label="高温补贴"><InputNumber style={{ width: '100%' }} min={0} precision={2} /></Form.Item>
                <Form.Item name="insuranceSubsidy" label="保险补贴"><InputNumber style={{ width: '100%' }} min={0} precision={2} /></Form.Item>
                <Form.Item name="housingFundSubsidy" label="公积金补偿"><InputNumber style={{ width: '100%' }} min={0} precision={2} /></Form.Item>
              </>
            )},
            { key: 'deduct', label: '代扣项', children: (
              <>
                <Form.Item name="socialInsuranceDeduct" label="社保代扣"><InputNumber style={{ width: '100%' }} min={0} precision={2} /></Form.Item>
                <Form.Item name="housingFundDeduct" label="公积金代扣"><InputNumber style={{ width: '100%' }} min={0} precision={2} /></Form.Item>
                <Form.Item name="taxDeduct" label="个税代扣"><InputNumber style={{ width: '100%' }} min={0} precision={2} /></Form.Item>
              </>
            )},
            { key: 'contact', label: '联系方式', children: (
              <>
                <Form.Item name="phones" label="电话（JSON数组）"><Input.TextArea rows={2} placeholder='[{"type":"工作","value":"13800138000"}]' /></Form.Item>
                <Form.Item name="emails" label="邮箱（JSON数组）"><Input.TextArea rows={2} placeholder='[{"type":"工作","value":"a@b.com"}]' /></Form.Item>
                <Form.Item name="bankCards" label="银行卡（JSON数组）"><Input.TextArea rows={3} placeholder='[{"bankName":"工商银行","account":"622202..."}]' /></Form.Item>
                <Form.Item name="resume" label="简历（JSON数组）"><Input.TextArea rows={4} placeholder='[{"startDate":"2020-01","endDate":"2023-01","company":"XX公司","position":"工程师"}]' /></Form.Item>
              </>
            )},
          ]} />
        </Form>
      </Drawer>

      <Modal title="事件时间线" open={timelineOpen} onCancel={() => setTimelineOpen(false)} footer={null} width={600}>
        <EventTimeline events={timelineEvents} />
      </Modal>
    </div>
  )
}
