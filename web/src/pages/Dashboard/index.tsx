import { useEffect, useState } from 'react';
import { Row, Col, Card, Statistic } from 'antd';
import {
  TeamOutlined,
  BankOutlined,
  ClockCircleOutlined,
  DollarOutlined,
} from '@ant-design/icons';
import { get } from '../../services/api';

export default function Dashboard() {
  const [stats, setStats] = useState({ persons: 0, orgs: 0, attendanceEvents: 0, salaryEvents: 0 });

  useEffect(() => {
    (async () => {
      try {
        const [pRes, oRes, aRes, sRes] = await Promise.all([
          get<any>('/persons', { page: 1, page_size: 1 }),
          get<any>('/organizations', { page: 1, page_size: 1 }),
          get<any>('/attendance-events', { page: 1, page_size: 1 }),
          get<any>('/salary-events', { page: 1, page_size: 1 }),
        ]);
        setStats({
          persons: pRes.data?.total || 0,
          orgs: oRes.data?.total || 0,
          attendanceEvents: aRes.data?.total || 0,
          salaryEvents: sRes.data?.total || 0,
        });
      } catch { /* ignore */ }
    })();
  }, []);

  return (
    <div>
      <h2 style={{ marginBottom: 24 }}>系统概览</h2>
      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic title="人员数量" value={stats.persons} prefix={<TeamOutlined />} />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic title="组织数量" value={stats.orgs} prefix={<BankOutlined />} />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic title="假勤事件" value={stats.attendanceEvents} prefix={<ClockCircleOutlined />} suffix="条" />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic title="工资事件" value={stats.salaryEvents} prefix={<DollarOutlined />} suffix="条" />
          </Card>
        </Col>
      </Row>
    </div>
  );
}
