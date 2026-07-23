import { useEffect, useState } from 'react';
import { Table, Button, Space, Select, Modal, Form, Input, DatePicker, InputNumber, Tabs, message, Popconfirm } from 'antd';
import { PlusOutlined, SearchOutlined, ExportOutlined, CalculatorOutlined } from '@ant-design/icons';
import dayjs from 'dayjs';
import { get, post, put, del } from '../../services/api';
import { useAuth } from '../../hooks/useAuth';
import * as XLSX from 'xlsx';

const eventTypes = [
  { label: '业绩', value: 'performance' },
  { label: '奖惩', value: 'reward_punish' },
  { label: '借款扣除', value: 'loan_deduct' },
  { label: '个税扣除', value: 'tax_deduct' },
  { label: '其他', value: 'other' },
];

export default function Salary() {
  const [summaries, setSummaries] = useState<any[]>([]);
  const [events, setEvents] = useState<any[]>([]);
  const [summaryTotal, setSummaryTotal] = useState(0);
  const [eventTotal, setEventTotal] = useState(0);
  const [loading, setLoading] = useState(false);
  const [page, setPage] = useState(1);
  const [period, setPeriod] = useState('');
  const [personID, setPersonID] = useState<number | undefined>(undefined);
  const [personOptions, setPersonOptions] = useState<{ label: string; value: number }[]>([]);
  const [modalOpen, setModalOpen] = useState(false);
  const [editEvent, setEditEvent] = useState<any>(null);
  const [form] = Form.useForm();
  const { hasPermission } = useAuth();
  const [activeTab, setActiveTab] = useState('summary');

  useEffect(() => {
    (async () => {
      try {
        const res = await get<any>('/persons', { page: 1, page_size: 1000 });
        if (res.code === 0) {
          const opts = (res.data.list || []).map((p: any) => ({
            label: `${p.name || p.info?.name || '-'} (ID:${p.id})`,
            value: p.id,
          }));
          setPersonOptions(opts);
        }
      } catch { /* ignore */ }
    })();
  }, []);

  const fetchSummaries = async () => {
    setLoading(true);
    try {
      const params: any = { page, page_size: 20, period };
      if (personID) params.person_id = personID;
      const res = await get<any>('/salary-summaries', params);
      if (res.code === 0) { setSummaries(res.data.list || []); setSummaryTotal(res.data.total || 0); }
    } finally { setLoading(false); }
  };

  const fetchEvents = async () => {
    setLoading(true);
    try {
      const params: any = { page, page_size: 20, period };
      if (personID) params.person_id = personID;
      const res = await get<any>('/salary-events', params);
      if (res.code === 0) { setEvents(res.data.list || []); setEventTotal(res.data.total || 0); }
    } finally { setLoading(false); }
  };

  useEffect(() => {
    if (activeTab === 'summary') fetchSummaries();
    else fetchEvents();
  }, [page, period, personID, activeTab]);

  const handleCalculate = async () => {
    if (!period) { message.warning('请先选择期间'); return; }
    const params: any = { period };
    if (personID) params.person_id = personID;
    const res = await post<any>('/salary/calculate', params);
    if (res.code === 0) { message.success('核算完成'); fetchSummaries(); }
    else message.error(res.message);
  };

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      const payload = {
        person_id: values.person_id,
        period: values.period ? (typeof values.period === 'string' ? values.period : dayjs(values.period).format('YYYY-MM')) : '',
        event_type: values.event_type,
        amount: values.amount || 0,
        detail: values.detail || '',
      };
      if (editEvent) {
        const res = await put<any>(`/salary-events/${editEvent.id}`, payload);
        if (res.code === 0) { message.success('更新成功'); setModalOpen(false); setEditEvent(null); form.resetFields(); fetchEvents(); }
        else message.error(res.message);
      } else {
        const res = await post<any>('/salary-events', payload);
        if (res.code === 0) { message.success('添加成功'); setModalOpen(false); form.resetFields(); fetchEvents(); }
        else message.error(res.message);
      }
    } catch { /* ignore */ }
  };

  const handleDelete = async (eventId: number) => {
    const res = await del<any>(`/salary-events/${eventId}`);
    if (res.code === 0) { message.success('删除成功'); fetchEvents(); }
    else message.error(res.message);
  };

  const handleExport = () => {
    const ws = XLSX.utils.json_to_sheet(summaries.map((d: any) => ({
      人员: d.person_name || `ID:${d.person_id}`,
      期间: d.period,
      出勤工资: d.attendance_salary,
      全勤奖: d.full_attendance_bonus,
      加班工资: d.overtime_salary,
      绩效工资: d.performance_salary,
      餐补: d.meal_subsidy,
      社保代扣: d.social_insurance_deduct,
      公积金代扣: d.housing_fund_deduct,
      实发合计: d.total_salary,
    })));
    const wb = XLSX.utils.book_new();
    XLSX.utils.book_append_sheet(wb, ws, '工资汇总');
    XLSX.writeFile(wb, `工资汇总_${period || '全部'}.xlsx`);
  };

  const summaryColumns = [
    { title: '人员', dataIndex: 'person_name', width: 100, render: (v: string, r: any) => v || `ID:${r.person_id}` },
    { title: '期间', dataIndex: 'period', width: 100 },
    { title: '出勤工资', dataIndex: 'attendance_salary', width: 100, render: (v: number) => `¥${v}` },
    { title: '全勤奖', dataIndex: 'full_attendance_bonus', width: 90, render: (v: number) => `¥${v}` },
    { title: '加班工资', dataIndex: 'overtime_salary', width: 100, render: (v: number) => `¥${v}` },
    { title: '绩效工资', dataIndex: 'performance_salary', width: 100, render: (v: number) => `¥${v}` },
    { title: '社保代扣', dataIndex: 'social_insurance_deduct', width: 90, render: (v: number) => `¥${v}` },
    { title: '公积金代扣', dataIndex: 'housing_fund_deduct', width: 90, render: (v: number) => `¥${v}` },
    { title: '奖惩', dataIndex: 'reward_punish', width: 80, render: (v: number) => `¥${v}` },
    { title: '借款扣除', dataIndex: 'loan_deduct', width: 90, render: (v: number) => `¥${v}` },
    { title: '个税扣除', dataIndex: 'tax_deduct', width: 90, render: (v: number) => `¥${v}` },
    {
      title: '实发合计', dataIndex: 'total_salary', width: 110, fixed: 'right' as const,
      render: (v: number) => <strong>¥{v}</strong>,
    },
  ];

  const eventColumns = [
    { title: 'ID', dataIndex: 'id', width: 60 },
    { title: '人员', dataIndex: 'person_name', width: 100, render: (v: string, r: any) => v || `ID:${r.person_id}` },
    { title: '期间', dataIndex: 'period', width: 100 },
    { title: '类型', dataIndex: 'event_type', width: 100, render: (t: string) => eventTypes.find(e => e.value === t)?.label || t },
    { title: '金额', dataIndex: 'amount', width: 100, render: (v: number) => `¥${v}` },
    { title: '描述', dataIndex: 'detail', width: 150, ellipsis: true },
    { title: '操作', key: 'actions', width: 150, render: (_: any, r: any) => (
      <Space size="small">
        <Button type="link" size="small" onClick={() => { setEditEvent(r); form.setFieldsValue({ ...r, period: r.period ? dayjs(r.period + '-01') : undefined }); setModalOpen(true); }}>编辑</Button>
        <Popconfirm title="确定删除?" onConfirm={() => handleDelete(r.id)}>
          <Button type="link" size="small" danger>删除</Button>
        </Popconfirm>
      </Space>
    )},
  ];

  return (
    <div>
      <Space style={{ marginBottom: 16, justifyContent: 'space-between', width: '100%' }}>
        <Space>
          <Select
            showSearch
            placeholder="选择人员（留空查全部）"
            value={personID}
            onChange={(v) => { setPersonID(v); setPage(1); }}
            allowClear
            style={{ width: 220 }}
            filterOption={(input, option) => (option?.label as string || '').includes(input)}
            options={personOptions}
          />
          <DatePicker
            picker="month"
            placeholder="选择期间"
            value={period ? dayjs(period + '-01') : null}
            onChange={(d) => { setPeriod(d ? d.format('YYYY-MM') : ''); setPage(1); }}
            allowClear
            style={{ width: 140 }}
          />
          <Button type="primary" icon={<SearchOutlined />} onClick={() => { setPage(1); activeTab === 'summary' ? fetchSummaries() : fetchEvents(); }}>搜索</Button>
        </Space>
        <Space>
          {hasPermission('salary', 'write') && (
            <>
              <Button icon={<CalculatorOutlined />} onClick={handleCalculate}>工资核算</Button>
              <Button type="primary" icon={<PlusOutlined />} onClick={() => { setEditEvent(null); form.resetFields(); setModalOpen(true); }}>添加工资事件</Button>
            </>
          )}
          <Button icon={<ExportOutlined />} onClick={handleExport}>导出</Button>
        </Space>
      </Space>

      <Tabs activeKey={activeTab} onChange={setActiveTab} items={[
        {
          key: 'summary', label: '工资汇总',
          children: <Table rowKey="id" columns={summaryColumns} dataSource={summaries} loading={loading}
            pagination={{ current: page, total: summaryTotal, pageSize: 20, onChange: setPage }}
            scroll={{ x: 1400 }} size="small" />,
        },
        {
          key: 'events', label: '工资事件',
          children: <Table rowKey="id" columns={eventColumns} dataSource={events} loading={loading}
            pagination={{ current: page, total: eventTotal, pageSize: 20, onChange: setPage }}
            scroll={{ x: 800 }} size="small" />,
        },
      ]} />

      <Modal title={editEvent ? '编辑工资事件' : '添加工资事件'} open={modalOpen}
        onCancel={() => { setModalOpen(false); setEditEvent(null); form.resetFields(); }} onOk={handleSubmit}
        okText="确定" cancelText="取消" destroyOnClose>
        <Form form={form} layout="vertical">
          <Form.Item name="person_id" label="人员" rules={[{ required: true }]}>
            <Select showSearch placeholder="选择人员" options={personOptions}
              filterOption={(input, option) => (option?.label as string || '').includes(input)} />
          </Form.Item>
          <Form.Item name="period" label="期间" rules={[{ required: true }]}>
            <DatePicker picker="month" style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="event_type" label="事件类型" rules={[{ required: true }]}><Select options={eventTypes} /></Form.Item>
          <Form.Item name="amount" label="金额" rules={[{ required: true }]}><InputNumber style={{ width: '100%' }} /></Form.Item>
          <Form.Item name="detail" label="描述"><Input.TextArea rows={2} /></Form.Item>
        </Form>
      </Modal>
    </div>
  );
}
