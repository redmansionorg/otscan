import { useState, useEffect } from 'react';
import { Card, Input, Result, Descriptions, Tag, Alert, Typography, Collapse, Steps } from 'antd';
import {
  SafetyCertificateOutlined, CheckCircleOutlined, CloseCircleOutlined,
  LinkOutlined, LockOutlined, AuditOutlined, ThunderboltOutlined,
  CloudOutlined, PlusOutlined, ArrowRightOutlined,
} from '@ant-design/icons';
import { Link, useSearchParams } from 'react-router-dom';
import { verifyRUID, type VerifyData, type ParsedOTSProof } from '../api/client';
import dayjs from 'dayjs';

const { Text, Paragraph } = Typography;

// ─── Merkle Inclusion Proof ──────────────────────────────────────

function parseMerkleProof(proofHex: string) {
  if (!proofHex?.startsWith('0x')) return null;
  const raw = proofHex.slice(2);
  if (raw.length < 136) return null;

  const ruid = '0x' + raw.slice(0, 64);
  const rootHash = '0x' + raw.slice(64, 128);
  const count = parseInt(raw.slice(128, 136), 16);

  const steps: { sibling: string; direction: string }[] = [];
  let offset = 136;
  for (let i = 0; i < count && offset + 66 <= raw.length; i++) {
    const sibling = '0x' + raw.slice(offset, offset + 64);
    const dirByte = parseInt(raw.slice(offset + 64, offset + 66), 16);
    steps.push({ sibling, direction: dirByte === 0 ? 'Left' : 'Right' });
    offset += 66;
  }
  return { ruid, rootHash, steps };
}

function MerkleProofView({ proofHex }: { proofHex: string }) {
  const parsed = parseMerkleProof(proofHex);
  if (!parsed || parsed.steps.length === 0) {
    return <Text code copyable style={{ fontSize: 11, wordBreak: 'break-all' }}>{proofHex}</Text>;
  }

  return (
    <Steps
      direction="vertical"
      size="small"
      current={parsed.steps.length + 1}
      items={[
        {
          title: 'Leaf (RUID)',
          description: <Text code copyable style={{ fontSize: 11 }}>{parsed.ruid}</Text>,
          icon: <AuditOutlined style={{ color: '#1890ff' }} />,
        },
        ...parsed.steps.map((step, i) => ({
          title: `Step ${i + 1} — Sibling ${step.direction}`,
          description: <Text code copyable style={{ fontSize: 11 }}>{step.sibling}</Text>,
          icon: <LinkOutlined style={{ color: '#722ed1' }} />,
        })),
        {
          title: 'Merkle Root (Keccak256)',
          description: <Text code copyable style={{ fontSize: 11 }}>{parsed.rootHash}</Text>,
          icon: <LockOutlined style={{ color: '#52c41a' }} />,
        },
      ]}
    />
  );
}

// ─── OTS Proof Path ──────────────────────────────────────────────

const opIcon = (op: string) => {
  switch (op) {
    case 'sha256': case 'ripemd160': case 'keccak256':
      return <ThunderboltOutlined style={{ color: '#fa8c16' }} />;
    case 'append': case 'prepend':
      return <PlusOutlined style={{ color: '#1890ff' }} />;
    default:
      return <ArrowRightOutlined style={{ color: '#999' }} />;
  }
};

const opLabel = (op: string) => {
  switch (op) {
    case 'sha256': return 'SHA-256';
    case 'ripemd160': return 'RIPEMD-160';
    case 'keccak256': return 'KECCAK-256';
    case 'append': return 'Append';
    case 'prepend': return 'Prepend';
    case 'reverse': return 'Reverse';
    default: return op;
  }
};

