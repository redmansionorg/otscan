import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Card, Table, Alert, Empty } from 'antd';
import { Link } from 'react-router-dom';
import { fetchConflicts, type ConflictListData, type ConflictSummary } from '../api/client';

const columns = [
  {
    title: 'AUID', dataIndex: 'auid', key: 'auid', ellipsis: true,
    render: (v: string) => (
      <Link to={`/claims?auid=${v}`} style={{ fontFamily: 'monospace', fontSize: 12 }}>{v}</Link>
    ),
  },
  { title: 'PUID Count', dataIndex: 'puidCount', key: 'puidCount', width: 110 },
  { title: 'Claim Count', dataIndex: 'claimCount', key: 'claimCount', width: 110 },
  { title: 'Earliest Block', dataIndex: 'earliestBlock', key: 'earliestBlock', width: 130 },
  { title: 'Latest Block', dataIndex: 'latestBlock', key: 'latestBlock', width: 130 },
];

export default function Conflicts() {
  const [page, setPage] = useState(1);
  const pageSize = 20;

  const { data, isLoading, error } = useQuery<ConflictListData>({
    queryKey: ['conflicts', page],
    queryFn: () => fetchConflicts((page - 1) * pageSize, pageSize),
  });

  return (
    <Card title="Copyright Conflicts">
      {error && <Alert type="error" message={String(error)} style={{ marginBottom: 16 }} />}
      {data && (data.items?.length ?? 0) === 0 && !isLoading ? (
        <Empty description="No conflicts detected. All AUIDs have unique PUID claimants." />
      ) : (
        <Table<ConflictSummary>
          dataSource={data?.items || []}
          columns={columns}
          rowKey="auid"
          loading={isLoading}
          pagination={{
            current: page,
            total: data?.total || 0,
            pageSize,
            onChange: setPage,
            showTotal: (t) => `${t} conflicts`,
          }}
          size="middle"
        />
      )}
    </Card>
  );
}
