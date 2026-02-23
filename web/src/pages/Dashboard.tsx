import { useState, useCallback } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { Card, Col, Row, Statistic, Table, Tag, Badge, Spin, Alert, Select, Typography } from 'antd';
import { CheckCircleOutlined, CloseCircleOutlined, SyncOutlined, ClockCircleOutlined, WifiOutlined } from '@ant-design/icons';
import { Link } from 'react-router-dom';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Legend } from 'recharts';
import { fetchDashboard, fetchNodeHistory, fetchClaimStats, type DashboardData, type BatchSummary, type NodeHistoryPoint, type ClaimStats } from '../api/client';
import { useWebSocket, type WSEvent } from '../api/websocket';
import dayjs from 'dayjs';

const { Text } = Typography;

const statusColor: Record<string, string> = {
  pending: 'blue', submitted: 'orange', confirmed: 'green', anchored: 'gold', failed: 'red',
};

const shortAddr = (addr?: string) => addr ? `${addr.slice(0, 6)}...${addr.slice(-4)}` : '-';

const batchColumns = [
  {
    title: 'ID', dataIndex: 'onChainID', key: 'id', width: 60,
    render: (v: number, r: BatchSummary) => <Link to={`/batches/${r.batchID}`}>{v || '-'}</Link>,
  },
  { title: 'Batch ID', dataIndex: 'batchID', key: 'batchID', ellipsis: true, width: 180 },
  {
    title: 'Block Range', key: 'range', width: 140,
    render: (_: unknown, r: BatchSummary) => `${r.startBlock} - ${r.endBlock}`,
  },
  { title: 'RUIDs', dataIndex: 'ruidCount', key: 'ruidCount', width: 70 },
  {
    title: 'Status', dataIndex: 'status', key: 'status', width: 100,
    render: (s: string) => <Tag color={statusColor[s] || 'default'}>{s}</Tag>,
  },
  {
    title: 'Anchor Node', dataIndex: 'anchoredBy', key: 'anchoredBy', width: 130,
    render: (v: string) => v ? <span title={v}>{shortAddr(v)}</span> : '-',
  },
];

const NODE_COLORS = ['#1890ff', '#52c41a', '#fa8c16', '#eb2f96', '#722ed1'];

