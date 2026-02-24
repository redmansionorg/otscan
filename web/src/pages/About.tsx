import { useEffect } from 'react';
import { Typography, Card, Steps, Tag, Divider } from 'antd';
import {
  SafetyCertificateOutlined, DatabaseOutlined, LinkOutlined,
  CloudServerOutlined, CheckCircleOutlined, FileProtectOutlined,
  BlockOutlined, TeamOutlined, AuditOutlined, ThunderboltOutlined,
} from '@ant-design/icons';
import { Link } from 'react-router-dom';

const { Title, Paragraph, Text } = Typography;

export default function About() {
  useEffect(() => { document.title = 'About | Redmansion \u00B7 OTScan'; }, []);

  return (
    <div>
      {/* Hero */}
      <div className="otscan-hero">
        <h1>Redmansion &middot; OTScan</h1>
        <div className="subtitle">Blockchain Copyright Timestamping &amp; Verification Explorer</div>
      </div>

      <div className="page-container" style={{ maxWidth: 960 }}>

        {/* What is RMC */}
        <Card style={{ marginBottom: 20 }}>
          <Title level={4} style={{ color: '#21325b', marginBottom: 16 }}>
            <BlockOutlined style={{ marginRight: 8, color: '#3498db' }} />
            What is RMC?
          </Title>
          <Paragraph style={{ fontSize: 14, lineHeight: 1.8 }}>
            <strong>RMC (Redmansion Chain)</strong> is a purpose-built blockchain for digital copyright timestamping,
            forked from <Text code>go-ethereum / BSC</Text> with the <strong>Parlia PoSA</strong> consensus engine.
            It provides fast, deterministic finality and a native <strong>OpenTimestamps (OTS) module</strong> that
            anchors copyright claims to the <strong>Bitcoin blockchain</strong>, creating immutable proof of existence.
          </Paragraph>
          <div style={{ display: 'flex', flexWrap: 'wrap', gap: 12, marginTop: 8 }}>
            <Tag color="blue" style={{ fontSize: 13, padding: '4px 12px' }}>5 Validators</Tag>
            <Tag color="blue" style={{ fontSize: 13, padding: '4px 12px' }}>3-Second Blocks</Tag>
            <Tag color="blue" style={{ fontSize: 13, padding: '4px 12px' }}>PoSA Consensus</Tag>
            <Tag color="gold" style={{ fontSize: 13, padding: '4px 12px' }}>Bitcoin Anchoring</Tag>
            <Tag color="green" style={{ fontSize: 13, padding: '4px 12px' }}>OpenTimestamps Protocol</Tag>
          </div>
        </Card>

        {/* How It Works */}
        <Card style={{ marginBottom: 20 }}>
          <Title level={4} style={{ color: '#21325b', marginBottom: 20 }}>
            <ThunderboltOutlined style={{ marginRight: 8, color: '#fa8c16' }} />
            How It Works
          </Title>
          <Steps
            direction="vertical"
            current={5}
            items={[
              {
                title: <Text strong>Claim Submission</Text>,
                description: (
                  <Paragraph style={{ margin: 0, color: '#666' }}>
                    A copyright holder submits a digital claim to the RMC chain. Each claim is assigned a unique
                    <Text code>RUID</Text> (Registration UID), optionally linked to an <Text code>AUID</Text> (Asset)
                    and <Text code>PUID</Text> (Proprietor).
                  </Paragraph>
                ),
                icon: <FileProtectOutlined style={{ color: '#1890ff' }} />,
              },
              {
                title: <Text strong>Batch Aggregation</Text>,
                description: (
                  <Paragraph style={{ margin: 0, color: '#666' }}>
                    Multiple RUIDs are collected into a <strong>batch</strong> and aggregated into a <strong>Merkle Tree</strong>.
                    The Merkle Root serves as a compact cryptographic fingerprint of all included claims.
                  </Paragraph>
                ),
                icon: <DatabaseOutlined style={{ color: '#722ed1' }} />,
              },
              {
                title: <Text strong>On-Chain Anchoring</Text>,
                description: (
                  <Paragraph style={{ margin: 0, color: '#666' }}>
                    The batch is anchored to the RMC chain via the <strong>OTS system contract</strong>{' '}
                    (<Text code>0x...9000</Text>), recording the Merkle Root and batch metadata on-chain
                    through a validator's system transaction.
                  </Paragraph>
                ),
                icon: <LinkOutlined style={{ color: '#52c41a' }} />,
              },
              {
                title: <Text strong>Bitcoin Timestamping</Text>,
                description: (
                  <Paragraph style={{ margin: 0, color: '#666' }}>
                    The OTS module computes <Text code>SHA256(Merkle Root)</Text> as the <strong>OTS Digest</strong>,
                    then submits it to <strong>OpenTimestamps calendar servers</strong>. The calendar servers aggregate
                    multiple digests and ultimately anchor them into a <strong>Bitcoin transaction</strong>.
                  </Paragraph>
                ),
                icon: <CloudServerOutlined style={{ color: '#fa8c16' }} />,
              },
              {
                title: <Text strong>Verification</Text>,
                description: (
                  <Paragraph style={{ margin: 0, color: '#666' }}>
                    Anyone can <Link to="/verify">verify a RUID</Link> by reconstructing the full proof chain:{' '}
                    <strong>RUID &rarr; Merkle Proof &rarr; Merkle Root &rarr; OTS Digest &rarr; Bitcoin Block</strong>.
                    A <strong>PDF certificate</strong> can be exported as a formal proof of existence.
                  </Paragraph>
                ),
                icon: <CheckCircleOutlined style={{ color: '#52c41a' }} />,
              },
            ]}
          />
        </Card>

        {/* What is OTScan */}
        <Card style={{ marginBottom: 20 }}>
          <Title level={4} style={{ color: '#21325b', marginBottom: 16 }}>
            <SafetyCertificateOutlined style={{ marginRight: 8, color: '#52c41a' }} />
            What is OTScan?
          </Title>
          <Paragraph style={{ fontSize: 14, lineHeight: 1.8 }}>
            <strong>OTScan</strong> is the dedicated blockchain explorer and verification tool for the RMC chain.
            It aggregates data from all validator nodes, indexes on-chain events, and provides a user-friendly
            interface for browsing, searching, and verifying copyright claims.
          </Paragraph>
          <Divider style={{ margin: '16px 0' }} />
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(280px, 1fr))', gap: 12 }}>
            {[
              { label: 'Dashboard', desc: 'Chain overview, live stats, and recent activity', link: '/' },
              { label: 'Batches', desc: 'Browse all Merkle batches with block ranges and status', link: '/batches' },
              { label: 'Claims', desc: 'Search copyright claim records by RUID, AUID, PUID, or claimant', link: '/claims' },
              { label: 'Assets', desc: 'Published digital assets (AUID) registry', link: '/publish/assets' },
              { label: 'Proprietor', desc: 'Copyright proprietors (PUID) directory', link: '/publish/persons' },
              { label: 'Verify', desc: 'RUID verification with full proof chain and PDF certificate export', link: '/verify' },
              { label: 'Nodes', desc: 'Validator node monitoring, calendar status, and health', link: '/nodes' },
            ].map(item => (
              <Link key={item.link} to={item.link} style={{ textDecoration: 'none' }}>
                <div style={{
                  padding: '12px 16px', borderRadius: 8, border: '1px solid #f0f0f0',
                  transition: 'all 0.2s', cursor: 'pointer',
                }}>
                  <Text strong style={{ color: '#1e88e5', fontSize: 14 }}>{item.label}</Text>
                  <div style={{ color: '#666', fontSize: 12, marginTop: 4 }}>{item.desc}</div>
                </div>
              </Link>
            ))}
          </div>
        </Card>

        {/* Technical Architecture */}
        <Card style={{ marginBottom: 20 }}>
          <Title level={4} style={{ color: '#21325b', marginBottom: 16 }}>
            <CloudServerOutlined style={{ marginRight: 8, color: '#722ed1' }} />
            Technical Architecture
          </Title>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(300px, 1fr))', gap: 20 }}>
            {/* RMC Chain */}
            <div>
              <Title level={5} style={{ color: '#21325b' }}>RMC Chain</Title>
              <div style={{ fontFamily: 'monospace', fontSize: 13, lineHeight: 2, color: '#555', paddingLeft: 8, borderLeft: '3px solid #3498db' }}>
                <div><strong>go-ethereum / BSC fork</strong></div>
                <div>&nbsp;&nbsp;Parlia PoSA Consensus</div>
                <div>&nbsp;&nbsp;OTS System Contract <Text code>0x...9000</Text></div>
                <div>&nbsp;&nbsp;OTS Module (Merkle + OpenTimestamps)</div>
                <div>&nbsp;&nbsp;LevelDB (batch metadata, proofs)</div>
              </div>
            </div>
            {/* OTScan */}
            <div>
              <Title level={5} style={{ color: '#21325b' }}>OTScan</Title>
              <div style={{ fontFamily: 'monospace', fontSize: 13, lineHeight: 2, color: '#555', paddingLeft: 8, borderLeft: '3px solid #52c41a' }}>
                <div><strong>Go Backend</strong> (Gin HTTP + WebSocket)</div>
                <div>&nbsp;&nbsp;Multi-node RPC aggregation</div>
                <div>&nbsp;&nbsp;Batch &amp; claim indexer</div>
                <div><strong>PostgreSQL</strong> (claims, batches, node status)</div>
                <div><strong>Redis</strong> (node status cache)</div>
                <div><strong>React</strong> + Ant Design + Recharts</div>
              </div>
            </div>
          </div>
        </Card>

        {/* Key Concepts */}
        <Card style={{ marginBottom: 20 }}>
          <Title level={4} style={{ color: '#21325b', marginBottom: 16 }}>
            <AuditOutlined style={{ marginRight: 8, color: '#d48806' }} />
            Key Concepts
          </Title>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(280px, 1fr))', gap: 12 }}>
            {[
              { term: 'RUID', full: 'Registration UID', desc: 'Unique identifier for each copyright claim submitted to the chain' },
              { term: 'AUID', full: 'Asset UID', desc: 'Unique identifier for a digital asset (e.g., a creative work)' },
              { term: 'PUID', full: 'Proprietor UID', desc: 'Unique identifier for the copyright proprietor / owner' },
              { term: 'Batch', full: 'Merkle Batch', desc: 'A group of RUIDs aggregated into a Merkle Tree and anchored on-chain' },
              { term: 'OTS', full: 'OpenTimestamps', desc: 'Open protocol for creating Bitcoin-backed timestamps' },
              { term: 'Leaf Index', full: 'Merkle Leaf Position', desc: 'Position of a RUID within its batch (0 = earliest submitted)' },
            ].map(item => (
              <div key={item.term} style={{ padding: '12px 16px', borderRadius: 8, background: '#fafafa', border: '1px solid #f0f0f0' }}>
                <div>
                  <Tag color="blue" style={{ fontSize: 13, fontWeight: 600 }}>{item.term}</Tag>
                  <Text type="secondary" style={{ fontSize: 12 }}>{item.full}</Text>
                </div>
                <div style={{ color: '#555', fontSize: 13, marginTop: 6 }}>{item.desc}</div>
              </div>
            ))}
          </div>
        </Card>

        {/* Footer brand */}
        <div style={{ textAlign: 'center', padding: '16px 0 8px', color: '#999', fontSize: 13 }}>
          <TeamOutlined style={{ marginRight: 6 }} />
          Built by <strong>Redmansion</strong>
        </div>
      </div>
    </div>
  );
}
