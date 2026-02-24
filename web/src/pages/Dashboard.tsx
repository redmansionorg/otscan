import { useState, useCallback } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { Input, Tag, Badge, Spin, Alert, Row, Col, message } from 'antd';
import { SearchOutlined, BlockOutlined, TeamOutlined, SafetyCertificateOutlined, DatabaseOutlined, FileTextOutlined } from '@ant-design/icons';
import { Link, useNavigate } from 'react-router-dom';
import { AreaChart, Area, XAxis, YAxis, Tooltip, ResponsiveContainer } from 'recharts';
import { fetchDashboard, fetchClaimStats, fetchBatches, lookupHash, type DashboardData, type BatchSummary, type ClaimRecord, type ClaimStats, type BatchListResponse } from '../api/client';
import { useWebSocket, type WSEvent } from '../api/websocket';

const statusClass: Record<string, string> = {
  pending: 'status-pending', submitted: 'status-submitted', confirmed: 'status-confirmed',
  anchored: 'status-anchored', failed: 'status-failed',
};

const shortAddr = (addr?: string) => addr ? `${addr.slice(0, 6)}...${addr.slice(-4)}` : '-';

export default function Dashboard() {
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const [searchValue, setSearchValue] = useState('');
  const [searching, setSearching] = useState(false);

  // WebSocket for live updates - refresh all dashboard queries
  const handleWSEvent = useCallback((event: WSEvent) => {
    if (event.type === 'node_status' || event.type === 'batch_update' || event.type === 'claim_sync') {
      queryClient.invalidateQueries({ queryKey: ['dashboard'] });
      queryClient.invalidateQueries({ queryKey: ['claimStats'] });
      queryClient.invalidateQueries({ queryKey: ['trendBatches'] });
    }
  }, [queryClient]);

  const { connected } = useWebSocket(handleWSEvent);

  const { data, isLoading, error } = useQuery<DashboardData>({
    queryKey: ['dashboard'],
    queryFn: fetchDashboard,
  });

  const { data: claimStats } = useQuery<ClaimStats>({
    queryKey: ['claimStats'],
    queryFn: fetchClaimStats,
  });

  // Fetch last 16 batches for RUID Trends chart
  const { data: trendBatches } = useQuery<BatchListResponse>({
    queryKey: ['trendBatches'],
    queryFn: () => fetchBatches({ page: '1', pageSize: '16' }),
  });

  if (isLoading) return <Spin size="large" style={{ display: 'block', margin: '100px auto' }} />;
  if (error) return <Alert type="error" message="Failed to load dashboard" description={String(error)} />;
  if (!data) return null;

  // Smart search handler
  const handleSearch = async (value: string) => {
    const v = value.trim();
    if (!v) return;
    // Address format (0x + 40 hex) → claimant search
    if (/^0x[0-9a-fA-F]{40}$/.test(v)) {
      navigate(`/claims?claimant=${v}`);
      return;
    }
    // Batch ID format
    if (/^batch-/.test(v) || /^\d+$/.test(v)) {
      navigate(`/batches/${v}`);
      return;
    }
    // Hash format (0x + 64 hex) → lookup RUID/AUID/PUID via backend
    if (/^0x[0-9a-fA-F]{64}$/.test(v)) {
      setSearching(true);
      try {
        const result = await lookupHash(v);
        switch (result.type) {
          case 'ruid': navigate(`/verify?ruid=${v}`); break;
          case 'auid': navigate(`/claims?auid=${v}`); break;
          case 'puid': navigate(`/claims?puid=${v}`); break;
          default: message.warning('No matching RUID, AUID, or PUID found');
        }
      } catch {
        message.error('Search failed');
      } finally {
        setSearching(false);
      }
      return;
    }
    message.warning('Unrecognized format. Enter a RUID/AUID/PUID (0x + 64 hex), address (0x + 40 hex), or Batch ID.');
  };

  // RUID Trends chart data (reverse to show oldest→newest)
  const trendData = (trendBatches?.batches || [])
    .slice()
    .reverse()
    .map(b => ({
      name: b.onChainID ? `#${b.onChainID}` : b.batchID.slice(0, 8),
      ruids: b.ruidCount,
    }));

  const recentBatches = (data.recentBatches || []).slice(0, 5);
  const recentClaims = (data.recentClaims || []).slice(0, 5);

  return (
    <div>
      {/* Hero Section */}
      <div className="otscan-hero">
        <h1>OTScan - RMC Copyright Timestamp Explorer</h1>
        <div className="subtitle">Blockchain-based copyright timestamping and verification</div>
        <div className="search-box">
          <Input.Search
            placeholder="Search by RUID / AUID / PUID / Batch ID / Address"
            enterButton={<SearchOutlined />}
            size="large"
            value={searchValue}
            onChange={e => setSearchValue(e.target.value)}
            onSearch={handleSearch}
            loading={searching}
            style={{ borderRadius: 8 }}
          />
        </div>
        <div className="chain-stats">
          <span><BlockOutlined /> Block: {data.latestBlock?.toLocaleString()}</span>
          <span>|</span>
          <span><TeamOutlined /> Validators: {data.nodesHealthy}/{data.nodeCount}</span>
          <span>|</span>
          <span><SafetyCertificateOutlined /> BTC Anchored: {data.anchoredBatches ?? (data.totalBatches - data.pendingBatches)}</span>
          <span>|</span>
          <span>
            <Badge status={connected ? 'success' : 'error'} />
            {connected ? 'Live' : 'Reconnecting'}
          </span>
        </div>
      </div>

      <div className="otscan-content">
        {/* Statistics Row */}
        <Row gutter={[16, 16]} style={{ marginBottom: 20 }}>
          <Col xs={12} sm={12} md={6}>
            <div className="stat-card">
              <div className="stat-label">Total Batches</div>
              <div className="stat-value">{data.totalBatches}</div>
              <div className="stat-sub">
                {data.anchoredBatches ?? (data.totalBatches - data.pendingBatches)} anchored, {data.pendingBatches} pending
              </div>
            </div>
          </Col>
          <Col xs={12} sm={12} md={6}>
            <div className="stat-card">
              <div className="stat-label">Total Claims</div>
              <div className="stat-value">{data.totalClaims?.toLocaleString()}</div>
              <div className="stat-sub">
                {claimStats ? `${claimStats.publishedCount.toLocaleString()} published` : 'loading...'}
              </div>
            </div>
          </Col>
          <Col xs={12} sm={12} md={6}>
            <div className="stat-card">
              <div className="stat-label">Published / Total</div>
              <div className="stat-value">
                {claimStats ? `${claimStats.publishedCount.toLocaleString()} / ${data.totalClaims?.toLocaleString()}` : '-'}
              </div>
              <div className="stat-sub">
                {claimStats ? `${claimStats.uniqueAuids} AUIDs, ${claimStats.uniquePuids} PUIDs` : ''}
              </div>
            </div>
          </Col>
          <Col xs={12} sm={12} md={6}>
            <div className="trends-chart">
              <div className="chart-title">RUID Trends</div>
              <div className="chart-subtitle">
                Last {trendData.length} batches
              </div>
              {trendData.length > 0 ? (
                <ResponsiveContainer width="100%" height={80}>
                  <AreaChart data={trendData}>
                    <defs>
                      <linearGradient id="ruidGradient" x1="0" y1="0" x2="0" y2="1">
                        <stop offset="5%" stopColor="#3498db" stopOpacity={0.3} />
                        <stop offset="95%" stopColor="#3498db" stopOpacity={0} />
                      </linearGradient>
                    </defs>
                    <XAxis dataKey="name" hide />
                    <YAxis hide />
                    <Tooltip formatter={(v: number | undefined) => [`${(v ?? 0).toLocaleString()} RUIDs`, 'Count']} />
                    <Area type="monotone" dataKey="ruids" stroke="#3498db" fill="url(#ruidGradient)" strokeWidth={2} />
                  </AreaChart>
                </ResponsiveContainer>
              ) : (
                <div style={{ color: '#ccc', fontSize: 12, paddingTop: 16 }}>No batch data yet</div>
              )}
            </div>
          </Col>
        </Row>

        {/* Dual Column: Latest Batches + Latest Claims */}
        <Row gutter={[16, 16]} style={{ marginBottom: 20 }}>
          <Col xs={24} lg={12}>
            <div className="section-card">
              <div className="card-header">
                <span><DatabaseOutlined /> Latest Batches</span>
                <Link to="/batches">View All &rarr;</Link>
              </div>
              <div className="card-body">
                {recentBatches.map((b: BatchSummary) => (
                  <div className="latest-item" key={b.batchID}>
                    <div className="item-icon">
                      <DatabaseOutlined />
                    </div>
                    <div className="item-info">
                      <div className="item-title">
                        <Link to={`/batches/${b.batchID}`}>
                          {b.onChainID ? `#${b.onChainID}` : ''} {b.batchID.length > 24 ? b.batchID.slice(0, 24) + '...' : b.batchID}
                        </Link>
                      </div>
                      <div className="item-meta">
                        Blk {b.startBlock}-{b.endBlock} | {b.ruidCount.toLocaleString()} RUIDs
                        {b.anchoredBy ? ` | ${shortAddr(b.anchoredBy)}` : ''}
                      </div>
                    </div>
                    <div className="item-right">
                      <span className={`status-tag ${statusClass[b.status] || ''}`}>{b.status}</span>
                    </div>
                  </div>
                ))}
                {recentBatches.length === 0 && (
                  <div style={{ padding: 20, textAlign: 'center', color: '#999' }}>No batches yet</div>
                )}
              </div>
            </div>
          </Col>

          <Col xs={24} lg={12}>
            <div className="section-card">
              <div className="card-header">
                <span><FileTextOutlined /> Latest Claims</span>
                <Link to="/claims">View All &rarr;</Link>
              </div>
              <div className="card-body">
                {recentClaims.length > 0 ? recentClaims.map((c: ClaimRecord) => (
                  <div className="latest-item" key={c.ruid}>
                    <div className="item-icon">
                      <FileTextOutlined />
                    </div>
                    <div className="item-info">
                      <div className="item-title">
                        <Link to={`/verify?ruid=${c.ruid}`} title={c.ruid}>
                          {shortAddr(c.ruid)}
                        </Link>
                      </div>
                      <div className="item-meta">
                        Claimant {shortAddr(c.claimant)} | Block {c.submitBlock}
                      </div>
                    </div>
                    <div className="item-right">
                      <span className={`status-tag ${c.published ? 'status-confirmed' : 'status-pending'}`}>
                        {c.published ? 'Published' : 'Pending'}
                      </span>
                    </div>
                  </div>
                )) : (
                  <div style={{ padding: 20, textAlign: 'center', color: '#999' }}>No claims yet</div>
                )}
              </div>
            </div>
          </Col>
        </Row>

        {/* Node Status Bar */}
        <div className="node-status-bar">
          <span style={{ fontWeight: 600, fontSize: 13, color: '#21325b', marginRight: 8 }}>
            <TeamOutlined /> Nodes:
          </span>
          {data.nodes.map(n => (
            <span className="node-item" key={n.name}>
              <span className={`node-dot ${n.status}`} />
              <Link to="/nodes" className="explorer-link">{n.name}</Link>
              <Tag color={n.status === 'healthy' ? 'green' : n.status === 'degraded' ? 'orange' : 'red'} style={{ fontSize: 11, marginLeft: 2 }}>
                #{n.blockNumber}
              </Tag>
            </span>
          ))}
        </div>
      </div>
    </div>
  );
}