export default function Dashboard() {
  const queryClient = useQueryClient();
  const [selectedNode, setSelectedNode] = useState<string>('');

  // WebSocket for live updates
  const handleWSEvent = useCallback((event: WSEvent) => {
    if (event.type === 'node_status' || event.type === 'batch_update') {
      queryClient.invalidateQueries({ queryKey: ['dashboard'] });
    }
  }, [queryClient]);

  const { connected } = useWebSocket(handleWSEvent);

  const { data, isLoading, error } = useQuery<DashboardData>({
    queryKey: ['dashboard'],
    queryFn: fetchDashboard,
  });

  // Fetch node history for trend chart
  const historyNode = selectedNode || (data?.nodes?.[0]?.name ?? '');
  const { data: history } = useQuery<NodeHistoryPoint[]>({
    queryKey: ['nodeHistory', historyNode],
    queryFn: () => fetchNodeHistory(historyNode, 180),
    enabled: !!historyNode,
    refetchInterval: 30000,
  });

  const { data: claimStats } = useQuery<ClaimStats>({
    queryKey: ['claimStats'],
    queryFn: fetchClaimStats,
  });

  if (isLoading) return <Spin size="large" style={{ display: 'block', margin: '100px auto' }} />;
  if (error) return <Alert type="error" message="Failed to load dashboard" description={String(error)} />;
  if (!data) return null;

  // Process history data for chart
  const chartData = (history || []).map(h => ({
    time: dayjs(h.recordedAt).format('HH:mm:ss'),
    blockNumber: h.blockNumber,
    pendingCount: h.pendingCount,
    status: h.status === 'healthy' ? 1 : 0,
  }));

  return (
    <div>
      {/* Live connection indicator */}
      <div style={{ display: 'flex', alignItems: 'center', marginBottom: 16, gap: 8 }}>
        <Badge status={connected ? 'success' : 'error'} />
        <Text type="secondary" style={{ fontSize: 12 }}>
          <WifiOutlined /> {connected ? 'Live updates connected' : 'Reconnecting...'}
        </Text>
      </div>

      <Row gutter={[16, 16]}>
        <Col xs={12} sm={6}>
          <Card>
            <Statistic
              title="Nodes"
              value={data.nodesHealthy}
              suffix={`/ ${data.nodeCount}`}
              valueStyle={{ color: data.nodesHealthy === data.nodeCount ? '#3f8600' : '#cf1322' }}
              prefix={data.nodesHealthy === data.nodeCount ? <CheckCircleOutlined /> : <CloseCircleOutlined />}
            />
          </Card>
        </Col>
        <Col xs={12} sm={6}>
          <Card><Statistic title="Latest Block" value={data.latestBlock} prefix={<SyncOutlined />} /></Card>
        </Col>
        <Col xs={12} sm={6}>
          <Card><Statistic title="Total Batches" value={data.totalBatches} prefix={<ClockCircleOutlined />} /></Card>
        </Col>
        <Col xs={12} sm={6}>
          <Card><Statistic title="Total Claims" value={data.totalClaims} /></Card>
        </Col>
      </Row>

      {claimStats && (
        <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
          <Col xs={12} sm={6}>
            <Card><Statistic title="Published Claims" value={claimStats.publishedCount} /></Card>
          </Col>
          <Col xs={12} sm={6}>
            <Card><Statistic title="Unique AUIDs" value={claimStats.uniqueAuids} /></Card>
          </Col>
          <Col xs={12} sm={6}>
            <Card><Statistic title="Unique PUIDs" value={claimStats.uniquePuids} /></Card>
          </Col>
          <Col xs={12} sm={6}>
            <Card>
              <Statistic
                title="Conflict AUIDs"
                value={claimStats.conflictAuids}
                valueStyle={claimStats.conflictAuids > 0 ? { color: '#cf1322' } : undefined}
              />
            </Card>
          </Col>
        </Row>
      )}

      {/* Trend Chart */}
      <Card
        title="Node Trends"
        style={{ marginTop: 16 }}
        extra={
          <Select
            value={historyNode}
            onChange={setSelectedNode}
            style={{ width: 120 }}
            size="small"
          >
            {data.nodes.map(n => (
              <Select.Option key={n.name} value={n.name}>{n.name}</Select.Option>
            ))}
          </Select>
        }
      >
        {chartData.length > 0 ? (
          <ResponsiveContainer width="100%" height={250}>
            <LineChart data={chartData}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="time" fontSize={11} interval="preserveStartEnd" />
              <YAxis yAxisId="block" orientation="left" fontSize={11} />
              <YAxis yAxisId="pending" orientation="right" fontSize={11} />
              <Tooltip />
              <Legend />
              <Line yAxisId="block" type="monotone" dataKey="blockNumber" name="Block Height" stroke="#1890ff" dot={false} strokeWidth={2} />
              <Line yAxisId="pending" type="monotone" dataKey="pendingCount" name="Pending Count" stroke="#fa8c16" dot={false} strokeWidth={2} />
            </LineChart>
          </ResponsiveContainer>
        ) : (
          <Text type="secondary">Collecting data... Trend chart will appear after a few polling cycles.</Text>
        )}
      </Card>

      <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
        <Col xs={24} lg={14}>
          <Card title="Recent Batches" extra={<Link to="/batches">View All</Link>}>
            <Table
              dataSource={data.recentBatches || []}
              columns={batchColumns}
              rowKey="batchID"
              pagination={false}
              size="small"
            />
          </Card>
        </Col>
        <Col xs={24} lg={10}>
          <Card title="Node Status" extra={<Link to="/nodes">Details</Link>}>
            {data.nodes.map((n, i) => (
              <div key={n.name} style={{ display: 'flex', justifyContent: 'space-between', padding: '8px 0', borderBottom: '1px solid #f0f0f0' }}>
                <span>
                  <Badge color={NODE_COLORS[i % NODE_COLORS.length]} />
                  <Link to={`/nodes`}>{n.name}</Link>
                  {' '}
                  <Tag color={n.status === 'healthy' ? 'green' : n.status === 'degraded' ? 'orange' : 'red'} style={{ fontSize: 11 }}>{n.status}</Tag>
                </span>
                <span style={{ color: '#999', fontSize: 13 }}>
                  Block {n.blockNumber} | Pending {n.pendingCount}
                </span>
              </div>
            ))}
          </Card>
        </Col>
      </Row>
    </div>
  );
}
