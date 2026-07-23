import { useEffect, useState } from 'react';
import { Table, Button, Space, Input, Modal, Form, DatePicker, message, Row, Col } from 'antd';
import { PlusOutlined, SearchOutlined, EyeOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import dayjs from 'dayjs';
import { get, post } from '../../services/api';
import { useAuth } from '../../hooks/useAuth';

export default function OrganizationList() {
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
      const res = await get<any>('/organizations', { page, page_size: 20, keyword });
      if (res.code === 0) { setData(res.data.list || []); setTotal(res.data.total || 0); }
    } finally { setLoading(false); }
  };

  useEffect(() => { fetchData(); }, [page, keyword]);

  const handleCreate = async () => {
    try {
      const values = await form.validateFields();
      const payload = {
        org_id: 0,
        effective_date: values.effective_date ? dayjs(values.effective_date).format('YYYY-MM-DD') : dayjs().format('YYYY-MM-DD'),
        event_type: 'establish',
        data: {
          company_name: values.company_name,
          credit_code: values.credit_code || '',
          address: values.address || '',
          contact_phone: values.contact_phone || '',
          bank_name: values.bank_name || '',
          bank_account: values.bank_account || '',
        },
      };
      const res = await post<any>('/organizations', payload);
      if (res.code === 0) { message.success('添加成功'); setModalOpen(false); form.resetFields(); fetchData(); }
      else message.error(res.message);
    } catch { /* ignore */ }
  };

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 60 },
    { title: '公司名称', key: 'name', render: (_: any, r: any) => r.info?.company_name || r.name || '-' },
    { title: '信用代码', key: 'code', render: (_: any, r: any) => r.info?.credit_code || '-' },
    { title: '联系电话', key: 'phone', render: (_: any, r: any) => r.info?.contact_phone || '-' },
    { title: '开户行', key: 'bank', render: (_: any, r: any) => r.info?.bank_name || '-' },
    { title: '银行账号', key: 'account', render: (_: any, r: any) => r.info?.bank_account || '-' },
    { title: '操作', key: 'actions', width: 100, render: (_: any, r: any) => (
      <Button type="link" icon={<EyeOutlined />} onClick={() => navigate(`/organization/${r.id}`)}>详情</Button>
    )},
  ];

  return (
    <div>
      <Space style={{ marginBottom: 16, justifyContent: 'space-between', width: '100%' }}>
        <Space>
          <Input prefix={<SearchOutlined />} placeholder="搜索组织" value={keyword}
            onChange={e => { setKeyword(e.target.value); setPage(1); }} allowClear />
          <Button type="primary" icon={<SearchOutlined />} onClick={() => { setPage(1); fetchData(); }}>搜索</Button>
        </Space>
        {hasPermission('organization', 'write') && (
          <Button type="primary" icon={<PlusOutlined />} onClick={() => setModalOpen(true)}>添加组织</Button>
        )}
      </Space>

      <Table rowKey="id" columns={columns} dataSource={data} loading={loading}
        pagination={{ current: page, total, pageSize: 20, onChange: setPage, showTotal: t => `共 ${t} 条` }}
        scroll={{ x: 900 }} size="middle" />

      <Modal title="添加组织" open={modalOpen} onOk={handleCreate}
        onCancel={() => { setModalOpen(false); form.resetFields(); }} okText="确定" cancelText="取消">
        <Form form={form} layout="vertical">
          <Form.Item name="company_name" label="公司名称" rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="effective_date" label="生效日期" rules={[{ required: true }]}><DatePicker style={{ width: '100%' }} /></Form.Item>
          <Row gutter={16}>
            <Col span={12}><Form.Item name="credit_code" label="统一社会信用代码"><Input /></Form.Item></Col>
            <Col span={12}><Form.Item name="contact_phone" label="联系电话"><Input /></Form.Item></Col>
          </Row>
          <Form.Item name="address" label="地址"><Input /></Form.Item>
          <Row gutter={16}>
            <Col span={12}><Form.Item name="bank_name" label="开户行"><Input /></Form.Item></Col>
            <Col span={12}><Form.Item name="bank_account" label="银行账号"><Input /></Form.Item></Col>
          </Row>
        </Form>
      </Modal>
    </div>
  );
}
