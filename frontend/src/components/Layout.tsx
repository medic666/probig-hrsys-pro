import { useState, useEffect } from 'react'
import { Layout, Menu, Button, theme, Avatar, Dropdown } from 'antd'
import {
  DashboardOutlined, UserOutlined, FileTextOutlined, ClockCircleOutlined,
  DollarOutlined, DatabaseOutlined, AuditOutlined, LogoutOutlined, MenuFoldOutlined, MenuUnfoldOutlined
} from '@ant-design/icons'
import { useNavigate, useLocation } from 'react-router-dom'
import { getMenus } from '../api/auth'

const { Header, Sider, Content } = Layout

const iconMap: Record<string, React.ReactNode> = {
  DashboardOutlined: <DashboardOutlined />,
  UserOutlined: <UserOutlined />,
  FileTextOutlined: <FileTextOutlined />,
  ClockCircleOutlined: <ClockCircleOutlined />,
  DollarOutlined: <DollarOutlined />,
  DatabaseOutlined: <DatabaseOutlined />,
  AuditOutlined: <AuditOutlined />,
}

export default function AppLayout({ children }: { children: React.ReactNode }) {
  const [collapsed, setCollapsed] = useState(false)
  const [menus, setMenus] = useState<any[]>([])
  const navigate = useNavigate()
  const location = useLocation()
  const { token: themeToken } = theme.useToken()

  useEffect(() => {
    getMenus().then((res) => setMenus(res.data || [])).catch(() => {})
  }, [])

  const handleLogout = () => {
    localStorage.removeItem('token')
    localStorage.removeItem('user')
    navigate('/login')
  }

  const user = JSON.parse(localStorage.getItem('user') || '{}')

  const menuItems = menus.map((m) => ({
    key: m.path,
    icon: iconMap[m.icon] || null,
    label: m.label,
  }))

  const selectedKey = '/' + (location.pathname.split('/')[1] || '')

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider trigger={null} collapsible collapsed={collapsed} theme="dark">
        <div style={{ height: 48, display: 'flex', alignItems: 'center', justifyContent: 'center', color: '#fff', fontWeight: 'bold', fontSize: collapsed ? 14 : 16 }}>
          {collapsed ? 'HR' : '人事资产管理系统'}
        </div>
        <Menu
          theme="dark"
          mode="inline"
          selectedKeys={[selectedKey]}
          items={menuItems}
          onClick={({ key }) => navigate(key)}
        />
      </Sider>
      <Layout>
        <Header style={{ padding: '0 24px', background: themeToken.colorBgContainer, display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <Button type="text" icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />} onClick={() => setCollapsed(!collapsed)} />
          <Dropdown menu={{ items: [{ key: 'logout', icon: <LogoutOutlined />, label: '退出登录', onClick: handleLogout }] }}>
            <div style={{ cursor: 'pointer', display: 'flex', alignItems: 'center', gap: 8 }}>
              <Avatar icon={<UserOutlined />} />
              <span>{user.username || '管理员'}</span>
            </div>
          </Dropdown>
        </Header>
        <Content style={{ margin: 16, padding: 24, background: themeToken.colorBgContainer, borderRadius: 8, overflow: 'auto' }}>
          {children}
        </Content>
      </Layout>
    </Layout>
  )
}
