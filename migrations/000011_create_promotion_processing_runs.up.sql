CREATE TABLE IF NOT EXISTS promotion_processing_runs (
    id BIGSERIAL PRIMARY KEY,
    queue_id BIGINT NULL REFERENCES queues(id) ON DELETE SET NULL,
    processed_queues INT NOT NULL,
    promotions_created INT NOT NULL,
    conflicts INT NOT NULL,
    duration_ms INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
