package db

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/kineticbot/kinetic-v3/internal/domain/hierarchy"
	"github.com/kineticbot/kinetic-v3/internal/platform/replayparser"
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

	replayBytes, err := decodeReplayInput(input)
	if err != nil {
		return hierarchy.ReplayIngestionResult{}, err
	}
	contentHash := sha256.Sum256(replayBytes)
	replaySHA256 := hex.EncodeToString(contentHash[:])
	contentSizeBytes := int64(len(replayBytes))
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
	state,
	replay_body
)
VALUES ($1, $2, $3, $4, $5, $6, 'received', $7)
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
			"base64:"+base64.StdEncoding.EncodeToString(replayBytes),
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
VALUES ($1, $2, $3, $4, 'queued', '{}')
RETURNING id, replay_evidence_id, parser_name, parser_version, parser_config_digest, status, output_json, created_at;`
	if err := tx.QueryRowContext(
		ctx,
		createParseRunStmt,
		result.Evidence.ID,
		input.ParserName,
		input.ParserVersion,
		input.ParserConfigDigest,
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
SELECT context_type, context_id, home_team_id, away_team_id, payload_json, provenance_json
FROM result_submissions
WHERE id = $1;`
		var submissionContextType string
		var submissionContextID, submissionHomeTeamID, submissionAwayTeamID int64
		var submissionPayload, submissionProvenance json.RawMessage
		if err := tx.QueryRowContext(ctx, verifySubmissionStmt, *input.ResultSubmissionID).Scan(
			&submissionContextType,
			&submissionContextID,
			&submissionHomeTeamID,
			&submissionAwayTeamID,
			&submissionPayload,
			&submissionProvenance,
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

		autofillJSON, conflictJSON, provenanceJSON, err := buildReplayReviewMetadata(nil, submissionPayload, submissionProvenance)
		if err != nil {
			return hierarchy.ReplayIngestionResult{}, err
		}
		result.AutofillJSON = autofillJSON
		result.ConflictJSON = conflictJSON
		const updateReviewStmt = `
UPDATE result_submissions
SET provenance_json = $2
WHERE id = $1;`
		if _, err := tx.ExecContext(ctx, updateReviewStmt, *input.ResultSubmissionID, provenanceJSON); err != nil {
			return hierarchy.ReplayIngestionResult{}, mapSQLError(err)
		}
	}

	if err := tx.Commit(); err != nil {
		return hierarchy.ReplayIngestionResult{}, err
	}
	return result, nil
}

func decodeReplayInput(input hierarchy.IngestReplayEvidenceInput) ([]byte, error) {
	if strings.TrimSpace(input.ReplayBodyBase64) != "" {
		decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(input.ReplayBodyBase64))
		if err != nil {
			return nil, fmt.Errorf("%w: replayBodyBase64 must be valid base64", hierarchy.ErrInvalidInput)
		}
		if len(decoded) == 0 {
			return nil, fmt.Errorf("%w: replayBodyBase64 must not be empty", hierarchy.ErrInvalidInput)
		}
		return decoded, nil
	}
	return []byte(input.ReplayBody), nil
}

func buildReplayReviewMetadata(parserOutput, submissionPayload, existingProvenance json.RawMessage) (json.RawMessage, json.RawMessage, json.RawMessage, error) {
	autofillJSON := normalizeJSON(parserOutput)
	parserScore, parserScoreOK, err := hierarchy.ExtractRocketLeagueScore(autofillJSON)
	if err != nil {
		return nil, nil, nil, err
	}
	submissionScore, submissionScoreOK, err := hierarchy.ExtractRocketLeagueScore(submissionPayload)
	if err != nil {
		return nil, nil, nil, err
	}

	review := map[string]any{
		"status": "parser_unavailable",
		"source": "parser",
	}
	var conflictJSON json.RawMessage
	if parserScoreOK {
		review["parserScore"] = map[string]any{"home": parserScore.Home, "away": parserScore.Away}
		review["status"] = "autofill_available"
	}
	if parserScoreOK && submissionScoreOK {
		review["manualScore"] = map[string]any{"home": submissionScore.Home, "away": submissionScore.Away}
		if parserScore.Home == submissionScore.Home && parserScore.Away == submissionScore.Away {
			review["status"] = "matched"
		} else {
			review["status"] = "conflict"
			conflictJSON = mustJSON(review)
		}
	}

	provenance := map[string]any{}
	if len(existingProvenance) > 0 {
		_ = json.Unmarshal(existingProvenance, &provenance)
	}
	provenance["replayReview"] = review
	return autofillJSON, conflictJSON, mustJSON(provenance), nil
}

