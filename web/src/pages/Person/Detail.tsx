import { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Card, Descriptions, Table, Button, Space, Tag, Modal, Tabs, message, Popconfirm } from 'antd';
import { ArrowLeftOutlined, PlusOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons';
import { get, post, put, del } from '../../services/api';
import { useAuth } from '../../hooks/useAuth';
import PersonEventForm from './EventForm';

export default function PersonDetail() {
  const { id } = useParams();
  const navigate = useNavigate();
  const { hasPermission } = useAuth();
  const [entity, setEntity] = useState<any>({});
  const [events, setEvents] = useState<any[]>([]);
  const [snapshots, setSnapshots] = useState<any[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const [editEvent, setEditEvent] = useState<any>(null);

  const fetchDetail = async () => {
    setLoading(true);
    try {
      const res = await get<any>(`/persons/${id}`);
      if (res.code === 0) {
        setEntity(res.data.entity || {});
        setEvents(res.data.events || []);
        setSnapshots(res.data.snapshots || []);
      }
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { fetchDetail(); }, [id]);

  const handleDeleteEvent = async (eventId: number) => {
    const res = await del<any>(`/person-events/${eventId}`);
    if (res.code === 0) {
      message.success('删除成功');
      fetchDetail();
    } else {
      message.error(res.message);
    }
  };

  const handleSubmit = async (values: any) => {
    const payload = {
      person_id: Number(id),
      effective_date: values.effective_date,
      event_type: values.event_type,
      data: values.data,
    };
    if (editEvent) {
      const res = await put<any>(`/person-events/${editEvent.id}`, payload);
      if (res.code === 0) { message.success('更新成功'); setModalOpen(false); setEditEvent(null); fetchDetail(); }
      else message.error(res.message);
    } else {
      const res = await post<any>('/person-events', payload);
      if (res.code === 0) { message.success('添加成功'); setModalOpen(false); fetchDetail(); }
      else message.error(res.message);
    }
  };

  const latestSnapshot = snapshots.length > 0 ? JSON.parse(snapshots[0].snapshot_data || '{}') : {};
  const eventTypeMap: Record<string, string> = {
    onboard: '入职', position_change: '调岗', info_change: '信息变更', offboard: '离职',
  };

  const eventColumns = [
    { title: 'ID', dataIndex: 'id', width: 60 },
    { title: '生效日期', dataIndex: 'effective_date', width: 110 },
    {
      title: '事件类型', dataIndex: 'event_type', width: 100,
      render: (t: string) => <Tag>{eventTypeMap[t] || t}</Tag>,
    },
    {
      title: '关键信息', key: 'info', ellipsis: true,
      render: (_: any, r: any) => {
        try {
          const d = JSON.parse(r.payload || '{}');
          return `姓名:${d.name || '-'} 基本工资:${d.basic_salary || '-'}`;
        } catch { return '-'; }
      },
    },
    {
      title: '操作', key: 'actions', width: 160,
      render: (_: any, r: any) => (
        <Space size="small">
          <Button type="link" size="small" icon={<EditOutlined />} onClick={() => {
            try {
              const d = JSON.parse(r.payload || '{}');
              setEditEvent({ ...r, data: d });
              setModalOpen(true);
            } catch { setEditEvent(r); setModalOpen(true); }
          }}>编辑</Button>
          <Popconfirm title="确定删除此事件?" onConfirm={() => handleDeleteEvent(r.id)}>
            <Button type="link" size="small" danger icon={<DeleteOutlined />}>删除</Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  const snapshotColumns = [
    { title: 'ID', dataIndex: 'id', width: 60 },
    { title: '生效日期', dataIndex: 'effective_date', width: 110 },
    {
      title: '快照信息', key: 'info', ellipsis: true,
      render: (_: any, r: any) => {
        try {
          const d = JSON.parse(r.snapshot_data || '{}');
          return `姓名:${d.name || '-'} 工资:${d.basic_salary || 0} 考勤组:${d.attendance_group || '-'}`;
        } catch { return '-'; }
      },
    },
  ];

  return (
    <div>
      <Space style={{ marginBottom: 16 }}>
        <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/person')}>返回列表</Button>
        {hasPermission('person', 'write') && (
          <Button type="primary" icon={<PlusOutlined />} onClick={() => { setEditEvent(null); setModalOpen(true); }}>
            添加事件
          </Button>
        )}
      </Space>

      <Card title="人员基本信息" style={{ marginBottom: 16 }}>
        <Descriptions bordered column={{ xs: 1, sm: 2, md: 3 }} size="small">
          <Descriptions.Item label="ID">{entity.id}</Descriptions.Item>
          <Descriptions.Item label="姓名">{latestSnapshot.name || entity.name || '-'}</Descriptions.Item>
          <Descriptions.Item label="别名">{latestSnapshot.alias || '-'}</Descriptions.Item>
          <Descriptions.Item label="性别">{latestSnapshot.gender || '-'}</Descriptions.Item>
          <Descriptions.Item label="生日">{latestSnapshot.birthday || '-'}</Descriptions.Item>
          <Descriptions.Item label="入职日期">{latestSnapshot.entry_date || '-'}</Descriptions.Item>
          <Descriptions.Item label="考勤组">{latestSnapshot.attendance_group || '-'}</Descriptions.Item>
          <Descriptions.Item label="基本工资">{latestSnapshot.basic_salary ? `¥${latestSnapshot.basic_salary}` : '-'}</Descriptions.Item>
          <Descriptions.Item label="绩效工资">{latestSnapshot.performance_salary ? `¥${latestSnapshot.performance_salary}` : '-'}</Descriptions.Item>
          <Descriptions.Item label="计薪天数">{latestSnapshot.salary_days || '-'}</Descriptions.Item>
          <Descriptions.Item label="职位津贴">{latestSnapshot.position_allowance || '-'}</Descriptions.Item>
          <Descriptions.Item label="餐补">{latestSnapshot.meal_subsidy || '-'}</Descriptions.Item>
          <Descriptions.Item label="房补">{latestSnapshot.housing_subsidy || '-'}</Descriptions.Item>
          <Descriptions.Item label="交通补贴">{latestSnapshot.transport_subsidy || '-'}</Descriptions.Item>
          <Descriptions.Item label="高温补贴">{latestSnapshot.heat_subsidy || '-'}</Descriptions.Item>
          <Descriptions.Item label="社保代扣">{latestSnapshot.social_insurance_deduct || '-'}</Descriptions.Item>
          <Descriptions.Item label="公积金代扣">{latestSnapshot.housing_fund_deduct || '-'}</Descriptions.Item>
          <Descriptions.Item label="身份证号">{latestSnapshot.id_number || '-'}</Descriptions.Item>
          <Descriptions.Item label="民族">{latestSnapshot.ethnicity || '-'}</Descriptions.Item>
          <Descriptions.Item label="籍贯">{latestSnapshot.native_place || '-'}</Descriptions.Item>
          <Descriptions.Item label="住址">{latestSnapshot.address || '-'}</Descriptions.Item>
          <Descriptions.Item label="政治面貌">{latestSnapshot.political_status || '-'}</Descriptions.Item>
          <Descriptions.Item label="婚姻状态">{latestSnapshot.marital_status || '-'}</Descriptions.Item>
          <Descriptions.Item label="状态">
            <Tag color={entity.status === 'active' ? 'green' : 'red'}>{entity.status === 'active' ? '在职' : '离职'}</Tag>
          </Descriptions.Item>
        </Descriptions>
      </Card>

      <Tabs defaultActiveKey="events" items={[
        {
          key: 'events', label: `事件记录 (${events.length})`,
          children: <Table rowKey="id" columns={eventColumns} dataSource={events} loading={loading}
            pagination={false} size="small" scroll={{ x: 700 }} />,
        },
        {
          key: 'snapshots', label: `历史快照 (${snapshots.length})`,
          children: <Table rowKey="id" columns={snapshotColumns} dataSource={snapshots} loading={loading}
            pagination={false} size="small" scroll={{ x: 400 }} />,
        },
      ]} />

      <Modal title={editEvent ? '编辑人员事件' : '添加人员事件'} open={modalOpen} width={800}
        onCancel={() => { setModalOpen(false); setEditEvent(null); }}
        footer={null} destroyOnClose>
        <PersonEventForm
          initialValues={editEvent ? {
            effective_date: editEvent.effective_date,
            event_type: editEvent.event_type,
            ...(editEvent.data || {}),
          } : undefined}
          onSubmit={handleSubmit}
          onCancel={() => { setModalOpen(false); setEditEvent(null); }}
        />
      </Modal>
    </div>
  );
}
