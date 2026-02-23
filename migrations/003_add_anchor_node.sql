-- Add anchor node tracking columns to batches table
ALTER TABLE batches ADD COLUMN IF NOT EXISTS anchored_by VARCHAR(42);
ALTER TABLE batches ADD COLUMN IF NOT EXISTS anchor_block BIGINT DEFAULT 0;
