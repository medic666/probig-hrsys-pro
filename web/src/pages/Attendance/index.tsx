import { useEffect, useState } from 'react';
import { Table, Button, Space, Select, Modal, Form, Input, DatePicker, InputNumber, Tabs, message, Popconfirm } from 'antd';
import { PlusOutlined, SearchOutlined, ExportOutlined, CalculatorOutlined } from '@ant-design/icons';
import dayjs from 'dayjs';
import { get, post, put, del } from '../../services/api';
import { useAuth } from '../../hooks/useAuth';
import * as XLSX from 'xlsx';

const eventTypes = [
  { label: '普通出勤', value: 'normal_attendance' },
  { label: '补班出勤', value: 'supplementary_attendance' },
  { label: '调休', value: 'compensatory_leave' },
  { label: '事假', value: 'personal_leave' },
  { label: '病假', value: 'sick_leave' },
  { label: '年假', value: 'annual_leave' },
  { label: '法定假', value: 'statutory_leave' },
  { label: '福利假', value: 'welfare_leave' },
  { label: '工作日加班', value: 'workday_overtime' },
  { label: '节假日加班', value: 'holiday_overtime' },
  { label: '缺卡', value: 'missing_clock' },
  { label: '迟到', value: 'late' },
  { label: '早退', value: 'early_leave' },
  { label: '年假配发', value: 'annual_leave_allot' },
  { label: '年假结转', value: 'annual_leave_carryover' },
];

