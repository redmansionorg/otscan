-- OTScan Phase 2 schema

CREATE TABLE IF NOT EXISTS nodes (
    id         SERIAL PRIMARY KEY,
    name       VARCHAR(64) NOT NULL UNIQUE,
    rpc_url    VARCHAR(255) NOT NULL,
    is_active  BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS node_status (
    id                   SERIAL PRIMARY KEY,
    node_id              INTEGER REFERENCES nodes(id) ON DELETE CASCADE,
    status               VARCHAR(32) NOT NULL DEFAULT 'unknown',
    block_number         BIGINT DEFAULT 0,
    ots_mode             VARCHAR(16),
    pending_count        INTEGER DEFAULT 0,
    total_created        INTEGER DEFAULT 0,
    total_confirmed      INTEGER DEFAULT 0,
    last_processed_block BIGINT DEFAULT 0,
    components           JSONB,
    last_anchor          TEXT,
    polled_at            TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(node_id)
);

CREATE TABLE IF NOT EXISTS node_status_history (
    id            SERIAL PRIMARY KEY,
    node_id       INTEGER REFERENCES nodes(id) ON DELETE CASCADE,
    block_number  BIGINT,
    pending_count INTEGER,
    status        VARCHAR(32),
    recorded_at   TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_nsh_node_time ON node_status_history(node_id, recorded_at DESC);

CREATE TABLE IF NOT EXISTS batches (
    id               SERIAL PRIMARY KEY,
    batch_id         VARCHAR(66) NOT NULL UNIQUE,
    on_chain_id      BIGINT,
    start_block      BIGINT NOT NULL,
    end_block        BIGINT NOT NULL,
    root_hash        VARCHAR(66) NOT NULL DEFAULT '',
    ots_digest       VARCHAR(66),
    ruid_count       INTEGER DEFAULT 0,
    trigger_type     VARCHAR(16),
    status           VARCHAR(16) NOT NULL,
    btc_tx_id        VARCHAR(66),
    btc_block_height BIGINT DEFAULT 0,
    btc_timestamp    BIGINT DEFAULT 0,
    created_at       BIGINT DEFAULT 0,
    updated_at       TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_batches_status ON batches(status);
CREATE INDEX IF NOT EXISTS idx_batches_on_chain_id ON batches(on_chain_id);

CREATE TABLE IF NOT EXISTS batch_ruids (
    id       SERIAL PRIMARY KEY,
    batch_id VARCHAR(66) NOT NULL REFERENCES batches(batch_id) ON DELETE CASCADE,
    ruid     VARCHAR(66) NOT NULL,
    UNIQUE(batch_id, ruid)
);
CREATE INDEX IF NOT EXISTS idx_batch_ruids_ruid ON batch_ruids(ruid);

CREATE TABLE IF NOT EXISTS claims (
    id            SERIAL PRIMARY KEY,
    ruid          VARCHAR(66) NOT NULL UNIQUE,
    claimant      VARCHAR(42),
    submit_block  BIGINT DEFAULT 0,
    submit_time   BIGINT DEFAULT 0,
    published     BOOLEAN DEFAULT false,
    auid          VARCHAR(66),
    puid          VARCHAR(66),
    publish_block BIGINT DEFAULT 0,
    publish_time  BIGINT DEFAULT 0,
    batch_id      VARCHAR(66),
    created_at    TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_claims_claimant ON claims(claimant);
CREATE INDEX IF NOT EXISTS idx_claims_auid ON claims(auid);
CREATE INDEX IF NOT EXISTS idx_claims_puid ON claims(puid);
CREATE INDEX IF NOT EXISTS idx_claims_batch ON claims(batch_id);
