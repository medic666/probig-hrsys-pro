import { useState, useEffect } from 'react'
import { Table, Button, Space, Select, Input, message, Card, Drawer, Form, InputNumber, Modal, Descriptions, Popconfirm } from 'antd'
import { SearchOutlined, ReloadOutlined, CalculatorOutlined, DollarOutlined, ExportOutlined } from '@ant-design/icons'
import { getSalaryRecords, calculateSalary, addAdjustment, deleteAdjustment, getAdjustments, getSalaryRecord } from '../api/salary'
import { getAllPersons } from '../api/persons'

export default function Salary() {
  const [data, setData] = useState<any[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(false)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const [personId, setPersonId] = useState<number | undefined>()
  const [yearMonth, setYearMonth] = useState('')
  const [persons, setPersons] = useState<any[]>([])
  const [detailOpen, setDetailOpen] = useState(false)
  const [detailRecord, setDetailRecord] = useState<any>(null)
  const [adjustDrawerOpen, setAdjustDrawerOpen] = useState(false)
  const [adjustPersonId, setAdjustPersonId] = useState<number | undefined>()
  const [adjustYearMonth, setAdjustYearMonth] = useState('')
  const [adjustments, setAdjustments] = useState<any[]>([])
  const [adjustForm] = Form.useForm()

  useEffect(() => { getAllPersons().then((res: any) => setPersons(res.data || [])).catch(() => {}) }, [])

  const fetchData = async () => {
    setLoading(true)
    try {
      const params: any = { page, pageSize }
      if (personId) params.personId = personId
      if (yearMonth) params.yearMonth = yearMonth
      const res: any = await getSalaryRecords(params)
      setData(res.data.list || [])
      setTotal(res.data.total || 0)
    } finally { setLoading(false) }
  }

  useEffect(() => { fetchData() }, [page, pageSize])

  const handleCalculate = async () => {
    if (!personId || !yearMonth) { message.warning('请选择人员和月份'); return }
    try {
      await calculateSalary(personId, yearMonth)
      message.success('工资计算完成')
      fetchData()
    } catch (e: any) { message.error(e.message || '计算失败') }
  }

  const handleViewDetail = async (record: any) => {
    try {
      const res: any = await getSalaryRecord(record.personId, record.yearMonth)
      setDetailRecord(res.data)
      setDetailOpen(true)
    } catch (e: any) { message.error('获取详情失败') }
  }

  const handleAdjustments = async () => {
    if (!adjustPersonId || !adjustYearMonth) { message.warning('请选择人员和月份'); return }
    try {
      const res: any = await getAdjustments(adjustPersonId, adjustYearMonth)
      setAdjustments(res.data || [])
      setAdjustDrawerOpen(true)
    } catch (e: any) { message.error(e.message) }
  }

  const handleAddAdjustment = async () => {
    const values = await adjustForm.validateFields()
    try {
      await addAdjustment({ ...values, personId: adjustPersonId, yearMonth: adjustYearMonth })
      message.success('添加成功')
      adjustForm.resetFields()
      handleAdjustments()
    } catch (e: any) { message.error(e.message || '添加失败') }
  }

  const handleDeleteAdjustment = async (id: number) => {
    await deleteAdjustment(id)
    message.success('删除成功')
    handleAdjustments()
  }

  const handleExport = async () => {
    if (!yearMonth) { message.warning('请选择月份'); return }
    try {
      const res = await fetch(`/api/v1/export/salary?yearMonth=${yearMonth}`, { headers: { Authorization: `Bearer ${localStorage.getItem('token')}` } })
      const blob = await res.blob()
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url; a.download = `salary_${yearMonth}.xlsx`; a.click()
      URL.revokeObjectURL(url)
    } catch (e) { message.error('导出失败') }
  }

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 60 },
    { title: '人员ID', dataIndex: 'personId', width: 70 },
    { title: '月份', dataIndex: 'yearMonth', width: 100 },
    { title: '基本工资', dataIndex: 'baseSalary', width: 100, render: (v: number) => v?.toFixed(2) },
    { title: '考勤工资', dataIndex: 'attendanceSalary', width: 100, render: (v: number) => v?.toFixed(2) },
    { title: '绩效工资', dataIndex: 'performanceSalary', width: 100, render: (v: number) => v?.toFixed(2) },
    { title: '补贴合计', dataIndex: 'totalAllowances', width: 100, render: (v: number) => v?.toFixed(2) },
    { title: '扣款合计', dataIndex: 'totalDeductions', width: 100, render: (v: number) => v?.toFixed(2) },
    { title: '实发工资', dataIndex: 'netSalary', width: 100, render: (v: number) => <strong style={{ color: '#1677ff' }}>{v?.toFixed(2)}</strong> },
    {
      title: '操作', key: 'actions', width: 120,
      render: (_: any, record: any) => (
        <Button size="small" onClick={() => handleViewDetail(record)} icon={<DollarOutlined />}>详情</Button>
      ),
    },
  ]

  return (
    <div>
      <Card size="small" style={{ marginBottom: 16 }}>
        <Space wrap>
          <Select placeholder="选择人员" value={personId} onChange={setPersonId} allowClear showSearch style={{ width: 200 }}
            options={persons.map((p: any) => ({ label: `${p.id} - ${p.name}`, value: p.id }))}
            filterOption={(input, option) => (option?.label as string)?.toLowerCase().includes(input.toLowerCase())} />
          <Input placeholder="月份(如2024-01)" value={yearMonth} onChange={(e) => setYearMonth(e.target.value)} style={{ width: 140 }} />
          <Button type="primary" icon={<CalculatorOutlined />} onClick={handleCalculate}>计算工资</Button>
          <Button onClick={() => { setAdjustPersonId(personId); setAdjustYearMonth(yearMonth); handleAdjustments() }}>调整事件</Button>
          <Button icon={<ExportOutlined />} onClick={handleExport}>导出</Button>
          <Button icon={<SearchOutlined />} onClick={() => { setPage(1); fetchData() }}>查询</Button>
          <Button icon={<ReloadOutlined />} onClick={() => { setPersonId(undefined); setYearMonth(''); setPage(1); setTimeout(fetchData, 0) }}>刷新</Button>
        </Space>
      </Card>

      <Table rowKey="id" columns={columns} dataSource={data} loading={loading}
        pagination={{ current: page, pageSize, total, showSizeChanger: true, onChange: (p, ps) => { setPage(p); setPageSize(ps) } }} />

      <Modal title="工资详情" open={detailOpen} onCancel={() => setDetailOpen(false)} footer={null} width={700}>
        {detailRecord && (
          <div>
            <Descriptions bordered size="small" column={2}>
              <Descriptions.Item label="基本工资">{detailRecord.baseSalary?.toFixed(2)}</Descriptions.Item>
              <Descriptions.Item label="考勤工资">{detailRecord.attendanceSalary?.toFixed(2)}</Descriptions.Item>
              <Descriptions.Item label="绩效工资">{detailRecord.performanceSalary?.toFixed(2)}</Descriptions.Item>
              <Descriptions.Item label="补贴合计">{detailRecord.totalAllowances?.toFixed(2)}</Descriptions.Item>
              <Descriptions.Item label="扣款合计">{detailRecord.totalDeductions?.toFixed(2)}</Descriptions.Item>
              <Descriptions.Item label="实发工资"><strong>{detailRecord.netSalary?.toFixed(2)}</strong></Descriptions.Item>
            </Descriptions>
            {detailRecord.detail && (
              <pre style={{ marginTop: 16, background: '#f5f5f5', padding: 12, borderRadius: 4, maxHeight: 400, overflow: 'auto' }}>
                {JSON.stringify(typeof detailRecord.detail === 'string' ? JSON.parse(detailRecord.detail) : detailRecord.detail, null, 2)}
              </pre>
            )}
          </div>
        )}
      </Modal>

      <Drawer title="工资调整事件" open={adjustDrawerOpen} onClose={() => setAdjustDrawerOpen(false)} width={500}>
        <Form form={adjustForm} layout="vertical" style={{ marginBottom: 16 }}>
          <Form.Item name="adjustmentType" label="类型" rules={[{ required: true }]}>
            <Select options={[{ label: '业绩', value: 'performance' }, { label: '奖励', value: 'bonus' }, { label: '惩罚', value: 'penalty' }, { label: '其他', value: 'other' }]} />
          </Form.Item>
          <Form.Item name="amount" label="金额" rules={[{ required: true }]}><InputNumber style={{ width: '100%' }} precision={2} /></Form.Item>
          <Form.Item name="description" label="描述"><Input.TextArea rows={3} /></Form.Item>
          <Button type="primary" onClick={handleAddAdjustment}>添加调整</Button>
        </Form>
        <Table rowKey="id" dataSource={adjustments} pagination={false} size="small"
          columns={[
            { title: 'ID', dataIndex: 'id', width: 60 },
            { title: '类型', dataIndex: 'adjustmentType', width: 80 },
            { title: '金额', dataIndex: 'amount', width: 100, render: (v: number) => <span style={{ color: v >= 0 ? 'green' : 'red' }}>{v?.toFixed(2)}</span> },
            { title: '描述', dataIndex: 'description' },
            { title: '操作', width: 60, render: (_: any, r: any) => (
              <Popconfirm title="删除?" onConfirm={() => handleDeleteAdjustment(r.id)}><Button size="small" danger>删除</Button></Popconfirm>
            )},
          ]} />
      </Drawer>
    </div>
  )
}
