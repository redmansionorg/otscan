import { useState, useEffect } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Card, Input, Select, Table, Tag, Alert, Space, Empty } from 'antd';
import { SearchOutlined } from '@ant-design/icons';
import { Link, useSearchParams } from 'react-router-dom';
import { fetchClaims, type ListData, type ClaimRecord } from '../api/client';

const { Option } = Select;

const columns = [
  {
    title: 'RUID', dataIndex: 'ruid', key: 'ruid', width: 200, ellipsis: true,
    render: (v: string) => <span style={{ fontFamily: 'monospace', fontSize: 12 }}>{v}</span>,
  },
  { title: 'Claimant', dataIndex: 'claimant', key: 'claimant', width: 160, ellipsis: true },
  { title: 'Submit Block', dataIndex: 'submitBlock', key: 'submitBlock', width: 110 },
  {
    title: 'Published', dataIndex: 'published', key: 'published', width: 90,
    render: (v: boolean) => <Tag color={v ? 'green' : 'default'}>{v ? 'Yes' : 'No'}</Tag>,
  },
  {
    title: 'AUID', dataIndex: 'auid', key: 'auid', width: 200, ellipsis: true,
    render: (v: string) => v ? (
      <Link to={`/claims?auid=${v}`} style={{ fontFamily: 'monospace', fontSize: 12 }}>{v}</Link>
    ) : null,
  },
  {
    title: 'PUID', dataIndex: 'puid', key: 'puid', width: 200, ellipsis: true,
    render: (v: string) => v ? (
      <Link to={`/claims?puid=${v}`} style={{ fontFamily: 'monospace', fontSize: 12 }}>{v}</Link>
    ) : null,
  },
];

export default function Claims() {
  const [searchParams] = useSearchParams();
  const [searchType, setSearchType] = useState('claimant');
  const [searchValue, setSearchValue] = useState('');
  const [query, setQuery] = useState<Record<string, string> | null>(null);

  // Auto-fill from URL params (e.g., /claims?auid=0x...)
  useEffect(() => {
    for (const key of ['auid', 'puid', 'claimant']) {
      const val = searchParams.get(key);
      if (val) {
        setSearchType(key);
        setSearchValue(val);
        setQuery({ [key]: val });
        break;
      }
    }
  }, [searchParams]);

  const { data, error } = useQuery<ListData>({
    queryKey: ['claims', query],
    queryFn: () => fetchClaims(query!),
    enabled: !!query,
  });

  const onSearch = (value: string) => {
    if (!value.trim()) return;
    setQuery({ [searchType]: value.trim() });
  };

  return (
    <Card title="Claim Records">
      <Space.Compact style={{ width: '100%', marginBottom: 16 }}>
        <Select value={searchType} onChange={setSearchType} style={{ width: 140 }}>
          <Option value="claimant">By Claimant</Option>
          <Option value="auid">By AUID</Option>
          <Option value="puid">By PUID</Option>
        </Select>
        <Input.Search
          placeholder={`Enter ${searchType}...`}
          enterButton={<><SearchOutlined /> Search</>}
          value={searchValue}
          onChange={(e) => setSearchValue(e.target.value)}
          onSearch={onSearch}
          allowClear
          style={{ flex: 1 }}
        />
      </Space.Compact>

      {error && <Alert type="error" message={String(error)} style={{ marginBottom: 16 }} />}

      {data ? (
        <Table<ClaimRecord>
          dataSource={data.items || []}
          columns={columns}
          rowKey="ruid"
          pagination={{ total: data.totalCount, pageSize: data.limit }}
          size="middle"
          locale={{ emptyText: 'No claims found' }}
        />
      ) : !query ? (
        <Empty description="Enter a search query to find claim records" />
      ) : null}
    </Card>
  );
}
