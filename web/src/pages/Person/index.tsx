import { useEffect, useState } from 'react';
import { Table, Button, Space, Input, Tag, Modal, Form, InputNumber, Select, DatePicker, message, Popconfirm, Row, Col } from 'antd';
import { PlusOutlined, SearchOutlined, ExportOutlined, EyeOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import dayjs from 'dayjs';
import { get, post } from '../../services/api';
import { useAuth } from '../../hooks/useAuth';
import * as XLSX from 'xlsx';

export default function PersonList() {
  const [data, setData] = useState<any[]>([]);
  const [loading, setLoading] = useState(false);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [keyword, setKeyword] = useState('');
  const [modalOpen, setModalOpen] = useState(false);
  const [form] = Form.useForm();
  const { hasPermission } = useAuth();
  const navigate = useNavigate();

  const fetchData = async () => {
    setLoading(true);
    try {
      const res = await get<any>('/persons', { page, page_size: 20, keyword });
      if (res.code === 0) {
        setData(res.data.list || []);
        setTotal(res.data.total || 0);
      }
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { fetchData(); }, [page, keyword]);

  const handleCreate = async () => {
    try {
      const values = await form.validateFields();
      const payload = {
        person_id: 0,
        effective_date: values.entry_date ? dayjs(values.entry_date).format('YYYY-MM-DD') : dayjs().format('YYYY-MM-DD'),
        event_type: 'onboard',
        data: {
          name: values.name,
          attendance_group: values.attendance_group || '',
          entry_date: values.entry_date ? dayjs(values.entry_date).format('YYYY-MM-DD') : '',
          basic_salary: values.basic_salary || 0,
          performance_salary: values.performance_salary || 0,
          salary_days: values.salary_days || 21.75,
          position_allowance: values.position_allowance || 0,
          meal_subsidy: values.meal_subsidy || 0,
          housing_subsidy: values.housing_subsidy || 0,
          transport_subsidy: values.transport_subsidy || 0,
          heat_subsidy: values.heat_subsidy || 0,
          insurance_compensation: values.insurance_compensation || 0,
          housing_fund_compensation: values.housing_fund_compensation || 0,
          social_insurance_deduct: values.social_insurance_deduct || 0,
          housing_fund_deduct: values.housing_fund_deduct || 0,
          phones: values.phones ? [values.phones] : [],
          emails: values.emails ? [values.emails] : [],
          id_number: values.id_number || '',
          gender: values.gender || '',
          birthday: values.birthday ? dayjs(values.birthday).format('YYYY-MM-DD') : '',
          ethnicity: values.ethnicity || '',
          native_place: values.native_place || '',
          address: values.address || '',
          bank_cards: values.bank_cards ? [values.bank_cards] : [],
          political_status: values.political_status || '',
          marital_status: values.marital_status || '',
          alias: values.alias || '',
        },
      };
      const res = await post<any>('/persons', payload);
      if (res.code === 0) {
        message.success('添加成功');
        setModalOpen(false);
        form.resetFields();
        fetchData();
      } else {
        message.error(res.message);
      }
    } catch { /* form validation error */ }
  };

  const handleExport = () => {
    const ws = XLSX.utils.json_to_sheet(data.map((d: any) => ({
      姓名: d.name || d.info?.name,
      考勤组: d.info?.attendance_group,
      入职日期: d.info?.entry_date,
      基本工资: d.info?.basic_salary,
      状态: d.status === 'active' ? '在职' : '离职',
    })));
    const wb = XLSX.utils.book_new();
    XLSX.utils.book_append_sheet(wb, ws, '人员列表');
    XLSX.writeFile(wb, '人员列表.xlsx');
  };

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 60 },
    {
      title: '姓名', key: 'name', width: 120,
      render: (_: any, r: any) => r.info?.name || r.name || '-',
    },
    { title: '考勤组', key: 'group', width: 100, render: (_: any, r: any) => r.info?.attendance_group || '-' },
    { title: '入职日期', key: 'entry', width: 110, render: (_: any, r: any) => r.info?.entry_date || '-' },
    { title: '基本工资', key: 'salary', width: 100, render: (_: any, r: any) => r.info?.basic_salary ? `¥${r.info.basic_salary}` : '-' },
    {
      title: '状态', dataIndex: 'status', width: 80,
      render: (s: string) => <Tag color={s === 'active' ? 'green' : 'red'}>{s === 'active' ? '在职' : '离职'}</Tag>,
    },
    {
      title: '操作', key: 'actions', width: 120, fixed: 'right' as const,
      render: (_: any, r: any) => (
        <Button type="link" icon={<EyeOutlined />} onClick={() => navigate(`/person/${r.id}`)}>详情</Button>
      ),
    },
  ];

  return (
    <div>
      <Space style={{ marginBottom: 16, justifyContent: 'space-between', width: '100%' }}>
        <Space>
          <Input prefix={<SearchOutlined />} placeholder="搜索姓名" value={keyword}
            onChange={e => { setKeyword(e.target.value); setPage(1); }} allowClear />
          <Button type="primary" icon={<SearchOutlined />} onClick={() => { setPage(1); fetchData(); }}>搜索</Button>
        </Space>
        <Space>
          <Button icon={<ExportOutlined />} onClick={handleExport}>导出</Button>
          {hasPermission('person', 'write') && (
            <Button type="primary" icon={<PlusOutlined />} onClick={() => setModalOpen(true)}>添加人员</Button>
          )}
        </Space>
      </Space>

      <Table rowKey="id" columns={columns} dataSource={data} loading={loading}
        pagination={{ current: page, total, pageSize: 20, onChange: setPage, showTotal: t => `共 ${t} 条` }}
        scroll={{ x: 900 }} size="middle" />

      <Modal title="添加人员" open={modalOpen} onOk={handleCreate} onCancel={() => { setModalOpen(false); form.resetFields(); }}
        width={720} okText="确定" cancelText="取消">
        <Form form={form} layout="vertical">
          <Row gutter={16}>
            <Col span={8}><Form.Item name="name" label="姓名" rules={[{ required: true }]}><Input /></Form.Item></Col>
            <Col span={8}><Form.Item name="attendance_group" label="考勤组"><Input /></Form.Item></Col>
            <Col span={8}><Form.Item name="entry_date" label="入职日期" rules={[{ required: true }]}><DatePicker style={{ width: '100%' }} /></Form.Item></Col>
            <Col span={6}><Form.Item name="basic_salary" label="基本工资"><InputNumber style={{ width: '100%' }} min={0} /></Form.Item></Col>
            <Col span={6}><Form.Item name="performance_salary" label="绩效工资"><InputNumber style={{ width: '100%' }} min={0} /></Form.Item></Col>
            <Col span={6}><Form.Item name="salary_days" label="计薪天数" initialValue={21.75}><InputNumber style={{ width: '100%' }} min={1} /></Form.Item></Col>
            <Col span={6}><Form.Item name="position_allowance" label="职位津贴"><InputNumber style={{ width: '100%' }} min={0} /></Form.Item></Col>
            <Col span={6}><Form.Item name="meal_subsidy" label="餐补"><InputNumber style={{ width: '100%' }} min={0} /></Form.Item></Col>
            <Col span={6}><Form.Item name="housing_subsidy" label="房补"><InputNumber style={{ width: '100%' }} min={0} /></Form.Item></Col>
            <Col span={6}><Form.Item name="transport_subsidy" label="交通补贴"><InputNumber style={{ width: '100%' }} min={0} /></Form.Item></Col>
            <Col span={6}><Form.Item name="heat_subsidy" label="高温补贴"><InputNumber style={{ width: '100%' }} min={0} /></Form.Item></Col>
            <Col span={6}><Form.Item name="social_insurance_deduct" label="社保代扣"><InputNumber style={{ width: '100%' }} min={0} /></Form.Item></Col>
            <Col span={6}><Form.Item name="housing_fund_deduct" label="公积金代扣"><InputNumber style={{ width: '100%' }} min={0} /></Form.Item></Col>
            <Col span={6}><Form.Item name="insurance_compensation" label="保险补偿"><InputNumber style={{ width: '100%' }} min={0} /></Form.Item></Col>
            <Col span={6}><Form.Item name="housing_fund_compensation" label="公积金补偿"><InputNumber style={{ width: '100%' }} min={0} /></Form.Item></Col>
            <Col span={6}><Form.Item name="gender" label="性别"><Select options={[{ label: '男', value: '男' }, { label: '女', value: '女' }]} /></Form.Item></Col>
            <Col span={6}><Form.Item name="birthday" label="生日"><DatePicker style={{ width: '100%' }} /></Form.Item></Col>
            <Col span={6}><Form.Item name="id_number" label="身份证号"><Input /></Form.Item></Col>
            <Col span={6}><Form.Item name="phones" label="电话"><Input /></Form.Item></Col>
            <Col span={6}><Form.Item name="emails" label="邮箱"><Input /></Form.Item></Col>
            <Col span={6}><Form.Item name="ethnicity" label="民族"><Input /></Form.Item></Col>
            <Col span={6}><Form.Item name="native_place" label="籍贯"><Input /></Form.Item></Col>
            <Col span={12}><Form.Item name="address" label="住址"><Input /></Form.Item></Col>
            <Col span={6}><Form.Item name="bank_cards" label="银行卡号"><Input /></Form.Item></Col>
            <Col span={6}><Form.Item name="political_status" label="政治面貌"><Input /></Form.Item></Col>
            <Col span={6}><Form.Item name="marital_status" label="婚姻状态"><Select options={[{ label: '未婚', value: '未婚' }, { label: '已婚', value: '已婚' }]} /></Form.Item></Col>
            <Col span={6}><Form.Item name="alias" label="别名"><Input /></Form.Item></Col>
          </Row>
        </Form>
      </Modal>
    </div>
  );
}
