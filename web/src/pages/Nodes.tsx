import { useCallback } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { Card, Table, Tag, Badge, Descriptions, Spin, Alert, Typography } from 'antd';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';
import { fetchNodes, fetchNodeHistory, fetchNodeCalendar, fetchCalendarURLStatus, type NodeStatus, type NodeHistoryPoint, type CalendarBatchInfo, type CalendarURLStatus } from '../api/client';
import { useWebSocket, type WSEvent } from '../api/websocket';
import dayjs from 'dayjs';

const { Text } = Typography;

const statusTag = (s: string) => {
  const color = s === 'healthy' ? 'green' : s === 'degraded' ? 'orange' : 'red';
  return <Tag color={color}>{s}</Tag>;
};

const batchStatusTag = (s: string) => {
  const colorMap: Record<string, string> = {
    pending: 'default',
    submitted: 'processing',
    confirmed: 'success',
    anchored: 'green',
    failed: 'error',
  };
  return <Tag color={colorMap[s] || 'default'}>{s}</Tag>;
};

const columns = [
  {
    title: 'Node', dataIndex: 'name', key: 'name', width: 100,
    render: (v: string, r: NodeStatus) => (
      <span><Badge status={r.status === 'healthy' ? 'success' : r.status === 'degraded' ? 'warning' : 'error'} /> {v}</span>
    ),
  },
  { title: 'Status', dataIndex: 'status', key: 'status', width: 100, render: statusTag },
  { title: 'Block', dataIndex: 'blockNumber', key: 'block', width: 100 },
  { title: 'Mode', dataIndex: 'otsMode', key: 'mode', width: 80 },
  { title: 'Pending', dataIndex: 'pendingCount', key: 'pending', width: 80 },
  { title: 'Created', dataIndex: 'totalCreated', key: 'created', width: 80 },
  { title: 'Confirmed', dataIndex: 'totalConfirmed', key: 'confirmed', width: 90 },
  { title: 'Last Processed', dataIndex: 'lastProcessedBlock', key: 'lastProc', width: 120 },
  { title: 'Last Anchor', dataIndex: 'lastAnchor', key: 'anchor', ellipsis: true },
];

