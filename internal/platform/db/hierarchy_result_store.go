package db

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/kineticbot/kinetic-v3/internal/domain/hierarchy"
)

const resultSubmissionColumns = `
	id,
	context_type,
	context_id,
	game_key,
	submitted_by_team_id,
	submitted_by_subject,
	submitted_by_display_name,
	home_team_id,
	away_team_id,
	winning_team_id,
	losing_team_id,
	state,
	payload_json,
	payload_digest,
	provenance_json,
	home_ratified_at,
	home_ratified_by_subject,
	home_ratified_by_display_name,
	away_ratified_at,
	away_ratified_by_subject,
	away_ratified_by_display_name,
	rejected_by_team_id,
	rejection_reason,
	rejected_at,
	created_at`

type rowScanner interface {
	Scan(dest ...any) error
}

func scanResultSubmission(scanner rowScanner, submission *hierarchy.ResultSubmission) error {
	return scanner.Scan(
		&submission.ID,
		&submission.ContextType,
		&submission.ContextID,
		&submission.GameKey,
		&submission.SubmittedByTeamID,
		&submission.SubmittedBySubject,
		&submission.SubmittedByDisplayName,
		&submission.HomeTeamID,
		&submission.AwayTeamID,
		&submission.WinningTeamID,
		&submission.LosingTeamID,
		&submission.State,
		&submission.PayloadJSON,
		&submission.PayloadDigest,
		&submission.ProvenanceJSON,
		&submission.HomeRatifiedAt,
		&submission.HomeRatifiedBySubject,
		&submission.HomeRatifiedByDisplayName,
		&submission.AwayRatifiedAt,
		&submission.AwayRatifiedBySubject,
		&submission.AwayRatifiedByDisplayName,
		&submission.RejectedByTeamID,
		&submission.RejectionReason,
		&submission.RejectedAt,
		&submission.CreatedAt,
	)
}

func (s *HierarchyStore) CreateResultSubmission(ctx context.Context, input hierarchy.CreateResultSubmissionInput) (hierarchy.ResultSubmission, error) {
	if err := hierarchy.ValidateCreateResultSubmissionInput(input); err != nil {
		return hierarchy.ResultSubmission{}, err
	}

	homeTeamID, awayTeamID, err := resolveContextTeams(ctx, s.db, input.ContextType, input.ContextID)
	if err != nil {
		return hierarchy.ResultSubmission{}, err
	}
	if input.SubmittedByTeamID != homeTeamID && input.SubmittedByTeamID != awayTeamID {
		return hierarchy.ResultSubmission{}, fmt.Errorf("%w: submittedByTeamId must be a context participant", hierarchy.ErrConflict)
	}
	if (input.WinningTeamID != homeTeamID && input.WinningTeamID != awayTeamID) ||
		(input.LosingTeamID != homeTeamID && input.LosingTeamID != awayTeamID) {
		return hierarchy.ResultSubmission{}, fmt.Errorf("%w: winning and losing teams must match context participants", hierarchy.ErrConflict)
	}

	payload := normalizeJSON(input.PayloadJSON)
	if score, ok, err := hierarchy.ExtractRocketLeagueScore(payload); err != nil {
		return hierarchy.ResultSubmission{}, err
	} else if ok {
		expectedWinner := awayTeamID
		if score.Home > score.Away {
			expectedWinner = homeTeamID
		}
		if input.WinningTeamID != expectedWinner {
			return hierarchy.ResultSubmission{}, fmt.Errorf("%w: winningTeamId must match payload score", hierarchy.ErrInvalidInput)
		}
	}
	provenance := normalizeJSON(input.ProvenanceJSON)
	gameKey := hierarchy.NormalizeGameKey(input.GameKey)
	payloadDigest := digestJSON(payload)

	const stmt = `
INSERT INTO result_submissions(
	context_type,
	context_id,
	game_key,
	submitted_by_team_id,
	submitted_by_subject,
	submitted_by_display_name,
	home_team_id,
	away_team_id,
	winning_team_id,
	losing_team_id,
	payload_json,
	payload_digest,
	provenance_json,
	home_ratified_at,
	home_ratified_by_subject,
	home_ratified_by_display_name,
	away_ratified_at,
	away_ratified_by_subject,
	away_ratified_by_display_name
)
VALUES (
	$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13,
	CASE WHEN $4 = $7 THEN NOW() ELSE NULL END,
	CASE WHEN $4 = $7 THEN $5 ELSE NULL END,
	CASE WHEN $4 = $7 THEN $6 ELSE NULL END,
	CASE WHEN $4 = $8 THEN NOW() ELSE NULL END,
	CASE WHEN $4 = $8 THEN $5 ELSE NULL END,
	CASE WHEN $4 = $8 THEN $6 ELSE NULL END
)
RETURNING` + resultSubmissionColumns + `;`

	var submission hierarchy.ResultSubmission
	err = scanResultSubmission(s.db.QueryRowContext(
		ctx,
		stmt,
		input.ContextType,
		input.ContextID,
		gameKey,
		input.SubmittedByTeamID,
		strings.TrimSpace(input.SubmittedBySubject),
		strings.TrimSpace(input.SubmittedByDisplayName),
		homeTeamID,
		awayTeamID,
		input.WinningTeamID,
		input.LosingTeamID,
		payload,
		payloadDigest,
		provenance,
	), &submission)
	if err != nil {
		return hierarchy.ResultSubmission{}, mapSQLError(err)
	}
	return submission, nil
}

