import { useState, useCallback } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { Card, Table, Tag, Spin, Alert } from 'antd';
import { Link } from 'react-router-dom';
import { fetchBatches, type BatchSummary, type BatchListResponse } from '../api/client';
import { useWebSocket, type WSEvent } from '../api/websocket';

const statusColor: Record<string, string> = {
  pending: 'blue', submitted: 'orange', confirmed: 'green', anchored: 'gold', failed: 'red',
};

const PAGE_SIZE = 20;

// Truncate address to short form: 0x1234...abcd
const shortAddr = (addr?: string) => addr ? `${addr.slice(0, 6)}...${addr.slice(-4)}` : '-';

const columns = [
  {
    title: 'On-Chain ID', dataIndex: 'onChainID', key: 'id', width: 100,
    render: (v: number) => v || '-',
  },
  {
    title: 'Batch ID', dataIndex: 'batchID', key: 'batchID', width: 200,
    render: (v: string) => <Link to={`/batches/${v}`}>{v}</Link>,
  },
  {
    title: 'Block Range', key: 'range', width: 160,
    render: (_: unknown, r: BatchSummary) => `${r.startBlock} - ${r.endBlock}`,
  },
  { title: 'RUIDs', dataIndex: 'ruidCount', key: 'ruidCount', width: 80 },
  {
    title: 'Status', dataIndex: 'status', key: 'status', width: 110,
    render: (s: string) => <Tag color={statusColor[s] || 'default'}>{s}</Tag>,
  },
  {
    title: 'Anchor Node', dataIndex: 'anchoredBy', key: 'anchoredBy', width: 140,
    render: (v: string) => v ? <span title={v}>{shortAddr(v)}</span> : '-',
  },
];

export default function Batches() {
  const queryClient = useQueryClient();
  const [page, setPage] = useState(1);

  const handleWSEvent = useCallback((event: WSEvent) => {
    if (event.type === 'batch_update') {
      queryClient.invalidateQueries({ queryKey: ['batches'] });
    }
  }, [queryClient]);

  useWebSocket(handleWSEvent);

  const { data, isLoading, error } = useQuery<BatchListResponse>({
    queryKey: ['batches', page],
    queryFn: () => fetchBatches({ page: String(page), pageSize: String(PAGE_SIZE) }),
  });

  if (isLoading) return <Spin size="large" style={{ display: 'block', margin: '100px auto' }} />;
  if (error) return <Alert type="error" message={String(error)} />;

  const batches = data?.batches || [];

  return (
    <Card title={`Batches (${data?.total || 0})`}>
      <Table
        dataSource={batches}
        columns={columns}
        rowKey="batchID"
        pagination={{
          current: page,
          pageSize: PAGE_SIZE,
          total: data?.total || 0,
          onChange: (p) => setPage(p),
          showSizeChanger: false,
        }}
        size="middle"
      />
    </Card>
  );
}
