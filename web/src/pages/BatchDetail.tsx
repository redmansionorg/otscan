import { useState, useEffect } from 'react';
import { useParams, Link } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { Tag, Table, Spin, Alert, Typography, Pagination } from 'antd';
import { DatabaseOutlined, CheckCircleOutlined, LinkOutlined, SafetyCertificateOutlined } from '@ant-design/icons';
import { fetchBatch, fetchBatchRUIDs, fetchProof, type BatchDetail as BD, type RUIDsData, type OTSProofData } from '../api/client';
import dayjs from 'dayjs';

const { Text } = Typography;
const RUID_PAGE_SIZE = 20;
const statusColor: Record<string, string> = {
  pending: 'blue', submitted: 'orange', confirmed: 'green', anchored: 'gold', failed: 'red',
};

function DetailRow({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div style={{ display: 'flex', borderBottom: '1px solid #f0f0f0', padding: '10px 0' }}>
      <div style={{ width: 160, flexShrink: 0, color: '#666', fontSize: 13 }}>{label}:</div>
      <div style={{ flex: 1, fontSize: 13 }}>{children}</div>
    </div>
  );
}

export default function BatchDetail() {
  const { id } = useParams<{ id: string }>();
  const [ruidPage, setRuidPage] = useState(1);

  const { data: batch, isLoading, error } = useQuery<BD>({
    queryKey: ['batch', id],
    queryFn: () => fetchBatch(id!),
    enabled: !!id,
  });

  const { data: ruids } = useQuery<RUIDsData>({
    queryKey: ['batchRuids', id, ruidPage],
    queryFn: () => fetchBatchRUIDs(id!, (ruidPage - 1) * RUID_PAGE_SIZE, RUID_PAGE_SIZE),
    enabled: !!id,
  });

  const { data: proof } = useQuery<OTSProofData>({
    queryKey: ['proof', id],
    queryFn: () => fetchProof(id!),
    enabled: !!id,
  });

  useEffect(() => {
    if (batch) {
      document.title = `Batch ${batch.onChainID ? `#${batch.onChainID}` : batch.batchID} | OTScan`;
    }
  }, [batch]);

  if (isLoading) return <Spin size="large" style={{ display: 'block', margin: '100px auto' }} />;
  if (error) return <Alert type="error" message={String(error)} style={{ margin: 24 }} />;
  if (!batch) return null;

  return (
    <div className="page-container">
      {/* Batch Overview */}
      <div className="page-card" style={{ marginBottom: 16 }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 12, marginBottom: 16 }}>
          <DatabaseOutlined style={{ fontSize: 20, color: '#3498db' }} />
          <h2 style={{ margin: 0, fontSize: 18, color: '#21325b' }}>
            Batch {batch.onChainID ? `#${batch.onChainID}` : ''}
          </h2>
          <Tag color={statusColor[batch.status]} style={{ fontSize: 13 }}>{batch.status}</Tag>
        </div>

        <DetailRow label="Batch ID">
          <Text code copyable style={{ fontSize: 12 }}>{batch.batchID}</Text>
        </DetailRow>
        <DetailRow label="On-Chain ID">{batch.onChainID || '-'}</DetailRow>
        <DetailRow label="Block Range">
          {batch.startBlock} &rarr; {batch.endBlock}
          <span style={{ color: '#999', marginLeft: 8 }}>({batch.endBlock - batch.startBlock} blocks)</span>
        </DetailRow>
        <DetailRow label="Root Hash">
          <Text code copyable style={{ fontSize: 12 }}>{batch.rootHash}</Text>
        </DetailRow>
        {batch.otsDigest && (
          <DetailRow label="OTS Digest">
            <Text code copyable style={{ fontSize: 12 }}>{batch.otsDigest}</Text>
          </DetailRow>
        )}
        <DetailRow label="RUID Count">
          <strong>{batch.ruidCount.toLocaleString()}</strong>
        </DetailRow>
        <DetailRow label="Trigger Type">{batch.triggerType || '-'}</DetailRow>
        <DetailRow label="Created">
          {batch.createdAt ? dayjs.unix(batch.createdAt).format('YYYY-MM-DD HH:mm:ss') : '-'}
        </DetailRow>
        {batch.anchoredBy && (
          <DetailRow label="Anchored By">
            <Text code copyable style={{ fontSize: 12 }}>{batch.anchoredBy}</Text>
          </DetailRow>
        )}
        {batch.anchorBlock ? (
          <DetailRow label="Anchor Block">{batch.anchorBlock}</DetailRow>
        ) : null}
      </div>

      {/* BTC Confirmation */}
      {(batch.btcTxID || batch.btcBlockHeight) && (
        <div className="page-card" style={{ marginBottom: 16 }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 12 }}>
            <LinkOutlined style={{ color: '#d48806' }} />
            <h3 style={{ margin: 0, fontSize: 16, color: '#21325b' }}>BTC Confirmation</h3>
            <Tag color="gold">Bitcoin</Tag>
          </div>
          <DetailRow label="BTC TxID">
            <Text code copyable style={{ fontSize: 12 }}>{batch.btcTxID}</Text>
          </DetailRow>
          <DetailRow label="BTC Block Height">{batch.btcBlockHeight}</DetailRow>
          <DetailRow label="BTC Timestamp">
            {batch.btcTimestamp ? dayjs.unix(batch.btcTimestamp).format('YYYY-MM-DD HH:mm:ss') : '-'}
          </DetailRow>
        </div>
      )}

      {/* OTS Proof */}
      {proof && proof.hasProof && (
        <div className="page-card" style={{ marginBottom: 16 }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 12 }}>
            <SafetyCertificateOutlined style={{ color: '#52c41a' }} />
            <h3 style={{ margin: 0, fontSize: 16, color: '#21325b' }}>OTS Proof</h3>
          </div>
          <DetailRow label="Has Proof"><Tag color="green"><CheckCircleOutlined /> Yes</Tag></DetailRow>
          <DetailRow label="BTC Confirmed">
            <Tag color={proof.btcConfirmed ? 'green' : 'orange'}>
              {proof.btcConfirmed ? 'Yes' : 'Pending'}
            </Tag>
          </DetailRow>
          {proof.otsProof && (
            <DetailRow label="Proof Data">
              <Text code style={{ fontSize: 11, wordBreak: 'break-all', maxHeight: 150, overflow: 'auto', display: 'block' }}>
                {proof.otsProof.substring(0, 200)}...
              </Text>
            </DetailRow>
          )}
        </div>
      )}

      {/* RUIDs */}
      {ruids && ruids.ruids && ruids.ruids.length > 0 && (
        <div className="page-card">
          <h3 style={{ margin: '0 0 12px', fontSize: 16, color: '#21325b' }}>
            RUIDs ({ruids.total.toLocaleString()})
          </h3>
          <div className="explorer-table">
            <Table
              dataSource={ruids.ruids.map((r, i) => ({ key: (ruidPage - 1) * RUID_PAGE_SIZE + i, ruid: r }))}
              columns={[
                { title: '#', key: 'idx', width: 60, render: (_: unknown, __: unknown, i: number) => (ruidPage - 1) * RUID_PAGE_SIZE + i + 1 },
                {
                  title: 'RUID', dataIndex: 'ruid', key: 'ruid',
                  render: (v: string) => (
                    <Link to={`/verify?ruid=${v}`} className="explorer-link">
                      <Text code style={{ fontSize: 12 }}>{v}</Text>
                    </Link>
                  ),
                },
              ]}
              pagination={false}
              size="small"
            />
          </div>
          <div style={{ display: 'flex', justifyContent: 'flex-end', marginTop: 16 }}>
            <Pagination
              current={ruidPage}
              pageSize={RUID_PAGE_SIZE}
              total={ruids.total}
              onChange={(p) => setRuidPage(p)}
              showSizeChanger={false}
              size="small"
              showTotal={(total, range) => `${range[0]}-${range[1]} of ${total}`}
            />
          </div>
        </div>
      )}
    </div>
  );
}