function OTSProofPathView({ parsed }: { parsed: ParsedOTSProof }) {
  const items = [
    {
      title: `Input Digest (${parsed.hashType.toUpperCase()})`,
      description: <Text code copyable style={{ fontSize: 11 }}>{parsed.digest}</Text>,
      icon: <AuditOutlined style={{ color: '#1890ff' }} />,
    },
    ...parsed.operations.map((op) => ({
      title: <span><Tag color={op.op === 'sha256' || op.op === 'ripemd160' ? 'orange' : 'blue'} style={{ fontSize: 11 }}>{opLabel(op.op)}</Tag></span>,
      description: op.argument ? (
        <Text code copyable style={{ fontSize: 11 }}>{op.argument}</Text>
      ) : (
        <Text type="secondary" style={{ fontSize: 11 }}>hash current value</Text>
      ),
      icon: opIcon(op.op),
    })),
    ...parsed.attestations.map((att) => ({
      title: att.type === 'bitcoin'
        ? <span><Tag color="gold">Bitcoin Attestation</Tag> Block #{att.btcBlockHeight?.toLocaleString()}</span>
        : <span><Tag color="blue">Pending</Tag> {att.calendarUrl}</span>,
      description: att.type === 'bitcoin'
        ? <Text type="success" style={{ fontSize: 12 }}>Anchored to Bitcoin blockchain</Text>
        : <Text type="secondary" style={{ fontSize: 12 }}>Awaiting confirmation from calendar server</Text>,
      icon: att.type === 'bitcoin'
        ? <LockOutlined style={{ color: '#faad14' }} />
        : <CloudOutlined style={{ color: '#1890ff' }} />,
    })),
  ];

  return (
    <Steps
      direction="vertical"
      size="small"
      current={items.length - 1}
      items={items}
    />
  );
}

// ─── Main Component ──────────────────────────────────────────────

