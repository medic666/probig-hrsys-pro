import { useEffect, useState } from 'react';
import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import { Layout, Menu, Button, theme, Dropdown } from 'antd';
import {
  DashboardOutlined,
  TeamOutlined,
  BankOutlined,
  ClockCircleOutlined,
  DollarOutlined,
  FileOutlined,
  AuditOutlined,
  LogoutOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  UserOutlined,
} from '@ant-design/icons';
import { useAuth } from '../hooks/useAuth';

const { Header, Sider, Content } = Layout;

const menuItems = [
  { key: '/', icon: <DashboardOutlined />, label: '首页' },
  { key: '/person', icon: <TeamOutlined />, label: '人员管理', module: 'person' },
  { key: '/organization', icon: <BankOutlined />, label: '组织管理', module: 'organization' },
  { key: '/attendance', icon: <ClockCircleOutlined />, label: '假勤管理', module: 'attendance' },
  { key: '/salary', icon: <DollarOutlined />, label: '工资管理', module: 'salary' },
  { key: '/file', icon: <FileOutlined />, label: '文件管理', module: 'file' },
  { key: '/audit', icon: <AuditOutlined />, label: '操作审计', module: 'audit' },
];

export default function MainLayout() {
  const [collapsed, setCollapsed] = useState(false);
  const navigate = useNavigate();
  const location = useLocation();
  const { user, logout, hasPermission } = useAuth();
  const { token: themeToken } = theme.useToken();

  const visibleMenuItems = menuItems.filter((item) => {
    if (!item.module) return true;
    return hasPermission(item.module, 'read');
  });

  const selectedKey = '/' + (location.pathname.split('/')[1] || '');

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider trigger={null} collapsible collapsed={collapsed} theme="dark">
        <div style={{
          height: 64,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          color: '#fff',
          fontSize: collapsed ? 14 : 18,
          fontWeight: 'bold',
          borderBottom: '1px solid rgba(255,255,255,0.1)',
        }}>
          {collapsed ? 'HM' : '人事资产管理'}
        </div>
        <Menu
          theme="dark"
          mode="inline"
          selectedKeys={[selectedKey === '/' ? '/' : selectedKey]}
          items={visibleMenuItems}
          onClick={({ key }) => navigate(key)}
        />
      </Sider>
      <Layout>
        <Header style={{
          padding: '0 24px',
          background: themeToken.colorBgContainer,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          borderBottom: `1px solid ${themeToken.colorBorderSecondary}`,
        }}>
          <Button
            type="text"
            icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
            onClick={() => setCollapsed(!collapsed)}
          />
          <Dropdown menu={{
            items: [
              { key: 'user', label: `${user?.username} (${user?.role_id === 1 ? '管理员' : '用户'})`, disabled: true },
              { type: 'divider' },
              { key: 'logout', icon: <LogoutOutlined />, label: '退出登录', onClick: logout },
            ],
          }}>
            <Button type="text" icon={<UserOutlined />}>{user?.username}</Button>
          </Dropdown>
        </Header>
        <Content style={{
          margin: 24,
          padding: 24,
          background: themeToken.colorBgContainer,
          borderRadius: themeToken.borderRadiusLG,
          minHeight: 280,
          overflow: 'auto',
        }}>
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  );
}
