package db

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/kineticbot/kinetic-v3/internal/domain/hierarchy"
)

func (s *HierarchyStore) IngestReplayEvidence(ctx context.Context, input hierarchy.IngestReplayEvidenceInput) (hierarchy.ReplayIngestionResult, error) {
	if err := hierarchy.ValidateIngestReplayEvidenceInput(input); err != nil {
		return hierarchy.ReplayIngestionResult{}, err
	}

	homeTeamID, awayTeamID, err := resolveContextTeams(ctx, s.db, input.ContextType, input.ContextID)
	if err != nil {
		return hierarchy.ReplayIngestionResult{}, err
	}
	if input.SubmittedByTeamID != homeTeamID && input.SubmittedByTeamID != awayTeamID {
		return hierarchy.ReplayIngestionResult{}, fmt.Errorf("%w: submittedByTeamId must be a context participant", hierarchy.ErrConflict)
	}

	contentHash := sha256.Sum256([]byte(input.ReplayBody))
	replaySHA256 := hex.EncodeToString(contentHash[:])
	contentSizeBytes := int64(len(input.ReplayBody))
	output := input.ParseOutputJSON
	if len(output) == 0 {
		output = []byte(`{}`)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return hierarchy.ReplayIngestionResult{}, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	result := hierarchy.ReplayIngestionResult{}

	const findEvidenceStmt = `
SELECT id, context_type, context_id, submitted_by_team_id, replay_sha256, content_size_bytes, storage_ref, state, created_at
FROM replay_evidence
WHERE replay_sha256 = $1
FOR UPDATE;`
	err = tx.QueryRowContext(ctx, findEvidenceStmt, replaySHA256).Scan(
		&result.Evidence.ID,
		&result.Evidence.ContextType,
		&result.Evidence.ContextID,
		&result.Evidence.SubmittedByTeamID,
		&result.Evidence.ReplaySHA256,
		&result.Evidence.ContentSizeBytes,
		&result.Evidence.StorageRef,
		&result.Evidence.State,
		&result.Evidence.CreatedAt,
	)
	switch {
	case err == nil:
		if result.Evidence.ContextType != input.ContextType || result.Evidence.ContextID != input.ContextID {
			return hierarchy.ReplayIngestionResult{}, fmt.Errorf("%w: replay hash already attached to another context", hierarchy.ErrConflict)
		}
		result.Duplicate = true
	case errors.Is(err, sql.ErrNoRows):
		const createEvidenceStmt = `
INSERT INTO replay_evidence(
	context_type,
	context_id,
	submitted_by_team_id,
	replay_sha256,
	content_size_bytes,
	storage_ref,
	state
)
VALUES ($1, $2, $3, $4, $5, $6, 'parsed')
RETURNING id, context_type, context_id, submitted_by_team_id, replay_sha256, content_size_bytes, storage_ref, state, created_at;`
		if err := tx.QueryRowContext(
			ctx,
			createEvidenceStmt,
			input.ContextType,
			input.ContextID,
			input.SubmittedByTeamID,
			replaySHA256,
			contentSizeBytes,
			fmt.Sprintf("inline-sha256:%s", replaySHA256),
		).Scan(
			&result.Evidence.ID,
			&result.Evidence.ContextType,
			&result.Evidence.ContextID,
			&result.Evidence.SubmittedByTeamID,
			&result.Evidence.ReplaySHA256,
			&result.Evidence.ContentSizeBytes,
			&result.Evidence.StorageRef,
			&result.Evidence.State,
			&result.Evidence.CreatedAt,
		); err != nil {
			return hierarchy.ReplayIngestionResult{}, mapSQLError(err)
		}
	default:
		return hierarchy.ReplayIngestionResult{}, err
	}

	const createParseRunStmt = `
INSERT INTO replay_parse_runs(
	replay_evidence_id,
	parser_name,
	parser_version,
	parser_config_digest,
	status,
	output_json
)
VALUES ($1, $2, $3, $4, 'parsed', $5)
RETURNING id, replay_evidence_id, parser_name, parser_version, parser_config_digest, status, output_json, created_at;`
	if err := tx.QueryRowContext(
		ctx,
		createParseRunStmt,
		result.Evidence.ID,
		input.ParserName,
		input.ParserVersion,
		input.ParserConfigDigest,
		output,
	).Scan(
		&result.ParseRun.ID,
		&result.ParseRun.ReplayEvidenceID,
		&result.ParseRun.ParserName,
		&result.ParseRun.ParserVersion,
		&result.ParseRun.ParserConfigDigest,
		&result.ParseRun.Status,
		&result.ParseRun.OutputJSON,
		&result.ParseRun.CreatedAt,
	); err != nil {
		return hierarchy.ReplayIngestionResult{}, mapSQLError(err)
	}

	if input.ResultSubmissionID != nil {
		const verifySubmissionStmt = `
SELECT context_type, context_id, home_team_id, away_team_id
FROM result_submissions
WHERE id = $1;`
		var submissionContextType string
		var submissionContextID, submissionHomeTeamID, submissionAwayTeamID int64
		if err := tx.QueryRowContext(ctx, verifySubmissionStmt, *input.ResultSubmissionID).Scan(
			&submissionContextType,
			&submissionContextID,
			&submissionHomeTeamID,
			&submissionAwayTeamID,
		); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return hierarchy.ReplayIngestionResult{}, fmt.Errorf("%w: result submission not found", hierarchy.ErrDependency)
			}
			return hierarchy.ReplayIngestionResult{}, err
		}
		if submissionContextType != input.ContextType || submissionContextID != input.ContextID {
			return hierarchy.ReplayIngestionResult{}, fmt.Errorf("%w: result submission context mismatch", hierarchy.ErrConflict)
		}
		if input.SubmittedByTeamID != submissionHomeTeamID && input.SubmittedByTeamID != submissionAwayTeamID {
			return hierarchy.ReplayIngestionResult{}, fmt.Errorf("%w: linkedByTeamId must be submission participant", hierarchy.ErrConflict)
		}

		const linkStmt = `
INSERT INTO result_submission_replay_links(result_submission_id, replay_evidence_id, linked_by_team_id)
VALUES ($1, $2, $3)
ON CONFLICT (result_submission_id, replay_evidence_id) DO NOTHING;`
		if _, err := tx.ExecContext(ctx, linkStmt, *input.ResultSubmissionID, result.Evidence.ID, input.SubmittedByTeamID); err != nil {
			return hierarchy.ReplayIngestionResult{}, mapSQLError(err)
		}
		result.LinkedSubmissionID = input.ResultSubmissionID
	}

	if err := tx.Commit(); err != nil {
		return hierarchy.ReplayIngestionResult{}, err
	}
	return result, nil
}

