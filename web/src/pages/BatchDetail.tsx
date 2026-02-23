import { useState } from 'react';
import { useParams } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { Card, Descriptions, Tag, Table, Spin, Alert, Typography, Pagination } from 'antd';
import { fetchBatch, fetchBatchRUIDs, fetchProof, type BatchDetail as BD, type RUIDsData, type OTSProofData } from '../api/client';
import dayjs from 'dayjs';

const { Text } = Typography;
const RUID_PAGE_SIZE = 20;
const statusColor: Record<string, string> = {
  pending: 'blue', submitted: 'orange', confirmed: 'green', anchored: 'gold', failed: 'red',
};

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

  if (isLoading) return <Spin size="large" style={{ display: 'block', margin: '100px auto' }} />;
  if (error) return <Alert type="error" message={String(error)} />;
  if (!batch) return null;

  return (
    <div>
      <Card title={`Batch: ${batch.batchID}`} extra={<Tag color={statusColor[batch.status]}>{batch.status}</Tag>}>
        <Descriptions column={2} bordered size="small">
          <Descriptions.Item label="Batch ID">{batch.batchID}</Descriptions.Item>
          <Descriptions.Item label="On-Chain ID">{batch.onChainID || '-'}</Descriptions.Item>
          <Descriptions.Item label="Start Block">{batch.startBlock}</Descriptions.Item>
          <Descriptions.Item label="End Block">{batch.endBlock}</Descriptions.Item>
          <Descriptions.Item label="Root Hash" span={2}>
            <Text code copyable style={{ fontSize: 12 }}>{batch.rootHash}</Text>
          </Descriptions.Item>
          {batch.otsDigest && (
            <Descriptions.Item label="OTS Digest" span={2}>
              <Text code copyable style={{ fontSize: 12 }}>{batch.otsDigest}</Text>
            </Descriptions.Item>
          )}
          <Descriptions.Item label="RUID Count">{batch.ruidCount}</Descriptions.Item>
          <Descriptions.Item label="Trigger Type">{batch.triggerType}</Descriptions.Item>
          <Descriptions.Item label="Created">{batch.createdAt ? dayjs.unix(batch.createdAt).format('YYYY-MM-DD HH:mm:ss') : '-'}</Descriptions.Item>
          <Descriptions.Item label="Status"><Tag color={statusColor[batch.status]}>{batch.status}</Tag></Descriptions.Item>
          {batch.anchoredBy && (
            <Descriptions.Item label="Anchored By" span={2}>
              <Text code copyable style={{ fontSize: 12 }}>{batch.anchoredBy}</Text>
            </Descriptions.Item>
          )}
          {batch.anchorBlock ? (
            <Descriptions.Item label="Anchor Block">{batch.anchorBlock}</Descriptions.Item>
          ) : null}
        </Descriptions>
      </Card>

      {(batch.btcTxID || batch.btcBlockHeight) && (
        <Card title="BTC Confirmation" style={{ marginTop: 16 }}>
          <Descriptions column={2} bordered size="small">
            <Descriptions.Item label="BTC TxID" span={2}>
              <Text code copyable style={{ fontSize: 12 }}>{batch.btcTxID}</Text>
            </Descriptions.Item>
            <Descriptions.Item label="BTC Block Height">{batch.btcBlockHeight}</Descriptions.Item>
            <Descriptions.Item label="BTC Timestamp">
              {batch.btcTimestamp ? dayjs.unix(batch.btcTimestamp).format('YYYY-MM-DD HH:mm:ss') : '-'}
            </Descriptions.Item>
          </Descriptions>
        </Card>
      )}

      {proof && proof.hasProof && (
        <Card title="OTS Proof" style={{ marginTop: 16 }}>
          <Descriptions column={1} bordered size="small">
            <Descriptions.Item label="Has Proof"><Tag color="green">Yes</Tag></Descriptions.Item>
            <Descriptions.Item label="BTC Confirmed"><Tag color={proof.btcConfirmed ? 'green' : 'orange'}>{proof.btcConfirmed ? 'Yes' : 'No'}</Tag></Descriptions.Item>
            {proof.otsProof && (
              <Descriptions.Item label="Proof Data">
                <Text code style={{ fontSize: 11, wordBreak: 'break-all', maxHeight: 200, overflow: 'auto', display: 'block' }}>
                  {proof.otsProof.substring(0, 200)}...
                </Text>
              </Descriptions.Item>
            )}
          </Descriptions>
        </Card>
      )}

      {ruids && ruids.ruids && ruids.ruids.length > 0 && (
        <Card title={`RUIDs (${ruids.total})`} style={{ marginTop: 16 }}>
          <Table
            dataSource={ruids.ruids.map((r, i) => ({ key: (ruidPage - 1) * RUID_PAGE_SIZE + i, ruid: r }))}
            columns={[
              { title: '#', key: 'idx', width: 60, render: (_: unknown, __: unknown, i: number) => (ruidPage - 1) * RUID_PAGE_SIZE + i + 1 },
              { title: 'RUID', dataIndex: 'ruid', key: 'ruid', render: (v: string) => <Text code copyable style={{ fontSize: 12 }}>{v}</Text> },
            ]}
            pagination={false}
            size="small"
          />
          <div style={{ display: 'flex', justifyContent: 'flex-end', marginTop: 16 }}>
            <Pagination
              current={ruidPage}
              pageSize={RUID_PAGE_SIZE}
              total={ruids.total}
              onChange={(p) => setRuidPage(p)}
              showSizeChanger={false}
              size="small"
            />
          </div>
        </Card>
      )}
    </div>
  );
}