export default function Verify() {
  const [searchParams] = useSearchParams();
  const [ruid, setRuid] = useState('');
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<VerifyData | null>(null);
  const [error, setError] = useState<string | null>(null);

  const handleVerify = async (value?: string) => {
    const v = (value || ruid).trim();
    if (!v) return;
    setRuid(v);
    setLoading(true);
    setError(null);
    setResult(null);
    try {
      const data = await verifyRUID(v);
      setResult(data);
    } catch (err) {
      setError(String(err));
    } finally {
      setLoading(false);
    }
  };

  // Auto-verify from URL query param
  useEffect(() => {
    const ruidParam = searchParams.get('ruid');
    if (ruidParam && ruidParam.startsWith('0x')) {
      setRuid(ruidParam);
      handleVerify(ruidParam);
    }
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  return (
    <div className="page-container" style={{ maxWidth: 850 }}>
      <Card title="RUID Verification" style={{ marginBottom: 16 }}>
        <p style={{ color: '#666', marginBottom: 16 }}>
          Verify that a RUID is included in an anchored batch with Bitcoin confirmation.
        </p>
        <Input.Search
          size="large"
          placeholder="Enter RUID (0x...)"
          enterButton={<><SafetyCertificateOutlined /> Verify</>}
          value={ruid}
          onChange={(e) => setRuid(e.target.value)}
          onSearch={() => handleVerify()}
          loading={loading}
        />
      </Card>

      {error && <Alert type="error" message="Verification Failed" description={error} style={{ marginBottom: 16 }} />}

      {result && (
        <>
          {/* Verification Result */}
          <Card style={{ marginBottom: 16 }}>
            <Result
              icon={result.verified ? <CheckCircleOutlined style={{ color: '#52c41a' }} /> : <CloseCircleOutlined style={{ color: '#ff4d4f' }} />}
              title={result.verified ? 'RUID Verified' : 'RUID Not Verified'}
              subTitle={result.message}
              style={{ padding: '16px 0' }}
            />
            <Descriptions bordered size="small" column={1}>
              <Descriptions.Item label="RUID">
                <Text code copyable>{result.ruid}</Text>
              </Descriptions.Item>
              <Descriptions.Item label="Verified">
                <Tag color={result.verified ? 'green' : 'red'}>{result.verified ? 'Yes' : 'No'}</Tag>
              </Descriptions.Item>
              {result.published !== undefined && (
                <Descriptions.Item label="Published">
                  <Tag color={result.published ? 'green' : 'default'}>{result.published ? 'Yes' : 'No'}</Tag>
                </Descriptions.Item>
              )}
              {result.claimant && (
                <Descriptions.Item label="Claimant">{result.claimant}</Descriptions.Item>
              )}
              {result.submitBlock && (
                <Descriptions.Item label="Submit Block">{result.submitBlock.toLocaleString()}</Descriptions.Item>
              )}
              {result.auid && (
                <Descriptions.Item label="AUID">
                  <Link to={`/claims?auid=${result.auid}`}>
                    <Text code copyable style={{ fontSize: 12 }}>{result.auid}</Text>
                  </Link>
                </Descriptions.Item>
              )}
              {result.puid && (
                <Descriptions.Item label="PUID">
                  <Link to={`/claims?puid=${result.puid}`}>
                    <Text code copyable style={{ fontSize: 12 }}>{result.puid}</Text>
                  </Link>
                </Descriptions.Item>
              )}
              {result.batchID && (
                <Descriptions.Item label="Batch ID">
                  <Link to={`/batches/${result.batchID}`}>{result.batchID}</Link>
                </Descriptions.Item>
              )}
              {result.rootHash && (
                <Descriptions.Item label="Merkle Root">
                  <Text code copyable style={{ fontSize: 12 }}>{result.rootHash}</Text>
                </Descriptions.Item>
              )}
              {result.otsDigest && (
                <Descriptions.Item label="OTS Digest">
                  <Text code copyable style={{ fontSize: 12 }}>{result.otsDigest}</Text>
                  <div style={{ marginTop: 4 }}>
                    <Text type="secondary" style={{ fontSize: 11 }}>= SHA256(Merkle Root)</Text>
                  </div>
                </Descriptions.Item>
              )}
              {result.btcBlockHeight !== undefined && result.btcBlockHeight > 0 && (
                <Descriptions.Item label="BTC Block Height">{result.btcBlockHeight.toLocaleString()}</Descriptions.Item>
              )}
              {result.btcTimestamp !== undefined && result.btcTimestamp > 0 && (
                <Descriptions.Item label="BTC Timestamp">
                  {dayjs.unix(result.btcTimestamp).format('YYYY-MM-DD HH:mm:ss UTC')}
                </Descriptions.Item>
              )}
            </Descriptions>
          </Card>

          {/* Cryptographic Proofs */}
          {result.verified && (result.merkleProof || result.parsedOTSProof) && (
            <Card title="Cryptographic Proof Chain">
              <Paragraph type="secondary" style={{ marginBottom: 16, fontSize: 13 }}>
                Complete verification path: RUID → Merkle Root → OTS Digest → Bitcoin Blockchain
              </Paragraph>
              <Collapse
                defaultActiveKey={['merkle', 'ots']}
                items={[
                  ...(result.merkleProof ? [{
                    key: 'merkle',
                    label: (
                      <span>
                        <Tag color="purple">Layer 1</Tag>
                        Merkle Inclusion Proof
                        <Text type="secondary" style={{ fontSize: 11, marginLeft: 8 }}>RUID → Batch Merkle Root</Text>
                      </span>
                    ),
                    children: <MerkleProofView proofHex={result.merkleProof} />,
                  }] : []),
                  ...(result.parsedOTSProof ? [{
                    key: 'ots',
                    label: (
                      <span>
                        <Tag color="gold">Layer 2</Tag>
                        OpenTimestamps Proof Path
                        <Text type="secondary" style={{ fontSize: 11, marginLeft: 8 }}>OTS Digest → Bitcoin Block</Text>
                      </span>
                    ),
                    children: <OTSProofPathView parsed={result.parsedOTSProof} />,
                  }] : []),
                  ...(result.otsProof ? [{
                    key: 'raw',
                    label: (
                      <span>
                        <Tag>Raw</Tag>
                        OTS Proof Binary
                      </span>
                    ),
                    children: (
                      <div style={{ background: '#f5f5f5', padding: 12, borderRadius: 6, maxHeight: 200, overflow: 'auto' }}>
                        <Text code copyable style={{ fontSize: 11, wordBreak: 'break-all' }}>{result.otsProof}</Text>
                      </div>
                    ),
                  }] : []),
                ]}
              />
            </Card>
          )}
        </>
      )}
    </div>
  );
}