func mustJSON(value any) json.RawMessage {
	raw, err := json.Marshal(value)
	if err != nil {
		return json.RawMessage(`{}`)
	}
	return raw
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
SELECT id, replay_evidence_id, parser_name, parser_version, parser_config_digest, status, output_json, failure_reason, started_at, finished_at, created_at
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
			&run.FailureReason,
			&run.StartedAt,
			&run.FinishedAt,
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
	if s.replayParser == nil {
		return fmt.Errorf("replay parser is not configured")
	}

	const queuedRunStmt = `
SELECT pr.id, re.replay_body
FROM replay_parse_runs pr
JOIN replay_evidence re ON re.id = pr.replay_evidence_id
WHERE pr.replay_evidence_id = $1 AND pr.status = 'queued'
ORDER BY pr.created_at ASC
LIMIT 1;`
	var parseRunID int64
	var encodedReplay string
	if err := s.db.QueryRowContext(ctx, queuedRunStmt, evidenceID).Scan(&parseRunID, &encodedReplay); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}

	result, err := s.db.ExecContext(ctx, `
UPDATE replay_parse_runs
SET status = 'running', started_at = NOW(), failure_reason = NULL
WHERE id = $1 AND status = 'queued';`, parseRunID)
	if err != nil {
		return err
	}
	claimed, err := result.RowsAffected()
	if err != nil || claimed == 0 {
		return err
	}

	replayBytes := []byte(encodedReplay)
	if encoded, ok := strings.CutPrefix(encodedReplay, "base64:"); ok {
		replayBytes, err = base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			return s.failReplayParse(ctx, parseRunID, evidenceID, fmt.Errorf("decode stored replay: %w", err))
		}
	}
	parsed, err := s.replayParser.Parse(ctx, replayBytes)
	if err != nil {
		return s.failReplayParse(ctx, parseRunID, evidenceID, err)
	}

	const submissionStmt = `
SELECT rs.id, rs.home_team_id, rs.away_team_id, rs.payload_json, rs.provenance_json
FROM result_submission_replay_links link
JOIN result_submissions rs ON rs.id = link.result_submission_id
WHERE link.replay_evidence_id = $1
LIMIT 1;`
	var submissionID, homeTeamID, awayTeamID int64
	var submissionPayload, submissionProvenance json.RawMessage
	err = s.db.QueryRowContext(ctx, submissionStmt, evidenceID).Scan(
		&submissionID,
		&homeTeamID,
		&awayTeamID,
		&submissionPayload,
		&submissionProvenance,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return s.finishUnlinkedReplayParse(ctx, parseRunID, evidenceID, parsed.RawJSON)
	}
	if err != nil {
		return s.failReplayParse(ctx, parseRunID, evidenceID, err)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return s.failReplayParse(ctx, parseRunID, evidenceID, err)
	}
	defer func() { _ = tx.Rollback() }()
	failTx := func(parseErr error) error {
		_ = tx.Rollback()
		return s.failReplayParse(ctx, parseRunID, evidenceID, parseErr)
	}
	var lockedSubmissionID int64
	if err := tx.QueryRowContext(ctx, `SELECT id FROM result_submissions WHERE id = $1 FOR UPDATE;`, submissionID).Scan(&lockedSubmissionID); err != nil {
		return failTx(err)
	}

	roundNumber := int32(1)
	if err := tx.QueryRowContext(ctx, `
SELECT COALESCE(MAX(round_number), 0) + 1
FROM rounds
WHERE submission_id = $1 AND replay_evidence_id IS DISTINCT FROM $2;`, submissionID, evidenceID).Scan(&roundNumber); err != nil {
		return failTx(err)
	}
	duration := int32(0)
	if parsed.Summary.DurationSeconds != nil && *parsed.Summary.DurationSeconds > 0 {
		duration = int32(*parsed.Summary.DurationSeconds)
	}

	var roundID int64
	if err := tx.QueryRowContext(ctx, `
INSERT INTO rounds(submission_id, replay_evidence_id, round_number, duration_seconds)
VALUES ($1, $2, $3, $4)
ON CONFLICT (replay_evidence_id) DO UPDATE
SET duration_seconds = EXCLUDED.duration_seconds
RETURNING id;`, submissionID, evidenceID, roundNumber, duration).Scan(&roundID); err != nil {
		return failTx(err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM player_stat_lines WHERE round_id = $1;`, roundID); err != nil {
		return failTx(err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM team_stat_lines WHERE round_id = $1;`, roundID); err != nil {
		return failTx(err)
	}

	unresolved := false
	type teamTotals struct{ goals, shots, saves int32 }
	totals := map[int32]*teamTotals{0: {}, 1: {}}
	rosterTeams := map[int32]map[int64]int{0: {}, 1: {}}
	for _, player := range parsed.Summary.Players {
		provider, accountID := replayIdentity(player.Identity)
		var playerID *int64
		if provider != "" && accountID != "" {
			var resolvedID int64
			err := tx.QueryRowContext(ctx, `
SELECT player_id
FROM platform_account_links
WHERE provider = $1 AND provider_account_id = $2 AND is_active = true AND player_id IS NOT NULL
LIMIT 1;`, provider, accountID).Scan(&resolvedID)
			switch {
			case err == nil:
				playerID = &resolvedID
				var rosterTeamID int64
				err = tx.QueryRowContext(ctx, `
SELECT team_id
FROM roster_memberships
WHERE player_id = $1 AND is_active = true
LIMIT 1;`, resolvedID).Scan(&rosterTeamID)
				switch {
				case err == nil && (rosterTeamID == homeTeamID || rosterTeamID == awayTeamID):
					rosterTeams[*player.Team][rosterTeamID]++
				case errors.Is(err, sql.ErrNoRows):
					unresolved = true
				case err != nil:
					return failTx(err)
				default:
					unresolved = true
				}
			case errors.Is(err, sql.ErrNoRows):
				unresolved = true
			default:
				return failTx(err)
			}
		} else {
			unresolved = true
		}

		identity := strings.TrimSpace(player.Identity.Name)
		if provider != "" && accountID != "" {
			identity = provider + ":" + accountID
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO player_stat_lines(round_id, player_id, replay_identity, goals, assists, saves, shots, score)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);`,
			roundID, playerID, identity, player.Stats.Goals, player.Stats.Assists,
			player.Stats.Saves, player.Stats.Shots, player.Stats.Score,
		); err != nil {
			return failTx(err)
		}
		team := *player.Team
		totals[team].shots += player.Stats.Shots
		totals[team].saves += player.Stats.Saves
	}

	scores := replayTeamScores(parsed.Summary.Teams)
	totals[0].goals = scores[0]
	totals[1].goals = scores[1]
	teamIDs, teamContextResolved := resolveReplayTeamIDs(rosterTeams, homeTeamID, awayTeamID)
	if !teamContextResolved {
		unresolved = true
	}
	for _, teamIndex := range []int32{0, 1} {
		team := totals[teamIndex]
		if _, err := tx.ExecContext(ctx, `
INSERT INTO team_stat_lines(round_id, team_id, goals, shots, saves)
VALUES ($1, $2, $3, $4, $5);`, roundID, teamIDs[teamIndex], team.goals, team.shots, team.saves); err != nil {
			return failTx(err)
		}
	}

	var conflictJSON, provenanceJSON json.RawMessage
	if teamContextResolved {
		homeScore, awayScore := scores[0], scores[1]
		if *teamIDs[0] == awayTeamID {
			homeScore, awayScore = scores[1], scores[0]
		}
		parserScorePayload := mustJSON(map[string]any{
			"score": map[string]int32{"home": homeScore, "away": awayScore},
		})
		_, conflictJSON, provenanceJSON, err = buildReplayReviewMetadata(parserScorePayload, submissionPayload, submissionProvenance)
		if err != nil {
			return failTx(err)
		}
	} else {
		provenance := map[string]any{}
		_ = json.Unmarshal(submissionProvenance, &provenance)
		provenance["replayReview"] = map[string]any{
			"status": "context_unresolved",
			"source": "parser",
		}
		provenanceJSON = mustJSON(provenance)
	}
	status := "parsed"
	evidenceState := "parsed"
	if unresolved || len(conflictJSON) > 0 {
		status = "needs_review"
		evidenceState = "needs_review"
	}
	if _, err := tx.ExecContext(ctx, `
UPDATE replay_parse_runs
SET status = $2, output_json = $3, finished_at = NOW(), failure_reason = NULL
WHERE id = $1;`, parseRunID, status, parsed.RawJSON); err != nil {
		return failTx(err)
	}
	if _, err := tx.ExecContext(ctx, `UPDATE replay_evidence SET state = $2 WHERE id = $1;`, evidenceID, evidenceState); err != nil {
		return failTx(err)
	}
	if _, err := tx.ExecContext(ctx, `UPDATE result_submissions SET provenance_json = $2 WHERE id = $1;`, submissionID, provenanceJSON); err != nil {
		return failTx(err)
	}
	if err := tx.Commit(); err != nil {
		return s.failReplayParse(ctx, parseRunID, evidenceID, err)
	}
	return nil
}

func (s *HierarchyStore) failReplayParse(ctx context.Context, parseRunID, evidenceID int64, parseErr error) error {
	cleanupCtx := ctx
	if ctx.Err() != nil {
		var cancel context.CancelFunc
		cleanupCtx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
	}
	_, _ = s.db.ExecContext(cleanupCtx, `
UPDATE replay_parse_runs
SET status = 'failed', failure_reason = $2, finished_at = NOW()
WHERE id = $1;`, parseRunID, parseErr.Error())
	_, _ = s.db.ExecContext(cleanupCtx, `UPDATE replay_evidence SET state = 'needs_review' WHERE id = $1;`, evidenceID)
	return parseErr
}

func (s *HierarchyStore) finishUnlinkedReplayParse(ctx context.Context, parseRunID, evidenceID int64, output json.RawMessage) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `
UPDATE replay_parse_runs SET status = 'parsed', output_json = $2, finished_at = NOW() WHERE id = $1;`, parseRunID, output); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE replay_evidence SET state = 'parsed' WHERE id = $1;`, evidenceID); err != nil {
		return err
	}
	return tx.Commit()
}

