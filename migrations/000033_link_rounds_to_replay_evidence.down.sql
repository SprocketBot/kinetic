ALTER TABLE rounds DROP CONSTRAINT IF EXISTS uq_rounds_replay_evidence;
ALTER TABLE rounds DROP COLUMN IF EXISTS replay_evidence_id;
