-- 002_add_conflict_indexes.sql
-- Phase 2: conflict detection indexes and sync progress tracking

-- Accelerate conflict scanning (find AUIDs with multiple PUIDs)
CREATE INDEX IF NOT EXISTS idx_claims_auid_puid
    ON claims(auid, puid) WHERE published = true;

-- Accelerate time-range queries
CREATE INDEX IF NOT EXISTS idx_claims_submit_block
    ON claims(submit_block);

-- Sync progress metadata table
CREATE TABLE IF NOT EXISTS sync_meta (
    key        VARCHAR(64) PRIMARY KEY,
    value      VARCHAR(255) NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