func replayTeamScores(teams []replayparser.Team) map[int32]int32 {
	scores := map[int32]int32{0: 0, 1: 0}
	for _, team := range teams {
		if team.Index == 0 || team.Index == 1 {
			scores[team.Index] = team.Score
		}
	}
	return scores
}

func resolveReplayTeamIDs(rosterTeams map[int32]map[int64]int, homeTeamID, awayTeamID int64) (map[int32]*int64, bool) {
	resolved := map[int32]*int64{0: nil, 1: nil}
	for _, parserTeam := range []int32{0, 1} {
		memberships := rosterTeams[parserTeam]
		if len(memberships) != 1 {
			return resolved, false
		}
		for teamID := range memberships {
			if teamID != homeTeamID && teamID != awayTeamID {
				return resolved, false
			}
			id := teamID
			resolved[parserTeam] = &id
		}
	}
	if *resolved[0] == *resolved[1] {
		return map[int32]*int64{0: nil, 1: nil}, false
	}
	return resolved, true
}

func replayIdentity(identity replayparser.PlayerIdentity) (string, string) {
	if identity.EpicID != nil && strings.TrimSpace(*identity.EpicID) != "" {
		return "epic", strings.TrimSpace(*identity.EpicID)
	}
	if identity.Platform == nil || identity.OnlineID == nil {
		return "", ""
	}
	platform := strings.ToLower(strings.TrimSpace(*identity.Platform))
	provider := ""
	switch {
	case strings.Contains(platform, "steam"):
		provider = "steam"
	case strings.Contains(platform, "xbox"), strings.Contains(platform, "dingo"):
		provider = "xbox"
	case strings.Contains(platform, "psn"), strings.Contains(platform, "playstation"), strings.Contains(platform, "ps4"):
		provider = "psn"
	case strings.Contains(platform, "epic"):
		provider = "epic"
	}
	accountID := strings.TrimSpace(*identity.OnlineID)
	if provider == "epic" && accountID == "0" {
		accountID = ""
	}
	return provider, accountID
}
