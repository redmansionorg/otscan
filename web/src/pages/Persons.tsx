import { useState, useEffect } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Table, Spin, Alert } from 'antd';
import { UserOutlined } from '@ant-design/icons';
import { Link } from 'react-router-dom';
import { fetchPersons, type PersonListData, type PersonSummary } from '../api/client';

const PAGE_SIZE = 20;

const columns = [
  {
    title: 'PUID', dataIndex: 'puid', key: 'puid', ellipsis: true, width: 360,
    render: (v: string) => (
      <Link to={`/claims?puid=${v}`} className="explorer-link" title={v}>
        <span style={{ fontFamily: 'monospace', fontSize: 13 }}>{v}</span>
      </Link>
    ),
  },
  {
    title: 'Assets (AUIDs)', dataIndex: 'assetCount', key: 'assetCount', width: 130,
    render: (v: number) => v.toLocaleString(),
    sorter: (a: PersonSummary, b: PersonSummary) => a.assetCount - b.assetCount,
  },
  {
    title: 'Total Claims', dataIndex: 'claimCount', key: 'claimCount', width: 120,
    render: (v: number) => v.toLocaleString(),
  },
];

export default function Persons() {
  const [page, setPage] = useState(1);
  useEffect(() => { document.title = 'Published Persons | OTScan'; }, []);
  const offset = (page - 1) * PAGE_SIZE;

  const { data, isLoading, error } = useQuery<PersonListData>({
    queryKey: ['persons', page],
    queryFn: () => fetchPersons(offset, PAGE_SIZE),
  });

  if (error) return <Alert type="error" message={String(error)} style={{ margin: 24 }} />;

  return (
    <div className="page-container">
      <div className="page-card">
        <h2 style={{ margin: '0 0 16px', fontSize: 18, color: '#21325b', display: 'flex', alignItems: 'center', gap: 8 }}>
          <UserOutlined /> Published Persons (PUID) {data ? `(${data.total.toLocaleString()})` : ''}
        </h2>
        {isLoading ? (
          <Spin size="large" style={{ display: 'block', margin: '60px auto' }} />
        ) : (
          <div className="explorer-table">
            <Table
              dataSource={data?.items || []}
              columns={columns}
              rowKey="puid"
              pagination={{
                current: page,
                pageSize: PAGE_SIZE,
                total: data?.total || 0,
                onChange: (p) => setPage(p),
                showSizeChanger: false,
                showTotal: (total, range) => `${range[0]}-${range[1]} of ${total}`,
              }}
              size="middle"
              locale={{ emptyText: 'No published persons found' }}
            />
          </div>
        )}
      </div>
    </div>
  );
}
