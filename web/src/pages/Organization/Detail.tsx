import { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Card, Descriptions, Table, Button, Space, Tag, Modal, Tabs, message, Popconfirm, Form, Input, Select, DatePicker, Row, Col } from 'antd';
import { ArrowLeftOutlined, PlusOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons';
import dayjs from 'dayjs';
import { get, post, put, del } from '../../services/api';
import { useAuth } from '../../hooks/useAuth';

export default function OrganizationDetail() {
  const { id } = useParams();
  const navigate = useNavigate();
  const { hasPermission } = useAuth();
  const [entity, setEntity] = useState<any>({});
  const [events, setEvents] = useState<any[]>([]);
  const [snapshots, setSnapshots] = useState<any[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const [editEvent, setEditEvent] = useState<any>(null);
  const [form] = Form.useForm();

  const fetchDetail = async () => {
    setLoading(true);
    try {
      const res = await get<any>(`/organizations/${id}`);
      if (res.code === 0) {
        setEntity(res.data.entity || {});
        setEvents(res.data.events || []);
        setSnapshots(res.data.snapshots || []);
      }
    } finally { setLoading(false); }
  };

  useEffect(() => { fetchDetail(); }, [id]);

  const handleDeleteEvent = async (eventId: number) => {
    const res = await del<any>(`/org-events/${eventId}`);
    if (res.code === 0) { message.success('删除成功'); fetchDetail(); }
    else message.error(res.message);
  };

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      const data: any = {};
      ['company_name', 'credit_code', 'address', 'contact_phone', 'bank_name', 'bank_account'].forEach(f => {
        data[f] = values[f] || '';
      });
      const payload = {
        org_id: Number(id),
        effective_date: values.effective_date ? dayjs(values.effective_date).format('YYYY-MM-DD') : dayjs().format('YYYY-MM-DD'),
        event_type: values.event_type || 'info_change',
        data,
      };
      if (editEvent) {
        const res = await put<any>(`/org-events/${editEvent.id}`, payload);
        if (res.code === 0) { message.success('更新成功'); setModalOpen(false); setEditEvent(null); fetchDetail(); }
        else message.error(res.message);
      } else {
        const res = await post<any>('/org-events', payload);
        if (res.code === 0) { message.success('添加成功'); setModalOpen(false); fetchDetail(); }
        else message.error(res.message);
      }
    } catch { /* ignore */ }
  };

  const openForm = (evt?: any) => {
    if (evt) {
      try {
        const d = JSON.parse(evt.payload || '{}');
        setEditEvent(evt);
        form.setFieldsValue({ ...evt, ...d, effective_date: evt.effective_date ? dayjs(evt.effective_date) : undefined });
      } catch { setEditEvent(evt); form.setFieldsValue(evt); }
    } else {
      setEditEvent(null);
      form.resetFields();
    }
    setModalOpen(true);
  };

  const latestSnapshot = snapshots.length > 0 ? JSON.parse(snapshots[0].snapshot_data || '{}') : {};
  const eventColumns = [
    { title: 'ID', dataIndex: 'id', width: 60 },
    { title: '生效日期', dataIndex: 'effective_date', width: 110 },
    { title: '事件类型', dataIndex: 'event_type', width: 100, render: (t: string) => <Tag>{t === 'establish' ? '成立' : t === 'info_change' ? '变更' : t === 'dissolve' ? '解散' : t}</Tag> },
    { title: '关键信息', key: 'info', ellipsis: true, render: (_: any, r: any) => {
      try { const d = JSON.parse(r.payload || '{}'); return `公司:${d.company_name || '-'}`; } catch { return '-'; }
    }},
    { title: '操作', key: 'actions', width: 160, render: (_: any, r: any) => (
      <Space size="small">
        <Button type="link" size="small" icon={<EditOutlined />} onClick={() => openForm(r)}>编辑</Button>
        <Popconfirm title="确定删除?" onConfirm={() => handleDeleteEvent(r.id)}>
          <Button type="link" size="small" danger icon={<DeleteOutlined />}>删除</Button>
        </Popconfirm>
      </Space>
    )},
  ];

  const snapshotColumns = [
    { title: 'ID', dataIndex: 'id', width: 60 },
    { title: '生效日期', dataIndex: 'effective_date', width: 110 },
    { title: '快照信息', key: 'info', ellipsis: true, render: (_: any, r: any) => {
      try { const d = JSON.parse(r.snapshot_data || '{}'); return `公司:${d.company_name || '-'}`; } catch { return '-'; }
    }},
  ];

  return (
    <div>
      <Space style={{ marginBottom: 16 }}>
        <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/organization')}>返回列表</Button>
        {hasPermission('organization', 'write') && (
          <Button type="primary" icon={<PlusOutlined />} onClick={() => openForm()}>添加事件</Button>
        )}
      </Space>

      <Card title="组织基本资料" style={{ marginBottom: 16 }}>
        <Descriptions bordered column={{ xs: 1, sm: 2, md: 2 }} size="small">
          <Descriptions.Item label="ID">{entity.id}</Descriptions.Item>
          <Descriptions.Item label="公司名称">{latestSnapshot.company_name || entity.name || '-'}</Descriptions.Item>
          <Descriptions.Item label="统一社会信用代码">{latestSnapshot.credit_code || '-'}</Descriptions.Item>
          <Descriptions.Item label="地址">{latestSnapshot.address || '-'}</Descriptions.Item>
          <Descriptions.Item label="联系电话">{latestSnapshot.contact_phone || '-'}</Descriptions.Item>
          <Descriptions.Item label="开户行">{latestSnapshot.bank_name || '-'}</Descriptions.Item>
          <Descriptions.Item label="银行账号">{latestSnapshot.bank_account || '-'}</Descriptions.Item>
        </Descriptions>
      </Card>

      <Tabs defaultActiveKey="events" items={[
        { key: 'events', label: `事件记录 (${events.length})`,
          children: <Table rowKey="id" columns={eventColumns} dataSource={events} loading={loading} pagination={false} size="small" scroll={{ x: 700 }} /> },
        { key: 'snapshots', label: `历史快照 (${snapshots.length})`,
          children: <Table rowKey="id" columns={snapshotColumns} dataSource={snapshots} loading={loading} pagination={false} size="small" scroll={{ x: 400 }} /> },
      ]} />

      <Modal title={editEvent ? '编辑组织事件' : '添加组织事件'} open={modalOpen} width={600}
        onCancel={() => { setModalOpen(false); setEditEvent(null); }} onOk={handleSubmit}
        okText="确定" cancelText="取消" destroyOnClose>
        <Form form={form} layout="vertical">
          <Row gutter={16}>
            <Col span={12}><Form.Item name="effective_date" label="生效日期" rules={[{ required: true }]}><DatePicker style={{ width: '100%' }} /></Form.Item></Col>
            <Col span={12}><Form.Item name="event_type" label="事件类型" rules={[{ required: true }]}><Select options={[{ label: '信息变更', value: 'info_change' }, { label: '解散', value: 'dissolve' }]} /></Form.Item></Col>
          </Row>
          <Form.Item name="company_name" label="公司名称"><Input /></Form.Item>
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