export default function Attendance() {
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
      const res = await get<any>('/attendance-summaries', params);
      if (res.code === 0) { setSummaries(res.data.list || []); setSummaryTotal(res.data.total || 0); }
    } finally { setLoading(false); }
  };

  const fetchEvents = async () => {
    setLoading(true);
    try {
      const params: any = { page, page_size: 20, period };
      if (personID) params.person_id = personID;
      const res = await get<any>('/attendance-events', params);
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
    const res = await post<any>('/attendance/calculate', params);
    if (res.code === 0) { message.success('核算完成'); fetchSummaries(); }
    else message.error(res.message);
  };

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      const payload = {
        person_id: values.person_id,
        date: dayjs(values.date).format('YYYY-MM-DD'),
        event_type: values.event_type,
        duration: values.duration || 1,
        remark: values.remark || '',
      };
      if (editEvent) {
        const res = await put<any>(`/attendance-events/${editEvent.id}`, payload);
        if (res.code === 0) { message.success('更新成功'); setModalOpen(false); setEditEvent(null); form.resetFields(); fetchEvents(); }
        else message.error(res.message);
      } else {
        const res = await post<any>('/attendance-events', payload);
        if (res.code === 0) { message.success('添加成功'); setModalOpen(false); form.resetFields(); fetchEvents(); }
        else message.error(res.message);
      }
    } catch { /* ignore */ }
  };

  const handleDelete = async (eventId: number) => {
    const res = await del<any>(`/attendance-events/${eventId}`);
    if (res.code === 0) { message.success('删除成功'); fetchEvents(); }
    else message.error(res.message);
  };

  const handleExport = () => {
    const ws = XLSX.utils.json_to_sheet(summaries.map((d: any) => ({
      人员: d.person_name || `ID:${d.person_id}`,
      期间: d.period,
      普通出勤: d.normal_attendance_days,
      事假: d.personal_leave_days,
      病假: d.sick_leave_days,
      年假: d.annual_leave_days,
      工作日加班: d.workday_overtime_days,
      节假日加班: d.holiday_overtime_days,
      迟到: d.late_count,
      早退: d.early_leave_count,
      缺卡: d.missing_clock_count,
    })));
    const wb = XLSX.utils.book_new();
    XLSX.utils.book_append_sheet(wb, ws, '假勤汇总');
    XLSX.writeFile(wb, `假勤汇总_${period || '全部'}.xlsx`);
  };

  const summaryColumns = [
    { title: '人员', dataIndex: 'person_name', width: 100, render: (v: string, r: any) => v || `ID:${r.person_id}` },
    { title: '期间', dataIndex: 'period', width: 100 },
    { title: '普通出勤', dataIndex: 'normal_attendance_days', width: 90 },
    { title: '补班出勤', dataIndex: 'supplementary_attendance_days', width: 90 },
    { title: '事假', dataIndex: 'personal_leave_days', width: 70 },
    { title: '病假', dataIndex: 'sick_leave_days', width: 70 },
    { title: '年假', dataIndex: 'annual_leave_days', width: 70 },
    { title: '加班(工)', dataIndex: 'workday_overtime_days', width: 80 },
    { title: '加班(假)', dataIndex: 'holiday_overtime_days', width: 80 },
    { title: '违纪', dataIndex: 'violation_count', width: 70 },
  ];

  const eventColumns = [
    { title: 'ID', dataIndex: 'id', width: 60 },
    { title: '人员', dataIndex: 'person_name', width: 100, render: (v: string, r: any) => v || `ID:${r.person_id}` },
    { title: '日期', dataIndex: 'date', width: 110 },
    { title: '类型', dataIndex: 'event_type', width: 110, render: (t: string) => eventTypes.find(e => e.value === t)?.label || t },
    { title: '天数/次数', dataIndex: 'duration', width: 90 },
    { title: '备注', dataIndex: 'remark', width: 120, ellipsis: true },
    { title: '操作', key: 'actions', width: 150, render: (_: any, r: any) => (
      <Space size="small">
        <Button type="link" size="small" onClick={() => { setEditEvent(r); form.setFieldsValue({ ...r, date: dayjs(r.date) }); setModalOpen(true); }}>编辑</Button>
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
          {hasPermission('attendance', 'write') && (
            <>
              <Button icon={<CalculatorOutlined />} onClick={handleCalculate}>考勤核算</Button>
              <Button type="primary" icon={<PlusOutlined />} onClick={() => { setEditEvent(null); form.resetFields(); setModalOpen(true); }}>添加假勤事件</Button>
            </>
          )}
          <Button icon={<ExportOutlined />} onClick={handleExport}>导出</Button>
        </Space>
      </Space>

      <Tabs activeKey={activeTab} onChange={setActiveTab} items={[
        {
          key: 'summary', label: '假勤汇总',
          children: <Table rowKey="id" columns={summaryColumns} dataSource={summaries} loading={loading}
            pagination={{ current: page, total: summaryTotal, pageSize: 20, onChange: setPage }}
            scroll={{ x: 1100 }} size="small" />,
        },
        {
          key: 'events', label: '假勤事件',
          children: <Table rowKey="id" columns={eventColumns} dataSource={events} loading={loading}
            pagination={{ current: page, total: eventTotal, pageSize: 20, onChange: setPage }}
            scroll={{ x: 800 }} size="small" />,
        },
      ]} />

      <Modal title={editEvent ? '编辑假勤事件' : '添加假勤事件'} open={modalOpen}
        onCancel={() => { setModalOpen(false); setEditEvent(null); form.resetFields(); }} onOk={handleSubmit}
        okText="确定" cancelText="取消" destroyOnClose>
        <Form form={form} layout="vertical">
          <Form.Item name="person_id" label="人员" rules={[{ required: true }]}>
            <Select showSearch placeholder="选择人员" options={personOptions}
              filterOption={(input, option) => (option?.label as string || '').includes(input)} />
          </Form.Item>
          <Form.Item name="date" label="日期" rules={[{ required: true }]}><DatePicker style={{ width: '100%' }} /></Form.Item>
          <Form.Item name="event_type" label="事件类型" rules={[{ required: true }]}><Select options={eventTypes} /></Form.Item>
          <Form.Item name="duration" label="天数/次数" initialValue={1}><InputNumber style={{ width: '100%' }} min={0} step={0.5} /></Form.Item>
          <Form.Item name="remark" label="备注"><Input.TextArea rows={2} /></Form.Item>
        </Form>
      </Modal>
    </div>
  );
}