const calendarColumns = [
  { title: 'Batch', dataIndex: 'batchID', key: 'batch', width: 180, ellipsis: true },
  { title: 'Status', dataIndex: 'status', key: 'status', width: 100, render: batchStatusTag },
  {
    title: 'Calendar Server', dataIndex: 'calendarServer', key: 'server', ellipsis: true,
    render: (v: string) => v ? <Text copyable={{ text: v }} style={{ fontSize: 12 }}>{v.replace(/^https?:\/\//, '')}</Text> : <Text type="secondary">-</Text>,
  },
  { title: 'Attempts', dataIndex: 'attemptCount', key: 'attempts', width: 80 },
  {
    title: 'Last Attempt', dataIndex: 'lastAttemptAt', key: 'lastAttempt', width: 160,
    render: (v: string) => v ? dayjs(v).format('MM-DD HH:mm:ss') : '-',
  },
  {
    title: 'Error', dataIndex: 'lastError', key: 'error', ellipsis: true,
    render: (v: string) => v ? <Text type="danger" style={{ fontSize: 12 }}>{v}</Text> : <Text type="secondary">-</Text>,
  },
  { title: 'Blocks', key: 'blocks', width: 140, render: (_: unknown, r: CalendarBatchInfo) => `${r.startBlock}-${r.endBlock}` },
  { title: 'RUIDs', dataIndex: 'ruidCount', key: 'ruids', width: 70 },
];

function NodeCalendarStatus({ name }: { name: string }) {
  const { data, isLoading, error } = useQuery<CalendarBatchInfo[]>({
    queryKey: ['nodeCalendar', name],
    queryFn: () => fetchNodeCalendar(name),
    refetchInterval: 30000,
  });

  if (isLoading) return <Spin size="small" />;
  if (error) return <Text type="danger">{String(error)}</Text>;
  if (!data || data.length === 0) return <Text type="secondary">No pending batches.</Text>;

  return (
    <Table
      dataSource={data}
      columns={calendarColumns}
      rowKey="batchID"
      pagination={false}
      size="small"
    />
  );
}

const calendarURLColumns = [
  {
    title: 'URL', dataIndex: 'url', key: 'url', ellipsis: true,
    render: (v: string) => <Text copyable={{ text: v }} style={{ fontSize: 12 }}>{v.replace(/^https?:\/\//, '')}</Text>,
  },
  {
    title: 'Priority', dataIndex: 'priority', key: 'priority', width: 80,
    render: (v: number) => <Tag color={v >= 80 ? 'green' : v >= 50 ? 'orange' : 'red'}>{v}</Tag>,
  },
  { title: 'Success', dataIndex: 'successCount', key: 'success', width: 70 },
  { title: 'Failure', dataIndex: 'failureCount', key: 'failure', width: 70 },
  {
    title: 'Last Status', dataIndex: 'lastStatus', key: 'lastStatus', width: 90,
    render: (v: string) => <Tag color={v === 'success' ? 'green' : v === 'failure' ? 'red' : 'default'}>{v || '-'}</Tag>,
  },
  {
    title: 'Last Attempt', dataIndex: 'lastAttemptAt', key: 'lastAttempt', width: 140,
    render: (v: string) => v ? dayjs(v).format('MM-DD HH:mm:ss') : '-',
  },
  {
    title: 'Last Error', dataIndex: 'lastError', key: 'lastError', ellipsis: true,
    render: (v: string) => v ? <Text type="danger" style={{ fontSize: 11 }}>{v}</Text> : <Text type="secondary">-</Text>,
  },
];

function NodeCalendarURLStatus({ name }: { name: string }) {
  const { data, isLoading, error } = useQuery<CalendarURLStatus[]>({
    queryKey: ['calendarURLStatus', name],
    queryFn: () => fetchCalendarURLStatus(name),
    refetchInterval: 30000,
  });

  if (isLoading) return <Spin size="small" />;
  if (error) return <Text type="danger">{String(error)}</Text>;
  if (!data || data.length === 0) return <Text type="secondary">No Calendar URLs configured.</Text>;

  return (
    <Table
      dataSource={data}
      columns={calendarURLColumns}
      rowKey="url"
      pagination={false}
      size="small"
    />
  );
}

function NodeHistoryChart({ name }: { name: string }) {
  const { data: history } = useQuery<NodeHistoryPoint[]>({
    queryKey: ['nodeHistory', name],
    queryFn: () => fetchNodeHistory(name, 180),
    refetchInterval: 30000,
  });

  const chartData = (history || []).map(h => ({
    time: dayjs(h.recordedAt).format('HH:mm'),
    blockNumber: h.blockNumber,
    pendingCount: h.pendingCount,
  }));

  if (chartData.length === 0) {
    return <Text type="secondary">No history data yet.</Text>;
  }

  return (
    <ResponsiveContainer width="100%" height={150}>
      <LineChart data={chartData}>
        <CartesianGrid strokeDasharray="3 3" />
        <XAxis dataKey="time" fontSize={10} interval="preserveStartEnd" />
        <YAxis yAxisId="block" orientation="left" fontSize={10} />
        <YAxis yAxisId="pending" orientation="right" fontSize={10} />
        <Tooltip />
        <Line yAxisId="block" type="monotone" dataKey="blockNumber" name="Block" stroke="#1890ff" dot={false} strokeWidth={1.5} />
        <Line yAxisId="pending" type="monotone" dataKey="pendingCount" name="Pending" stroke="#fa8c16" dot={false} strokeWidth={1.5} />
      </LineChart>
    </ResponsiveContainer>
  );
}

export default function Nodes() {
  const queryClient = useQueryClient();

  const handleWSEvent = useCallback((event: WSEvent) => {
    if (event.type === 'node_status') {
      queryClient.invalidateQueries({ queryKey: ['nodes'] });
    }
  }, [queryClient]);

  useWebSocket(handleWSEvent);

  const { data, isLoading, error } = useQuery<NodeStatus[]>({
    queryKey: ['nodes'],
    queryFn: fetchNodes,
  });

  if (isLoading) return <Spin size="large" style={{ display: 'block', margin: '100px auto' }} />;
  if (error) return <Alert type="error" message={String(error)} />;

  return (
    <div className="page-container">
    <Card title="Validator Nodes">
      <Table
        dataSource={data || []}
        columns={columns}
        rowKey="name"
        pagination={false}
        size="middle"
        expandable={{
          expandedRowRender: (record: NodeStatus) => (
            <div>
              <Descriptions size="small" column={2} bordered style={{ marginBottom: 16 }}>
                <Descriptions.Item label="RPC URL">{record.rpcUrl}</Descriptions.Item>
                <Descriptions.Item label="Block Number">{record.blockNumber}</Descriptions.Item>
                {record.components && Object.entries(record.components).map(([k, v]) => (
                  <Descriptions.Item key={k} label={`Component: ${k}`}>
                    <Badge status={v.healthy ? 'success' : 'error'} text={v.healthy ? 'Healthy' : v.message || 'Unhealthy'} />
                  </Descriptions.Item>
                ))}
              </Descriptions>
              <Card size="small" title={`${record.name} - Calendar Server Status`} style={{ marginBottom: 16 }}>
                <NodeCalendarStatus name={record.name} />
              </Card>
              <Card size="small" title={`${record.name} - Calendar URL Priority Status`} style={{ marginBottom: 16 }}>
                <NodeCalendarURLStatus name={record.name} />
              </Card>
              <Card size="small" title={`${record.name} - Block Height & Pending Count Trend`}>
                <NodeHistoryChart name={record.name} />
              </Card>
            </div>
          ),
        }}
      />
    </Card>
    </div>
  );
}