func normalizeJSON(raw json.RawMessage) json.RawMessage {
	if len(raw) == 0 {
		return json.RawMessage(`{}`)
	}
	var buffer bytes.Buffer
	if err := json.Compact(&buffer, raw); err != nil {
		return raw
	}
	return json.RawMessage(buffer.Bytes())
}

func digestJSON(raw json.RawMessage) string {
	normalized := normalizeJSON(raw)
	sum := sha256.Sum256(normalized)
	return hex.EncodeToString(sum[:])
}

func (s *HierarchyStore) ListResultSubmissions(ctx context.Context) ([]hierarchy.ResultSubmission, error) {
	const stmt = `
SELECT` + resultSubmissionColumns + `
FROM result_submissions
ORDER BY id ASC;`
	rows, err := s.db.QueryContext(ctx, stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	submissions := make([]hierarchy.ResultSubmission, 0)
	for rows.Next() {
		var submission hierarchy.ResultSubmission
		if err := scanResultSubmission(rows, &submission); err != nil {
			return nil, err
		}
		submissions = append(submissions, submission)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return submissions, nil
}

func (s *HierarchyStore) OverrideResultSubmission(ctx context.Context, input hierarchy.OverrideResultSubmissionInput) (hierarchy.ResultSubmission, error) {
	if err := hierarchy.ValidateOverrideResultSubmissionInput(input); err != nil {
		return hierarchy.ResultSubmission{}, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return hierarchy.ResultSubmission{}, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	const lockStmt = `
SELECT
	home_team_id,
	away_team_id,
	winning_team_id,
	losing_team_id,
	state
FROM result_submissions
WHERE id = $1
FOR UPDATE;`
	var homeTeamID, awayTeamID int64
	var previousWinningTeamID, previousLosingTeamID int64
	var previousState string
	if err := tx.QueryRowContext(ctx, lockStmt, input.SubmissionID).Scan(
		&homeTeamID,
		&awayTeamID,
		&previousWinningTeamID,
		&previousLosingTeamID,
		&previousState,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return hierarchy.ResultSubmission{}, fmt.Errorf("%w: submission not found", hierarchy.ErrConflict)
		}
		return hierarchy.ResultSubmission{}, err
	}

	if (input.WinningTeamID != homeTeamID && input.WinningTeamID != awayTeamID) ||
		(input.LosingTeamID != homeTeamID && input.LosingTeamID != awayTeamID) {
		return hierarchy.ResultSubmission{}, fmt.Errorf("%w: winning and losing teams must match submission participants", hierarchy.ErrConflict)
	}

	const overrideStmt = `
UPDATE result_submissions
SET
	winning_team_id = $2,
	losing_team_id = $3,
	state = 'ratified',
	home_ratified_at = COALESCE(home_ratified_at, NOW()),
	away_ratified_at = COALESCE(away_ratified_at, NOW()),
	rejected_by_team_id = NULL,
	rejection_reason = NULL,
	rejected_at = NULL
WHERE id = $1
RETURNING` + resultSubmissionColumns + `;`
	var submission hierarchy.ResultSubmission
	if err := scanResultSubmission(tx.QueryRowContext(ctx, overrideStmt, input.SubmissionID, input.WinningTeamID, input.LosingTeamID), &submission); err != nil {
		return hierarchy.ResultSubmission{}, mapSQLError(err)
	}

	const auditStmt = `
INSERT INTO result_overrides(
	submission_id,
	actor,
	reason,
	previous_winning_team_id,
	previous_losing_team_id,
	new_winning_team_id,
	new_losing_team_id,
	previous_state,
	new_state
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);`
	if _, err := tx.ExecContext(
		ctx,
		auditStmt,
		input.SubmissionID,
		input.Actor,
		input.Reason,
		previousWinningTeamID,
		previousLosingTeamID,
		input.WinningTeamID,
		input.LosingTeamID,
		previousState,
		"ratified",
	); err != nil {
		return hierarchy.ResultSubmission{}, mapSQLError(err)
	}

	if err := tx.Commit(); err != nil {
		return hierarchy.ResultSubmission{}, err
	}
	return submission, nil
}

func (s *HierarchyStore) ListResultOverrides(ctx context.Context) ([]hierarchy.ResultOverride, error) {
	const stmt = `
SELECT
	id,
	submission_id,
	actor,
	reason,
	previous_winning_team_id,
	previous_losing_team_id,
	new_winning_team_id,
	new_losing_team_id,
	previous_state,
	new_state,
	created_at
FROM result_overrides
ORDER BY id DESC;`
	rows, err := s.db.QueryContext(ctx, stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	overrides := make([]hierarchy.ResultOverride, 0)
	for rows.Next() {
		var override hierarchy.ResultOverride
		if err := rows.Scan(
			&override.ID,
			&override.SubmissionID,
			&override.Actor,
			&override.Reason,
			&override.PreviousWinningTeamID,
			&override.PreviousLosingTeamID,
			&override.NewWinningTeamID,
			&override.NewLosingTeamID,
			&override.PreviousState,
			&override.NewState,
			&override.CreatedAt,
		); err != nil {
			return nil, err
		}
		overrides = append(overrides, override)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return overrides, nil
}

func (s *HierarchyStore) RatifyResultSubmission(ctx context.Context, input hierarchy.RatifyResultSubmissionInput) (hierarchy.ResultSubmission, error) {
	if err := hierarchy.ValidateRatifyResultSubmissionInput(input); err != nil {
		return hierarchy.ResultSubmission{}, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return hierarchy.ResultSubmission{}, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	const lockStmt = `
SELECT home_team_id, away_team_id, submitted_by_team_id, submitted_by_subject, state, home_ratified_at, away_ratified_at
FROM result_submissions
WHERE id = $1
FOR UPDATE;`
	var homeTeamID, awayTeamID, submittedByTeamID int64
	var submittedBySubject string
	var state string
	var homeRatifiedAt, awayRatifiedAt *time.Time
	if err := tx.QueryRowContext(ctx, lockStmt, input.SubmissionID).Scan(
		&homeTeamID,
		&awayTeamID,
		&submittedByTeamID,
		&submittedBySubject,
		&state,
		&homeRatifiedAt,
		&awayRatifiedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return hierarchy.ResultSubmission{}, fmt.Errorf("%w: submission not found", hierarchy.ErrConflict)
		}
		return hierarchy.ResultSubmission{}, err
	}
	if state != "pending" {
		return hierarchy.ResultSubmission{}, fmt.Errorf("%w: only pending submissions can be ratified", hierarchy.ErrConflict)
	}
	if input.TeamID != homeTeamID && input.TeamID != awayTeamID {
		return hierarchy.ResultSubmission{}, fmt.Errorf("%w: ratifying team must be a participant", hierarchy.ErrConflict)
	}
	if input.TeamID == submittedByTeamID {
		return hierarchy.ResultSubmission{}, fmt.Errorf("%w: submitting team is already attested", hierarchy.ErrConflict)
	}
	if strings.EqualFold(strings.TrimSpace(input.RatifiedBySubject), strings.TrimSpace(submittedBySubject)) {
		return hierarchy.ResultSubmission{}, fmt.Errorf("%w: ratifying subject must differ from submitter", hierarchy.ErrConflict)
	}
	if (input.TeamID == homeTeamID && homeRatifiedAt != nil) || (input.TeamID == awayTeamID && awayRatifiedAt != nil) {
		return hierarchy.ResultSubmission{}, fmt.Errorf("%w: team already ratified", hierarchy.ErrConflict)
	}

	const ratifyStmt = `
UPDATE result_submissions
SET
	home_ratified_at = CASE
		WHEN $2 = home_team_id AND home_ratified_at IS NULL THEN NOW()
		ELSE home_ratified_at
	END,
	home_ratified_by_subject = CASE
		WHEN $2 = home_team_id AND home_ratified_by_subject IS NULL THEN $3
		ELSE home_ratified_by_subject
	END,
	home_ratified_by_display_name = CASE
		WHEN $2 = home_team_id AND home_ratified_by_display_name IS NULL THEN $4
		ELSE home_ratified_by_display_name
	END,
	away_ratified_at = CASE
		WHEN $2 = away_team_id AND away_ratified_at IS NULL THEN NOW()
		ELSE away_ratified_at
	END,
	away_ratified_by_subject = CASE
		WHEN $2 = away_team_id AND away_ratified_by_subject IS NULL THEN $3
		ELSE away_ratified_by_subject
	END,
	away_ratified_by_display_name = CASE
		WHEN $2 = away_team_id AND away_ratified_by_display_name IS NULL THEN $4
		ELSE away_ratified_by_display_name
	END,
	state = CASE
		WHEN
			(CASE WHEN $2 = home_team_id AND home_ratified_at IS NULL THEN NOW() ELSE home_ratified_at END) IS NOT NULL
			AND
			(CASE WHEN $2 = away_team_id AND away_ratified_at IS NULL THEN NOW() ELSE away_ratified_at END) IS NOT NULL
		THEN 'ratified'
		ELSE state
	END
WHERE id = $1
RETURNING` + resultSubmissionColumns + `;`
	var submission hierarchy.ResultSubmission
	if err := scanResultSubmission(
		tx.QueryRowContext(
			ctx,
			ratifyStmt,
			input.SubmissionID,
			input.TeamID,
			strings.TrimSpace(input.RatifiedBySubject),
			strings.TrimSpace(input.RatifiedByDisplayName),
		),
		&submission,
	); err != nil {
		return hierarchy.ResultSubmission{}, mapSQLError(err)
	}

	// When both sides have ratified, trigger automated Glicko-2 rating updates.
	if submission.State == "ratified" {
		if err := s.applyRatingUpdatesInTx(ctx, tx, submission); err != nil {
			return hierarchy.ResultSubmission{}, fmt.Errorf("rating update failed: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return hierarchy.ResultSubmission{}, err
	}
	return submission, nil
}

func (s *HierarchyStore) RejectResultSubmission(ctx context.Context, input hierarchy.RejectResultSubmissionInput) (hierarchy.ResultSubmission, error) {
	if err := hierarchy.ValidateRejectResultSubmissionInput(input); err != nil {
		return hierarchy.ResultSubmission{}, err
	}

	const stmt = `
UPDATE result_submissions
SET
	state = 'rejected',
	rejected_by_team_id = $2,
	rejection_reason = $3,
	rejected_at = NOW()
WHERE id = $1
  AND state = 'pending'
  AND ($2 = home_team_id OR $2 = away_team_id)
RETURNING` + resultSubmissionColumns + `;`
	var submission hierarchy.ResultSubmission
	err := scanResultSubmission(s.db.QueryRowContext(ctx, stmt, input.SubmissionID, input.TeamID, input.Reason), &submission)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return hierarchy.ResultSubmission{}, fmt.Errorf("%w: submission reject not allowed or not found", hierarchy.ErrConflict)
		}
		return hierarchy.ResultSubmission{}, mapSQLError(err)
	}
	return submission, nil
}

func resolveContextTeams(ctx context.Context, queryer interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}, contextType string, contextID int64) (int64, int64, error) {
	var stmt string
	switch contextType {
	case "scrim":
		stmt = `SELECT home_team_id, away_team_id FROM scrims WHERE id = $1;`
	case "match":
		stmt = `SELECT home_team_id, away_team_id FROM matches WHERE id = $1;`
	default:
		return 0, 0, fmt.Errorf("%w: unsupported context type", hierarchy.ErrInvalidInput)
	}

	var homeTeamID, awayTeamID int64
	if err := queryer.QueryRowContext(ctx, stmt, contextID).Scan(&homeTeamID, &awayTeamID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, 0, fmt.Errorf("%w: context not found", hierarchy.ErrDependency)
		}
		return 0, 0, err
	}
	return homeTeamID, awayTeamID, nil
}

func (s *HierarchyStore) GetResultSubmission(ctx context.Context, submissionID int64) (hierarchy.ResultSubmission, error) {
	const stmt = `
SELECT` + resultSubmissionColumns + `
FROM result_submissions
WHERE id = $1;`
	var submission hierarchy.ResultSubmission
	err := scanResultSubmission(s.db.QueryRowContext(ctx, stmt, submissionID), &submission)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return hierarchy.ResultSubmission{}, fmt.Errorf("%w: result submission not found", hierarchy.ErrDependency)
		}
		return hierarchy.ResultSubmission{}, mapSQLError(err)
	}
	return submission, nil
}

