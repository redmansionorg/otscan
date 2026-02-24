import { useState, useEffect } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Table, Tag, Spin, Alert } from 'antd';
import { FileProtectOutlined } from '@ant-design/icons';
import { Link } from 'react-router-dom';
import { fetchAssets, type AssetListData, type AssetSummary } from '../api/client';

const PAGE_SIZE = 20;

const columns = [
  {
    title: 'AUID', dataIndex: 'auid', key: 'auid', ellipsis: true, width: 360,
    render: (v: string) => (
      <Link to={`/claims?auid=${v}`} className="explorer-link" title={v}>
        <span style={{ fontFamily: 'monospace', fontSize: 13 }}>{v}</span>
      </Link>
    ),
  },
  {
    title: 'Claims', dataIndex: 'claimCount', key: 'claimCount', width: 100,
    render: (v: number) => v.toLocaleString(),
    sorter: (a: AssetSummary, b: AssetSummary) => a.claimCount - b.claimCount,
  },
  {
    title: 'PUIDs', dataIndex: 'puidCount', key: 'puidCount', width: 100,
    render: (v: number) => (
      <>
        {v}
        {v > 1 && <Tag color="red" style={{ marginLeft: 6, fontSize: 11 }}>Conflict</Tag>}
      </>
    ),
  },
];

export default function Assets() {
  const [page, setPage] = useState(1);
  useEffect(() => { document.title = 'Published Assets | OTScan'; }, []);
  const offset = (page - 1) * PAGE_SIZE;

  const { data, isLoading, error } = useQuery<AssetListData>({
    queryKey: ['assets', page],
    queryFn: () => fetchAssets(offset, PAGE_SIZE),
  });

  if (error) return <Alert type="error" message={String(error)} style={{ margin: 24 }} />;

  return (
    <div className="page-container">
      <div className="page-card">
        <h2 style={{ margin: '0 0 16px', fontSize: 18, color: '#21325b', display: 'flex', alignItems: 'center', gap: 8 }}>
          <FileProtectOutlined /> Published Assets (AUID) {data ? `(${data.total.toLocaleString()})` : ''}
        </h2>
        {isLoading ? (
          <Spin size="large" style={{ display: 'block', margin: '60px auto' }} />
        ) : (
          <div className="explorer-table">
            <Table
              dataSource={data?.items || []}
              columns={columns}
              rowKey="auid"
              pagination={{
                current: page,
                pageSize: PAGE_SIZE,
                total: data?.total || 0,
                onChange: (p) => setPage(p),
                showSizeChanger: false,
                showTotal: (total, range) => `${range[0]}-${range[1]} of ${total}`,
              }}
              size="middle"
              locale={{ emptyText: 'No published assets found' }}
            />
          </div>
        )}
      </div>
    </div>
  );
}
