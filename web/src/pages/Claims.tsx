import { useState, useEffect, useCallback } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { Table, Tag, Alert, Segmented, Input, Select, Space, Spin } from 'antd';
import { FileSearchOutlined, SearchOutlined } from '@ant-design/icons';
import { Link, useSearchParams } from 'react-router-dom';
import { fetchClaims, type ListData, type ClaimRecord } from '../api/client';
import { useWebSocket, type WSEvent } from '../api/websocket';

const { Option } = Select;

const shortAddr = (addr?: string) => addr ? `${addr.slice(0, 6)}...${addr.slice(-4)}` : '-';

const FILTER_OPTIONS = [
  { label: 'All', value: '' },
  { label: 'Anchored', value: 'anchored' },
  { label: 'Non-Anchored', value: 'non-anchored' },
  { label: 'Published', value: 'published' },
];

const PAGE_SIZE = 20;

const columns = [
  {
    title: 'RUID', dataIndex: 'ruid', key: 'ruid', ellipsis: true, width: 220,
    render: (v: string) => (
      <Link to={`/verify?ruid=${v}`} className="explorer-link" title={v}>
        <span style={{ fontFamily: 'monospace', fontSize: 12 }}>{v}</span>
      </Link>
    ),
  },
  {
    title: 'Claimant', dataIndex: 'claimant', key: 'claimant', width: 140,
    render: (v: string) => v ? (
      <Link to={`/claims?claimant=${v}`} className="explorer-link" title={v}>{shortAddr(v)}</Link>
    ) : '-',
  },
  { title: 'Submit Block', dataIndex: 'submitBlock', key: 'submitBlock', width: 110 },
  {
    title: 'Published', dataIndex: 'published', key: 'published', width: 90,
    render: (v: boolean) => <Tag color={v ? 'green' : 'default'}>{v ? 'Yes' : 'No'}</Tag>,
  },
  {
    title: 'AUID', dataIndex: 'auid', key: 'auid', width: 180, ellipsis: true,
    render: (v: string) => v ? (
      <Link to={`/claims?auid=${v}`} className="explorer-link" title={v}>
        <span style={{ fontFamily: 'monospace', fontSize: 12 }}>{shortAddr(v)}</span>
      </Link>
    ) : '-',
  },
  {
    title: 'PUID', dataIndex: 'puid', key: 'puid', width: 180, ellipsis: true,
    render: (v: string) => v ? (
      <Link to={`/claims?puid=${v}`} className="explorer-link" title={v}>
        <span style={{ fontFamily: 'monospace', fontSize: 12 }}>{shortAddr(v)}</span>
      </Link>
    ) : '-',
  },
];

