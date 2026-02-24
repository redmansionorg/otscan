import { useState, useCallback, useEffect } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { Table, Tag, Spin, Alert, Segmented } from 'antd';
import { Link, useSearchParams } from 'react-router-dom';
import { DatabaseOutlined } from '@ant-design/icons';
import { fetchBatches, type BatchSummary, type BatchListResponse } from '../api/client';
import { useWebSocket, type WSEvent } from '../api/websocket';

const statusColor: Record<string, string> = {
  pending: 'blue', submitted: 'orange', confirmed: 'green', anchored: 'gold', failed: 'red',
};

const PAGE_SIZE = 20;
const shortAddr = (addr?: string) => addr ? `${addr.slice(0, 6)}...${addr.slice(-4)}` : '-';

const FILTER_OPTIONS = [
  { label: 'All', value: '' },
  { label: 'Non-Empty', value: 'non-empty' },
  { label: 'Empty', value: 'empty' },
  { label: 'Pending', value: 'pending' },
];

const columns = [
  {
    title: 'On-Chain ID', dataIndex: 'onChainID', key: 'id', width: 100,
    render: (v: number, r: BatchSummary) => v ? <Link to={`/batches/${r.batchID}`} className="explorer-link">#{v}</Link> : '-',
  },
  {
    title: 'Batch ID', dataIndex: 'batchID', key: 'batchID', ellipsis: true, width: 220,
    render: (v: string) => <Link to={`/batches/${v}`} className="explorer-link" title={v}>{v}</Link>,
  },
  {
    title: 'Block Range', key: 'range', width: 160,
    render: (_: unknown, r: BatchSummary) => `${r.startBlock} - ${r.endBlock}`,
  },
  {
    title: 'RUIDs', dataIndex: 'ruidCount', key: 'ruidCount', width: 90,
    render: (v: number) => v.toLocaleString(),
  },
  {
    title: 'Status', dataIndex: 'status', key: 'status', width: 110,
    render: (s: string) => <Tag color={statusColor[s] || 'default'}>{s}</Tag>,
  },
  {
    title: 'Anchor Node', key: 'anchoredBy', width: 180,
    render: (_: unknown, r: BatchSummary) => {
      if (!r.anchoredBy) return '-';
      const label = r.anchoredByName
        ? `${r.anchoredByName} (${shortAddr(r.anchoredBy)})`
        : shortAddr(r.anchoredBy);
      return <span title={r.anchoredBy}>{label}</span>;
    },
  },
];

export default function Batches() {
  const queryClient = useQueryClient();
  const [searchParams, setSearchParams] = useSearchParams();
  const filter = searchParams.get('filter') || '';
  const [page, setPage] = useState(1);

  useEffect(() => {
    const label = filter ? `${filter.charAt(0).toUpperCase() + filter.slice(1)} Batches` : 'All Batches';
    document.title = `${label} | OTScan`;
  }, [filter]);

  const handleWSEvent = useCallback((event: WSEvent) => {
    if (event.type === 'batch_update') {
      queryClient.invalidateQueries({ queryKey: ['batches'] });
    }
  }, [queryClient]);

  useWebSocket(handleWSEvent);

  const { data, isLoading, error } = useQuery<BatchListResponse>({
    queryKey: ['batches', page, filter],
    queryFn: () => {
      const params: Record<string, string> = { page: String(page), pageSize: String(PAGE_SIZE) };
      if (filter) params.filter = filter;
      return fetchBatches(params);
    },
  });

  const handleFilterChange = (value: string | number) => {
    const v = String(value);
    setPage(1);
    if (v) {
      setSearchParams({ filter: v });
    } else {
      setSearchParams({});
    }
  };

  if (error) return <Alert type="error" message={String(error)} style={{ margin: 24 }} />;

  return (
    <div className="page-container">
      <div className="page-card">
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
          <h2 style={{ margin: 0, fontSize: 18, color: '#21325b', display: 'flex', alignItems: 'center', gap: 8 }}>
            <DatabaseOutlined /> Batches {data ? `(${data.total.toLocaleString()})` : ''}
          </h2>
          <Segmented
            options={FILTER_OPTIONS}
            value={filter}
            onChange={handleFilterChange}
            size="middle"
          />
        </div>
        {isLoading ? (
          <Spin size="large" style={{ display: 'block', margin: '60px auto' }} />
        ) : (
          <div className="explorer-table">
            <Table
              dataSource={data?.batches || []}
              columns={columns}
              rowKey="batchID"
              pagination={{
                current: page,
                pageSize: PAGE_SIZE,
                total: data?.total || 0,
                onChange: (p) => setPage(p),
                showSizeChanger: false,
                showTotal: (total, range) => `${range[0]}-${range[1]} of ${total}`,
              }}
              size="middle"
            />
          </div>
        )}
      </div>
    </div>
  );
}