func (s *HierarchyStore) ListResultSubmissionsFiltered(ctx context.Context, input hierarchy.ListResultSubmissionsInput) ([]hierarchy.ResultSubmission, error) {
	query := `
SELECT` + resultSubmissionColumns + `
FROM result_submissions`

	args := make([]any, 0)
	conditions := make([]string, 0)

	if input.State != nil {
		args = append(args, *input.State)
		conditions = append(conditions, fmt.Sprintf("state = $%d", len(args)))
	}
	if input.TeamID != nil {
		args = append(args, *input.TeamID)
		conditions = append(conditions, fmt.Sprintf("(home_team_id = $%d OR away_team_id = $%d)", len(args), len(args)))
	}

	if len(conditions) > 0 {
		query += "\nWHERE " + conditions[0]
		for i := 1; i < len(conditions); i++ {
			query += "\n  AND " + conditions[i]
		}
	}
	query += "\nORDER BY id ASC;"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	submissions := make([]hierarchy.ResultSubmission, 0)
	for rows.Next() {
		var submission hierarchy.ResultSubmission
		if err := scanResultSubmission(rows, &submission); err != nil {
			return nil, err
		}
		submissions = append(submissions, submission)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return submissions, nil
}

func (s *HierarchyStore) ResetResultSubmission(ctx context.Context, input hierarchy.ResetResultSubmissionInput) (hierarchy.ResultSubmission, error) {
	const stmt = `
UPDATE result_submissions
SET
	state = 'pending',
	rejected_by_team_id = NULL,
	rejection_reason = NULL,
	rejected_at = NULL,
	home_ratified_at = CASE WHEN submitted_by_team_id = home_team_id THEN NOW() ELSE NULL END,
	home_ratified_by_subject = CASE WHEN submitted_by_team_id = home_team_id THEN submitted_by_subject ELSE NULL END,
	home_ratified_by_display_name = CASE WHEN submitted_by_team_id = home_team_id THEN submitted_by_display_name ELSE NULL END,
	away_ratified_at = CASE WHEN submitted_by_team_id = away_team_id THEN NOW() ELSE NULL END,
	away_ratified_by_subject = CASE WHEN submitted_by_team_id = away_team_id THEN submitted_by_subject ELSE NULL END,
	away_ratified_by_display_name = CASE WHEN submitted_by_team_id = away_team_id THEN submitted_by_display_name ELSE NULL END
WHERE id = $1
  AND state = 'rejected'
RETURNING` + resultSubmissionColumns + `;`
	var submission hierarchy.ResultSubmission
	err := scanResultSubmission(s.db.QueryRowContext(ctx, stmt, input.SubmissionID), &submission)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return hierarchy.ResultSubmission{}, fmt.Errorf("%w: submission reset not allowed or not found", hierarchy.ErrConflict)
		}
		return hierarchy.ResultSubmission{}, mapSQLError(err)
	}
	return submission, nil
}
