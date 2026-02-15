CREATE TABLE IF NOT EXISTS exception_tickets (
    id BIGSERIAL PRIMARY KEY,
    category TEXT NOT NULL,
    context_type TEXT NOT NULL,
    context_id BIGINT NOT NULL,
    reported_by_team_id BIGINT NULL REFERENCES teams(id) ON DELETE SET NULL,
    state TEXT NOT NULL DEFAULT 'open',
    reason_code TEXT NOT NULL,
    severity INT NOT NULL DEFAULT 3,
    suggested_action TEXT NOT NULL,
    details_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    resolution_code TEXT NULL,
    opened_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    triaged_at TIMESTAMPTZ NULL,
    resolved_at TIMESTAMPTZ NULL,
    CONSTRAINT ck_exception_tickets_category_allowed CHECK (
        category IN (
            'scheduling_conflict',
            'no_show',
            'forfeit',
            'result_dispute',
            'replay_parse_failure',
            'replay_identity_mismatch',
            'roster_eligibility'
        )
    ),
    CONSTRAINT ck_exception_tickets_context_type_allowed CHECK (
        context_type IN ('match', 'scrim', 'result_submission', 'replay_evidence')
    ),
    CONSTRAINT ck_exception_tickets_state_allowed CHECK (
        state IN ('open', 'triaged', 'resolved')
    ),
    CONSTRAINT ck_exception_tickets_severity_range CHECK (severity >= 1 AND severity <= 5)
);

CREATE INDEX IF NOT EXISTS ix_exception_tickets_inbox
    ON exception_tickets(state, severity DESC, opened_at ASC);

CREATE INDEX IF NOT EXISTS ix_exception_tickets_context
    ON exception_tickets(context_type, context_id);

CREATE TABLE IF NOT EXISTS exception_actions (
    id BIGSERIAL PRIMARY KEY,
    ticket_id BIGINT NOT NULL REFERENCES exception_tickets(id) ON DELETE CASCADE,
    action_type TEXT NOT NULL,
    actor TEXT NOT NULL,
    automated BOOLEAN NOT NULL DEFAULT FALSE,
    notes TEXT NOT NULL DEFAULT '',
    minutes_spent INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT ck_exception_actions_type_allowed CHECK (
        action_type IN ('reported', 'triaged', 'resolved', 'auto_resolved', 'override')
    ),
    CONSTRAINT ck_exception_actions_minutes_nonnegative CHECK (minutes_spent >= 0)
);

CREATE INDEX IF NOT EXISTS ix_exception_actions_ticket
    ON exception_actions(ticket_id, created_at ASC);

CREATE INDEX IF NOT EXISTS ix_exception_actions_created
    ON exception_actions(created_at DESC);
