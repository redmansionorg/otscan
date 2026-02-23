import { BrowserRouter, Routes, Route, Link, useLocation } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ConfigProvider, Layout, Menu, theme } from 'antd';
import {
  DashboardOutlined,
  ClusterOutlined,
  DatabaseOutlined,
  FileSearchOutlined,
  SafetyCertificateOutlined,
  WarningOutlined,
} from '@ant-design/icons';
import Dashboard from './pages/Dashboard';
import Nodes from './pages/Nodes';
import Batches from './pages/Batches';
import BatchDetail from './pages/BatchDetail';
import Claims from './pages/Claims';
import Conflicts from './pages/Conflicts';
import Verify from './pages/Verify';

const { Header, Sider, Content } = Layout;
const queryClient = new QueryClient({
  defaultOptions: { queries: { refetchInterval: 15000, retry: 1 } },
});

const menuItems = [
  { key: '/', icon: <DashboardOutlined />, label: <Link to="/">Dashboard</Link> },
  { key: '/nodes', icon: <ClusterOutlined />, label: <Link to="/nodes">Nodes</Link> },
  { key: '/batches', icon: <DatabaseOutlined />, label: <Link to="/batches">Batches</Link> },
  { key: '/claims', icon: <FileSearchOutlined />, label: <Link to="/claims">Claims</Link> },
  { key: '/conflicts', icon: <WarningOutlined />, label: <Link to="/conflicts">Conflicts</Link> },
  { key: '/verify', icon: <SafetyCertificateOutlined />, label: <Link to="/verify">Verify</Link> },
];

function AppLayout() {
  const location = useLocation();
  const selectedKey = '/' + (location.pathname.split('/')[1] || '');

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider breakpoint="lg" collapsedWidth={60} theme="dark">
        <div style={{ height: 48, margin: 16, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
          <span style={{ color: '#fff', fontSize: 18, fontWeight: 700, letterSpacing: 2 }}>OTScan</span>
        </div>
        <Menu
          theme="dark"
          mode="inline"
          selectedKeys={[selectedKey]}
          items={menuItems}
        />
      </Sider>
      <Layout>
        <Header style={{ background: '#fff', padding: '0 24px', display: 'flex', alignItems: 'center', borderBottom: '1px solid #f0f0f0' }}>
          <span style={{ fontSize: 16, fontWeight: 500 }}>OTS Consensus Explorer</span>
          <span style={{ marginLeft: 'auto', color: '#999', fontSize: 13 }}>RMC Chain (ID: 192)</span>
        </Header>
        <Content style={{ margin: 24, minHeight: 280 }}>
          <Routes>
            <Route path="/" element={<Dashboard />} />
            <Route path="/nodes" element={<Nodes />} />
            <Route path="/batches" element={<Batches />} />
            <Route path="/batches/:id" element={<BatchDetail />} />
            <Route path="/claims" element={<Claims />} />
            <Route path="/conflicts" element={<Conflicts />} />
            <Route path="/verify" element={<Verify />} />
          </Routes>
        </Content>
      </Layout>
    </Layout>
  );
}

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <ConfigProvider theme={{ algorithm: theme.defaultAlgorithm }}>
        <BrowserRouter>
          <AppLayout />
        </BrowserRouter>
      </ConfigProvider>
    </QueryClientProvider>
  );
}
