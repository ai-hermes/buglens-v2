import { Layout, Menu, Typography } from 'antd'
import { Link, Outlet, useLocation } from 'react-router-dom'

const { Header, Content } = Layout

export function AppLayout() {
  const location = useLocation()

  return (
    <Layout className="app-layout">
      <Header className="app-header">
        <Typography.Text className="app-brand">Demo FE Project</Typography.Text>
        <Menu
          mode="horizontal"
          selectedKeys={[location.pathname]}
          items={[{ key: '/home', label: <Link to="/home">首页</Link> }]}
        />
      </Header>
      <Content className="app-content">
        <Outlet />
      </Content>
    </Layout>
  )
}
