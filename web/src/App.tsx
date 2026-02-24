import { useState } from 'react';
import { BrowserRouter, Routes, Route, Link, useLocation, useNavigate } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ConfigProvider, Dropdown, Drawer, Menu, theme } from 'antd';
import { DownOutlined, MenuOutlined } from '@ant-design/icons';
import type { MenuProps } from 'antd';
import Dashboard from './pages/Dashboard';
import Nodes from './pages/Nodes';
import Batches from './pages/Batches';
import BatchDetail from './pages/BatchDetail';
import Claims from './pages/Claims';
import Claimants from './pages/Claimants';
import Assets from './pages/Assets';
import Persons from './pages/Persons';
import Conflicts from './pages/Conflicts';
import Verify from './pages/Verify';

const queryClient = new QueryClient({
  defaultOptions: { queries: { refetchInterval: 15000, retry: 1 } },
});

function NavDropdown({ label, items }: { label: string; items: MenuProps['items'] }) {
  return (
    <Dropdown menu={{ items }} trigger={['hover']}>
      <span className="ant-dropdown-trigger">
        {label} <DownOutlined style={{ fontSize: 10 }} />
      </span>
    </Dropdown>
  );
}

function TopNavBar() {
  const navigate = useNavigate();
  const location = useLocation();
  const [drawerOpen, setDrawerOpen] = useState(false);

  const batchItems: MenuProps['items'] = [
    { key: 'all', label: 'All Batches', onClick: () => { navigate('/batches'); setDrawerOpen(false); } },
    { key: 'non-empty', label: 'Non-Empty', onClick: () => { navigate('/batches?filter=non-empty'); setDrawerOpen(false); } },
    { key: 'empty', label: 'Empty', onClick: () => { navigate('/batches?filter=empty'); setDrawerOpen(false); } },
    { key: 'pending', label: 'Pending', onClick: () => { navigate('/batches?filter=pending'); setDrawerOpen(false); } },
  ];

  const claimItems: MenuProps['items'] = [
    { key: 'all', label: 'All Claims', onClick: () => { navigate('/claims'); setDrawerOpen(false); } },
    { key: 'anchored', label: 'Anchored', onClick: () => { navigate('/claims?filter=anchored'); setDrawerOpen(false); } },
    { key: 'non-anchored', label: 'Non-Anchored', onClick: () => { navigate('/claims?filter=non-anchored'); setDrawerOpen(false); } },
    { key: 'published', label: 'Published', onClick: () => { navigate('/claims?filter=published'); setDrawerOpen(false); } },
    { key: 'claimant', label: 'Claimant', onClick: () => { navigate('/claimants'); setDrawerOpen(false); } },
  ];

  const publishItems: MenuProps['items'] = [
    { key: 'assets', label: 'Asset (AUID)', onClick: () => { navigate('/publish/assets'); setDrawerOpen(false); } },
    { key: 'persons', label: 'Person (PUID)', onClick: () => { navigate('/publish/persons'); setDrawerOpen(false); } },
    { key: 'conflicts', label: 'Conflicts', onClick: () => { navigate('/conflicts'); setDrawerOpen(false); } },
  ];

  const moreItems: MenuProps['items'] = [
    { key: 'nodes', label: 'Nodes', onClick: () => { navigate('/nodes'); setDrawerOpen(false); } },
    { key: 'about', label: 'About', onClick: () => { navigate('/about'); setDrawerOpen(false); } },
  ];

  const isActive = (path: string) => location.pathname === path || location.pathname.startsWith(path + '/');

  // Mobile drawer menu items
  const drawerMenuItems: MenuProps['items'] = [
    { key: '/', label: 'Home', onClick: () => { navigate('/'); setDrawerOpen(false); } },
    { key: '/batches', label: 'Batches', children: batchItems },
    { key: '/claims', label: 'Claims', children: claimItems },
    { key: '/publish', label: 'Publish', children: publishItems },
    { key: '/verify', label: 'Verify', onClick: () => { navigate('/verify'); setDrawerOpen(false); } },
    { key: '/more', label: 'More', children: moreItems },
  ];

  return (
    <>
      <nav className="otscan-navbar">
        <Link to="/" className="logo">
          <span>OT</span>Scan
        </Link>
        {/* Desktop menu */}
        <div className="otscan-nav-menu otscan-nav-desktop">
          <Link to="/" style={isActive('/') && location.pathname === '/' ? { color: '#1e88e5' } : undefined}>
            Home
          </Link>
          <NavDropdown label="Batches" items={batchItems} />
          <NavDropdown label="Claims" items={claimItems} />
          <NavDropdown label="Publish" items={publishItems} />
          <Link to="/verify" style={isActive('/verify') ? { color: '#1e88e5' } : undefined}>
            Verify
          </Link>
          <NavDropdown label="More" items={moreItems} />
        </div>
        {/* Mobile hamburger */}
        <div className="otscan-nav-mobile">
          <MenuOutlined
            style={{ fontSize: 20, cursor: 'pointer', color: '#333' }}
            onClick={() => setDrawerOpen(true)}
          />
        </div>
      </nav>
      <Drawer
        title="OTScan"
        placement="right"
        onClose={() => setDrawerOpen(false)}
        open={drawerOpen}
        width={280}
        styles={{ body: { padding: 0 } }}
      >
        <Menu
          mode="inline"
          items={drawerMenuItems}
          selectedKeys={[location.pathname]}
          style={{ border: 'none' }}
        />
      </Drawer>
    </>
  );
}

function AppLayout() {
  return (
    <div style={{ minHeight: '100vh', display: 'flex', flexDirection: 'column' }}>
      <TopNavBar />
      <div style={{ flex: 1 }}>
        <Routes>
          <Route path="/" element={<Dashboard />} />
          <Route path="/nodes" element={<Nodes />} />
          <Route path="/batches" element={<Batches />} />
          <Route path="/batches/:id" element={<BatchDetail />} />
          <Route path="/claims" element={<Claims />} />
          <Route path="/claimants" element={<Claimants />} />
          <Route path="/conflicts" element={<Conflicts />} />
          <Route path="/publish/assets" element={<Assets />} />
          <Route path="/publish/persons" element={<Persons />} />
          <Route path="/verify" element={<Verify />} />
        </Routes>
      </div>
      <footer className="otscan-footer">
        OTScan &copy; 2024 RMC &nbsp;|&nbsp;
        <Link to="/nodes">Nodes</Link>
        <Link to="/verify">Verify</Link>
      </footer>
    </div>
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
