import { useEffect, useState } from 'react';
import { Table, Button, Space, Input, Upload, Modal, message, Popconfirm, Tag } from 'antd';
import { UploadOutlined, SearchOutlined, DeleteOutlined, DownloadOutlined } from '@ant-design/icons';
import type { UploadFile } from 'antd';
import { get, del } from '../../services/api';
import { useAuth } from '../../hooks/useAuth';

export default function FileManagement() {
  const [data, setData] = useState<any[]>([]);
  const [loading, setLoading] = useState(false);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [keyword, setKeyword] = useState('');
  const [uploadOpen, setUploadOpen] = useState(false);
  const { hasPermission } = useAuth();

  const fetchData = async () => {
    setLoading(true);
    try {
      const res = await get<any>('/files', { page, page_size: 20, keyword });
      if (res.code === 0) { setData(res.data.list || []); setTotal(res.data.total || 0); }
    } finally { setLoading(false); }
  };

  useEffect(() => { fetchData(); }, [page, keyword]);

  const handleUpload = async (info: any) => {
    if (info.file.status === 'done') {
      message.success(`${info.file.name} 上传成功`);
      setUploadOpen(false);
      fetchData();
    } else if (info.file.status === 'error') {
      message.error(`${info.file.name} 上传失败`);
    }
  };

  const handleDelete = async (fileId: number) => {
    const res = await del<any>(`/files/${fileId}`);
    if (res.code === 0) { message.success('删除成功'); fetchData(); }
    else message.error(res.message);
  };

  const formatSize = (size: number) => {
    if (size < 1024) return `${size} B`;
    if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)} KB`;
    return `${(size / (1024 * 1024)).toFixed(1)} MB`;
  };

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 60 },
    { title: '原始文件名', dataIndex: 'original_name', ellipsis: true },
    { title: '存储名', dataIndex: 'filename', width: 250, ellipsis: true },
    { title: '大小', dataIndex: 'size', width: 100, render: (s: number) => formatSize(s) },
    { title: '类型', dataIndex: 'mime_type', width: 180, ellipsis: true },
    { title: '上传时间', dataIndex: 'created_at', width: 170 },
    {
      title: '操作', key: 'actions', width: 180,
      render: (_: any, r: any) => (
        <Space size="small">
          <Button type="link" size="small" icon={<DownloadOutlined />}
            onClick={() => window.open(`/api/files/${r.id}/download`)}>下载</Button>
          {hasPermission('file', 'delete') && (
            <Popconfirm title="确定删除?" onConfirm={() => handleDelete(r.id)}>
              <Button type="link" size="small" danger icon={<DeleteOutlined />}>删除</Button>
            </Popconfirm>
          )}
        </Space>
      ),
    },
  ];

  return (
    <div>
      <Space style={{ marginBottom: 16, justifyContent: 'space-between', width: '100%' }}>
        <Space>
          <Input prefix={<SearchOutlined />} placeholder="搜索文件名" value={keyword}
            onChange={e => { setKeyword(e.target.value); setPage(1); }} allowClear />
          <Button type="primary" icon={<SearchOutlined />} onClick={() => { setPage(1); fetchData(); }}>搜索</Button>
        </Space>
        {hasPermission('file', 'write') && (
          <Button type="primary" icon={<UploadOutlined />} onClick={() => setUploadOpen(true)}>上传文件</Button>
        )}
      </Space>

      <Table rowKey="id" columns={columns} dataSource={data} loading={loading}
        pagination={{ current: page, total, pageSize: 20, onChange: setPage, showTotal: t => `共 ${t} 条` }}
        scroll={{ x: 1000 }} size="middle" />

      <Modal title="上传文件" open={uploadOpen}
        onCancel={() => setUploadOpen(false)} footer={null} destroyOnClose>
        <Upload.Dragger
          name="file"
          action="/api/files/upload"
          headers={{ Authorization: `Bearer ${localStorage.getItem('token')}` }}
          onChange={handleUpload}
          showUploadList={false}
        >
          <p className="ant-upload-drag-icon"><UploadOutlined style={{ fontSize: 48, color: '#1677ff' }} /></p>
          <p>点击或拖拽文件到此区域上传</p>
        </Upload.Dragger>
      </Modal>
    </div>
  );
}
