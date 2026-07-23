import { useEffect, useState } from 'react';
import { Table, Select, Space, Button, Modal, Descriptions } from 'antd';
import { SearchOutlined, EyeOutlined } from '@ant-design/icons';
import dayjs from 'dayjs';
import { get } from '../../services/api';

const entityTypeMap: Record<string, string> = {
  person: '人员',
  person_event: '人员事件',
  organization: '组织',
  org_event: '组织事件',
  attendance_event: '假勤事件',
  attendance_summary: '假勤汇总',
  salary_event: '工资事件',
  salary_summary: '工资汇总',
  file: '文件',
  file_association: '文件关联',
};

const actionMap: Record<string, string> = {
  create: '创建',
  update: '更新',
  delete: '删除',
  calculate: '核算',
};

function renderPayload(payloadStr: string) {
  try {
    const obj = JSON.parse(payloadStr);
    return Object.entries(obj).map(([key, val]) => (
      <Descriptions.Item key={key} label={key}>
        {typeof val === 'object' ? JSON.stringify(val) : String(val)}
      </Descriptions.Item>
    ));
  } catch {
    return <Descriptions.Item label="原始数据">{payloadStr}</Descriptions.Item>;
  }
}

export default function Audit() {
  const [data, setData] = useState<any[]>([]);
  const [loading, setLoading] = useState(false);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [entityType, setEntityType] = useState('');
  const [action, setAction] = useState('');
  const [detailOpen, setDetailOpen] = useState(false);
  const [detailRecord, setDetailRecord] = useState<any>(null);

  const fetchData = async () => {
    setLoading(true);
    try {
      const params: any = { page, page_size: 20 };
      if (entityType) params.entity_type = entityType;
      if (action) params.action = action;
      const res = await get<any>('/audit-logs', params);
      if (res.code === 0) { setData(res.data.list || []); setTotal(res.data.total || 0); }
    } finally { setLoading(false); }
  };

  useEffect(() => { fetchData(); }, [page, entityType, action]);

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 60 },
    { title: '操作人', dataIndex: 'username', width: 100, render: (v: string, r: any) => v || `用户${r.user_id}` },
    { title: '操作', dataIndex: 'action', width: 80, render: (a: string) => actionMap[a] || a },
    { title: '操作对象', dataIndex: 'entity_type', width: 110, render: (t: string) => entityTypeMap[t] || t },
    { title: '对象ID', dataIndex: 'entity_id', width: 80, render: (v: number) => v ?? '-' },
    {
      title: '操作详情', key: 'detail', width: 100,
      render: (_: any, r: any) => (
        <Button type="link" size="small" icon={<EyeOutlined />} onClick={() => { setDetailRecord(r); setDetailOpen(true); }}>
          查看
        </Button>
      ),
    },
    { title: 'IP', dataIndex: 'ip_address', width: 130 },
    { title: '操作时间', dataIndex: 'created_at', width: 170, render: (t: string) => dayjs(t).format('YYYY-MM-DD HH:mm:ss') },
  ];

  return (
    <div>
      <Space style={{ marginBottom: 16 }} wrap>
        <Select placeholder="操作类型" value={action || undefined} onChange={(v) => { setAction(v || ''); setPage(1); }}
          allowClear style={{ width: 120 }} options={[
            { label: '创建', value: 'create' }, { label: '更新', value: 'update' }, { label: '删除', value: 'delete' },
            { label: '核算', value: 'calculate' },
          ]} />
        <Select placeholder="操作对象" value={entityType || undefined} onChange={(v) => { setEntityType(v || ''); setPage(1); }}
          showSearch allowClear style={{ width: 160 }}
          options={Object.entries(entityTypeMap).map(([k, v]) => ({ label: v, value: k }))}
          filterOption={(input, option) => (option?.label as string || '').includes(input)}
        />
        <Button type="primary" icon={<SearchOutlined />} onClick={() => { setPage(1); fetchData(); }}>搜索</Button>
      </Space>

      <Table rowKey="id" columns={columns} dataSource={data} loading={loading}
        pagination={{ current: page, total, pageSize: 20, onChange: setPage, showTotal: t => `共 ${t} 条` }}
        scroll={{ x: 1000 }} size="small" />

      <Modal title="操作详情" open={detailOpen} width={700}
        onCancel={() => { setDetailOpen(false); setDetailRecord(null); }}
        footer={<Button onClick={() => { setDetailOpen(false); setDetailRecord(null); }}>关闭</Button>}>
        {detailRecord && (
          <Descriptions bordered column={1} size="small">
            <Descriptions.Item label="操作人">{detailRecord.username || `用户${detailRecord.user_id}`}</Descriptions.Item>
            <Descriptions.Item label="操作">{actionMap[detailRecord.action] || detailRecord.action}</Descriptions.Item>
            <Descriptions.Item label="操作对象">{entityTypeMap[detailRecord.entity_type] || detailRecord.entity_type}</Descriptions.Item>
            <Descriptions.Item label="对象ID">{detailRecord.entity_id ?? '-'}</Descriptions.Item>
            <Descriptions.Item label="IP地址">{detailRecord.ip_address}</Descriptions.Item>
            <Descriptions.Item label="时间">{dayjs(detailRecord.created_at).format('YYYY-MM-DD HH:mm:ss')}</Descriptions.Item>
          </Descriptions>
        )}
        {detailRecord && (
          <Descriptions bordered column={1} size="small" style={{ marginTop: 16 }}
            title="变更内容">
            {renderPayload(detailRecord.payload)}
          </Descriptions>
        )}
      </Modal>
    </div>
  );
}
