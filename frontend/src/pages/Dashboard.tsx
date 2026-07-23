import { useState, useEffect } from 'react'
import { Row, Col, Card, Statistic } from 'antd'
import { UserOutlined, AuditOutlined, DatabaseOutlined } from '@ant-design/icons'
import { getDashboardStats } from '../api/auth'

export default function Dashboard() {
  const [stats, setStats] = useState({ personCount: 0, eventCount: 0, assetCount: 0 })

  useEffect(() => {
    getDashboardStats().then((res: any) => setStats(res.data)).catch(() => {})
  }, [])

  return (
    <div>
      <h2 style={{ marginBottom: 24 }}>仪表盘</h2>
      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} lg={8}>
          <Card>
            <Statistic title="人员总数" value={stats.personCount} prefix={<UserOutlined />} />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={8}>
          <Card>
            <Statistic title="审计事件" value={stats.eventCount} prefix={<AuditOutlined />} />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={8}>
          <Card>
            <Statistic title="资产总数" value={stats.assetCount} prefix={<DatabaseOutlined />} />
          </Card>
        </Col>
      </Row>
    </div>
  )
}
