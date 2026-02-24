import { useState, useEffect } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Table, Spin, Alert } from 'antd';
import { TeamOutlined } from '@ant-design/icons';
import { Link } from 'react-router-dom';
import { fetchClaimants, type ClaimantListData, type ClaimantSummary } from '../api/client';

const PAGE_SIZE = 20;
const columns = [
  {
    title: 'Claimant Address', dataIndex: 'claimant', key: 'claimant', width: 300,
    render: (v: string) => (
      <Link to={`/claims?claimant=${v}`} className="explorer-link" title={v}>
        <span style={{ fontFamily: 'monospace', fontSize: 13 }}>{v}</span>
      </Link>
    ),
  },
  {
    title: 'Total Claims', dataIndex: 'claimCount', key: 'claimCount', width: 130,
    render: (v: number) => v.toLocaleString(),
    sorter: (a: ClaimantSummary, b: ClaimantSummary) => a.claimCount - b.claimCount,
  },
  {
    title: 'Published', dataIndex: 'publishedCount', key: 'publishedCount', width: 120,
    render: (v: number) => v.toLocaleString(),
  },
  {
    title: 'Latest Block', dataIndex: 'latestBlock', key: 'latestBlock', width: 120,
  },
];

export default function Claimants() {
  const [page, setPage] = useState(1);
  useEffect(() => { document.title = 'Claimants | OTScan'; }, []);
  const offset = (page - 1) * PAGE_SIZE;

  const { data, isLoading, error } = useQuery<ClaimantListData>({
    queryKey: ['claimants', page],
    queryFn: () => fetchClaimants(offset, PAGE_SIZE),
  });

  if (error) return <Alert type="error" message={String(error)} style={{ margin: 24 }} />;

  return (
    <div className="page-container">
      <div className="page-card">
        <h2 style={{ margin: '0 0 16px', fontSize: 18, color: '#21325b', display: 'flex', alignItems: 'center', gap: 8 }}>
          <TeamOutlined /> Claimants {data ? `(${data.total.toLocaleString()})` : ''}
        </h2>
        {isLoading ? (
          <Spin size="large" style={{ display: 'block', margin: '60px auto' }} />
        ) : (
          <div className="explorer-table">
            <Table
              dataSource={data?.items || []}
              columns={columns}
              rowKey="claimant"
              pagination={{
                current: page,
                pageSize: PAGE_SIZE,
                total: data?.total || 0,
                onChange: (p) => setPage(p),
                showSizeChanger: false,
                showTotal: (total, range) => `${range[0]}-${range[1]} of ${total}`,
              }}
              size="middle"
              locale={{ emptyText: 'No claimants found' }}
            />
          </div>
        )}
      </div>
    </div>
  );
}
