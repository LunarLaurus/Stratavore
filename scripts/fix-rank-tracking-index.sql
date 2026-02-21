-- Fix: Add missing unique index for rank_tracking idempotency
-- This index prevents duplicate rank events during V2→V3 sync

-- Drop existing if present (idempotent)
DROP INDEX IF EXISTS idx_rank_tracking_unique_event;

-- Create unique index on event_type + event_date + description
-- This ensures ON CONFLICT DO NOTHING works in sync scripts
CREATE UNIQUE INDEX idx_rank_tracking_unique_event
ON rank_tracking(event_type, event_date, COALESCE(description, ''));

-- Verify creation
SELECT
    indexname,
    indexdef
FROM pg_indexes
WHERE tablename = 'rank_tracking'
  AND indexname = 'idx_rank_tracking_unique_event';