export default function Claims() {
  const queryClient = useQueryClient();
  const [searchParams, setSearchParams] = useSearchParams();
  const filter = searchParams.get('filter') || '';
  const [page, setPage] = useState(1);
  const [searchType, setSearchType] = useState('claimant');
  const [searchValue, setSearchValue] = useState('');

  useEffect(() => {
    const label = filter ? `${filter.charAt(0).toUpperCase() + filter.slice(1)} Claims` : 'All Claims';
    document.title = `${label} | OTScan`;
  }, [filter]);

  const handleWSEvent = useCallback((event: WSEvent) => {
    if (event.type === 'claim_sync') {
      queryClient.invalidateQueries({ queryKey: ['claimsList'] });
    }
  }, [queryClient]);

  useWebSocket(handleWSEvent);

  // Build query params based on filter, search, or URL params
  const [activeQuery, setActiveQuery] = useState<Record<string, string> | null>(null);

  // Auto-fill from URL params
  useEffect(() => {
    for (const key of ['auid', 'puid', 'claimant']) {
      const val = searchParams.get(key);
      if (val) {
        setSearchType(key);
        setSearchValue(val);
        setActiveQuery({ [key]: val, offset: '0', limit: String(PAGE_SIZE) });
        return;
      }
    }
    // No specific search param - use filter/list mode
    setActiveQuery(null);
  }, [searchParams]);

  // List mode query (filter or all)
  const isSearchMode = !!activeQuery;
  const offset = (page - 1) * PAGE_SIZE;
  const listParams: Record<string, string> = { offset: String(offset), limit: String(PAGE_SIZE) };
  if (filter) listParams.filter = filter;
  if (!filter && !isSearchMode) listParams.sort = 'latest';

  const { data: listData, isLoading: listLoading, error: listError } = useQuery<ListData>({
    queryKey: ['claimsList', filter, page],
    queryFn: () => fetchClaims(listParams),
    enabled: !isSearchMode,
  });

  const { data: searchData, isLoading: searchLoading, error: searchError } = useQuery<ListData>({
    queryKey: ['claimsSearch', activeQuery],
    queryFn: () => fetchClaims(activeQuery!),
    enabled: isSearchMode,
  });

  const data = isSearchMode ? searchData : listData;
  const loading = isSearchMode ? searchLoading : listLoading;
  const error = isSearchMode ? searchError : listError;

  const handleFilterChange = (value: string | number) => {
    const v = String(value);
    setPage(1);
    setActiveQuery(null);
    setSearchValue('');
    if (v) {
      setSearchParams({ filter: v });
    } else {
      setSearchParams({});
    }
  };

  const onSearch = (value: string) => {
    if (!value.trim()) return;
    setSearchParams({ [searchType]: value.trim() });
  };

  const clearSearch = () => {
    setSearchValue('');
    setActiveQuery(null);
    setSearchParams({});
  };

  return (
    <div className="page-container">
      <div className="page-card">
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16, flexWrap: 'wrap', gap: 12 }}>
          <h2 style={{ margin: 0, fontSize: 18, color: '#21325b', display: 'flex', alignItems: 'center', gap: 8 }}>
            <FileSearchOutlined /> Claims {data ? `(${(data.totalCount ?? 0).toLocaleString()})` : ''}
          </h2>
          {!isSearchMode && (
            <Segmented
              options={FILTER_OPTIONS}
              value={filter}
              onChange={handleFilterChange}
              size="middle"
            />
          )}
        </div>

        {/* Search bar */}
        <div style={{ marginBottom: 16 }}>
          <Space.Compact style={{ width: '100%' }}>
            <Select value={searchType} onChange={setSearchType} style={{ width: 140 }}>
              <Option value="claimant">By Claimant</Option>
              <Option value="auid">By AUID</Option>
              <Option value="puid">By PUID</Option>
            </Select>
            <Input.Search
              placeholder={`Search by ${searchType}...`}
              enterButton={<><SearchOutlined /> Search</>}
              value={searchValue}
              onChange={(e) => setSearchValue(e.target.value)}
              onSearch={onSearch}
              allowClear
              onClear={clearSearch}
              style={{ flex: 1 }}
            />
          </Space.Compact>
          {isSearchMode && (
            <div style={{ marginTop: 8, fontSize: 13, color: '#999' }}>
              Showing results for <strong>{searchType}</strong>: <code>{searchParams.get(searchType)}</code>
              {' '}<a onClick={clearSearch} style={{ color: '#1e88e5', cursor: 'pointer' }}>Clear search</a>
            </div>
          )}
        </div>

        {error && <Alert type="error" message={String(error)} style={{ marginBottom: 16 }} />}

        {loading ? (
          <Spin size="large" style={{ display: 'block', margin: '60px auto' }} />
        ) : (
          <div className="explorer-table">
            <Table<ClaimRecord>
              dataSource={data?.items || []}
              columns={columns}
              rowKey="ruid"
              pagination={isSearchMode ? {
                total: data?.totalCount ?? 0,
                pageSize: PAGE_SIZE,
              } : {
                current: page,
                pageSize: PAGE_SIZE,
                total: data?.totalCount ?? 0,
                onChange: (p) => setPage(p),
                showSizeChanger: false,
                showTotal: (total, range) => `${range[0]}-${range[1]} of ${total}`,
              }}
              size="middle"
              locale={{ emptyText: 'No claims found' }}
            />
          </div>
        )}
      </div>
    </div>
  );
}