func (s *HierarchyStore) ListReplayEvidence(ctx context.Context) ([]hierarchy.ReplayEvidence, error) {
	const stmt = `
SELECT id, context_type, context_id, submitted_by_team_id, replay_sha256, content_size_bytes, storage_ref, state, created_at
FROM replay_evidence
ORDER BY id ASC;`
	rows, err := s.db.QueryContext(ctx, stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	evidence := make([]hierarchy.ReplayEvidence, 0)
	for rows.Next() {
		var item hierarchy.ReplayEvidence
		if err := rows.Scan(
			&item.ID,
			&item.ContextType,
			&item.ContextID,
			&item.SubmittedByTeamID,
			&item.ReplaySHA256,
			&item.ContentSizeBytes,
			&item.StorageRef,
			&item.State,
			&item.CreatedAt,
		); err != nil {
			return nil, err
		}
		evidence = append(evidence, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return evidence, nil
}

func (s *HierarchyStore) ListReplayParseRuns(ctx context.Context) ([]hierarchy.ReplayParseRun, error) {
	const stmt = `
SELECT id, replay_evidence_id, parser_name, parser_version, parser_config_digest, status, output_json, created_at
FROM replay_parse_runs
ORDER BY id ASC;`
	rows, err := s.db.QueryContext(ctx, stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	runs := make([]hierarchy.ReplayParseRun, 0)
	for rows.Next() {
		var run hierarchy.ReplayParseRun
		if err := rows.Scan(
			&run.ID,
			&run.ReplayEvidenceID,
			&run.ParserName,
			&run.ParserVersion,
			&run.ParserConfigDigest,
			&run.Status,
			&run.OutputJSON,
			&run.CreatedAt,
		); err != nil {
			return nil, err
		}
		runs = append(runs, run)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return runs, nil
}

func (s *HierarchyStore) ListResultSubmissionReplayLinks(ctx context.Context) ([]hierarchy.ResultSubmissionReplayLink, error) {
	const stmt = `
SELECT id, result_submission_id, replay_evidence_id, linked_by_team_id, created_at
FROM result_submission_replay_links
ORDER BY id ASC;`
	rows, err := s.db.QueryContext(ctx, stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	links := make([]hierarchy.ResultSubmissionReplayLink, 0)
	for rows.Next() {
		var link hierarchy.ResultSubmissionReplayLink
		if err := rows.Scan(
			&link.ID,
			&link.ResultSubmissionID,
			&link.ReplayEvidenceID,
			&link.LinkedByTeamID,
			&link.CreatedAt,
		); err != nil {
			return nil, err
		}
		links = append(links, link)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return links, nil
}

func (s *HierarchyStore) TriggerReplayParse(ctx context.Context, evidenceID, contextID int64, contextType string) error {
	// Find the result submission linked to this evidence, if any.
	const findSubmissionStmt = `
SELECT result_submission_id
FROM result_submission_replay_links
WHERE replay_evidence_id = $1
LIMIT 1;`
	var submissionID int64
	err := s.db.QueryRowContext(ctx, findSubmissionStmt, evidenceID).Scan(&submissionID)
	if err != nil {
		// No submission linked yet — nothing to do.
		return nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	const numRounds = 3
	const playersPerTeam = 3 // standard 3v3

	// Stub round durations (seconds) — deterministic values.
	roundDurations := [numRounds]int32{300, 285, 315}

	for roundNum := int32(1); roundNum <= numRounds; roundNum++ {
		var roundID int64
		const insertRoundStmt = `
INSERT INTO rounds(submission_id, round_number, duration_seconds)
VALUES ($1, $2, $3)
ON CONFLICT (submission_id, round_number) DO NOTHING
RETURNING id;`
		err := tx.QueryRowContext(ctx, insertRoundStmt, submissionID, roundNum, roundDurations[roundNum-1]).Scan(&roundID)
		if err != nil {
			// Round already exists — skip stat lines for this round.
			continue
		}

		// Insert stub player stat lines for 6 players (3 per team).
		// Values are derived deterministically from submissionID, roundNum, and
		// the player's position index so that repeated calls produce consistent
		// data without importing math/rand.
		for pos := int32(0); pos < int32(playersPerTeam*2); pos++ {
			seed := submissionID*100 + int64(roundNum)*10 + int64(pos)
			goals := int32(seed % 3)   // 0–2
			assists := int32(seed % 2) // 0–1
			saves := int32(seed % 4)   // 0–3
			shots := goals + int32(seed%3)
			score := goals*100 + assists*50 + saves*30 + shots*20

			replayIdentity := fmt.Sprintf("stub-player-%d-%d-%d", submissionID, roundNum, pos)

			const insertStatStmt = `
INSERT INTO player_stat_lines(round_id, replay_identity, goals, assists, saves, shots, score)
VALUES ($1, $2, $3, $4, $5, $6, $7);`
			if _, err := tx.ExecContext(ctx, insertStatStmt,
				roundID, replayIdentity, goals, assists, saves, shots, score,
			); err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}
