ALTER TABLE rounds
    ADD COLUMN replay_evidence_id BIGINT REFERENCES replay_evidence(id) ON DELETE RESTRICT;

ALTER TABLE rounds
    ADD CONSTRAINT uq_rounds_replay_evidence UNIQUE (replay_evidence_id);
