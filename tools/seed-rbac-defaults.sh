#!/usr/bin/env bash
set -euo pipefail

# Seeds the stable RBAC vocabulary and baseline role-to-capability matrix.
#
# Optional global grants are comma-separated authenticated subjects, for example:
#   DATABASE_URL=... ADMIN_SUBJECTS='discord:123,discord:456' \
#     OPERATOR_SUBJECTS='discord:789' ./tools/seed-rbac-defaults.sh
#
# Scoped roles (fm, gm, agm, captain) are deliberately not granted here. They
# belong in role_assignments and require a real player plus franchise/club/team
# scope, which prevents this bootstrap script from creating overbroad access.

: "${DATABASE_URL:?DATABASE_URL is required}"

admin_subjects="${ADMIN_SUBJECTS:-}"
operator_subjects="${OPERATOR_SUBJECTS:-}"

psql "$DATABASE_URL" \
  --set=ON_ERROR_STOP=1 \
  --set=admin_subjects="$admin_subjects" \
  --set=operator_subjects="$operator_subjects" <<'SQL'
BEGIN;

INSERT INTO roles(name) VALUES
  ('admin'), ('operator'), ('observer'), ('player'),
  ('fm'), ('gm'), ('agm'), ('captain')
ON CONFLICT (name) DO NOTHING;

-- Global platform authority. Only grant these through user_role_bindings.
INSERT INTO policies(role_name, resource, action) VALUES
  ('admin', '*', '*'),
  ('operator', '*', '*')
ON CONFLICT (role_name, resource, action) DO NOTHING;

-- Read-only access is intentionally broad; resource-specific visibility still
-- belongs at query boundaries where private data is introduced.
INSERT INTO policies(role_name, resource, action)
SELECT 'observer', resource, 'read'
FROM (VALUES
  ('league'), ('franchise'), ('club'), ('team'), ('player'),
  ('roster_membership'), ('role_assignment'), ('queue'), ('queue_entry'),
  ('queue_ban'), ('scrim'), ('result_submission'), ('result_override'),
  ('replay_evidence'), ('platform_account_link'), ('player_rating'),
  ('rating_adjustment'), ('matchmaking_decision'), ('exception_ticket'),
  ('season'), ('schedule_group'), ('fixture'), ('match'), ('eligibility'),
  ('player_stat'), ('skill_group'), ('skill_group_transition'),
  ('organization_config'), ('player_notification'), ('game')
) AS resources(resource)
ON CONFLICT (role_name, resource, action) DO NOTHING;

-- Player actions require the authenticated user's resolved game player at the
-- HTTP boundary; this table intentionally only describes capability.
INSERT INTO policies(role_name, resource, action) VALUES
  ('player', 'scrim', 'update'),
  ('player', 'result_submission', 'create'),
  ('player', 'replay_evidence', 'create')
ON CONFLICT (role_name, resource, action) DO NOTHING;

-- FM, GM, AGM, and Captain rows are scoped by role_assignments. Their resource
-- matrix is data-driven; the franchise/club/team containment check remains in
-- code so a policy edit cannot broaden a scope accidentally.
INSERT INTO policies(role_name, resource, action) VALUES
  ('fm', 'role_assignment', 'create'), ('fm', 'role_assignment', 'revoke'),
  ('fm', 'roster_membership', 'create'), ('fm', 'roster_membership', 'delete'),
  ('fm', 'result_override', 'create'), ('fm', 'result_submission', 'ratify'),
  ('fm', 'result_submission', 'reject'), ('fm', 'result_submission', 'reset'),
  ('fm', 'player', 'update'),
  ('gm', 'role_assignment', 'create'), ('gm', 'role_assignment', 'revoke'),
  ('gm', 'roster_membership', 'create'), ('gm', 'roster_membership', 'delete'),
  ('gm', 'result_submission', 'ratify'), ('gm', 'result_submission', 'reject'),
  ('agm', 'roster_membership', 'create'), ('agm', 'roster_membership', 'delete'),
  ('agm', 'result_submission', 'ratify'), ('agm', 'result_submission', 'reject'),
  ('captain', 'role_assignment', 'create'), ('captain', 'role_assignment', 'revoke'),
  ('captain', 'result_submission', 'ratify'), ('captain', 'result_submission', 'reject'),
  ('captain', 'scrim', 'create')
ON CONFLICT (role_name, resource, action) DO NOTHING;

INSERT INTO user_role_bindings(subject, role_name)
SELECT btrim(subject), 'admin'
FROM unnest(string_to_array(:'admin_subjects', ',')) AS raw(subject)
WHERE btrim(subject) <> ''
ON CONFLICT (subject, role_name) DO NOTHING;

INSERT INTO user_role_bindings(subject, role_name)
SELECT btrim(subject), 'operator'
FROM unnest(string_to_array(:'operator_subjects', ',')) AS raw(subject)
WHERE btrim(subject) <> ''
ON CONFLICT (subject, role_name) DO NOTHING;

COMMIT;
SQL
