package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/kineticbot/kinetic-v3/internal/domain/hierarchy"
)

func (s *HierarchyStore) ReportException(ctx context.Context, input hierarchy.ReportExceptionInput) (hierarchy.ExceptionTicket, error) {
	if err := hierarchy.ValidateReportExceptionInput(input); err != nil {
		return hierarchy.ExceptionTicket{}, err
	}
	if err := ensureExceptionContextExists(ctx, s.db, input.ContextType, input.ContextID); err != nil {
		return hierarchy.ExceptionTicket{}, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return hierarchy.ExceptionTicket{}, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	ticket, err := insertExceptionTicket(ctx, tx, input)
	if err != nil {
		return hierarchy.ExceptionTicket{}, err
	}

	if _, err := tx.ExecContext(
		ctx,
		`INSERT INTO exception_actions(ticket_id, action_type, actor, automated, notes, minutes_spent) VALUES ($1, 'reported', $2, FALSE, $3, 0)`,
		ticket.ID,
		"system",
		ticket.ReasonCode,
	); err != nil {
		return hierarchy.ExceptionTicket{}, err
	}

	if err := tx.Commit(); err != nil {
		return hierarchy.ExceptionTicket{}, err
	}
	return ticket, nil
}

func (s *HierarchyStore) ListOperatorInbox(ctx context.Context) ([]hierarchy.ExceptionTicket, error) {
	const stmt = `
SELECT id, category, context_type, context_id, reported_by_team_id, state, reason_code, severity, suggested_action, details_json, resolution_code, opened_at, triaged_at, resolved_at
FROM exception_tickets
WHERE state <> 'resolved'
ORDER BY severity DESC, opened_at ASC, id ASC;`
	rows, err := s.db.QueryContext(ctx, stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tickets := make([]hierarchy.ExceptionTicket, 0)
	for rows.Next() {
		var ticket hierarchy.ExceptionTicket
		if err := rows.Scan(
			&ticket.ID,
			&ticket.Category,
			&ticket.ContextType,
			&ticket.ContextID,
			&ticket.ReportedByTeamID,
			&ticket.State,
			&ticket.ReasonCode,
			&ticket.Severity,
			&ticket.SuggestedAction,
			&ticket.DetailsJSON,
			&ticket.ResolutionCode,
			&ticket.OpenedAt,
			&ticket.TriagedAt,
			&ticket.ResolvedAt,
		); err != nil {
			return nil, err
		}
		tickets = append(tickets, ticket)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tickets, nil
}

func (s *HierarchyStore) TriageException(ctx context.Context, input hierarchy.TriageExceptionInput) (hierarchy.ExceptionTicket, error) {
	if err := hierarchy.ValidateTriageExceptionInput(input); err != nil {
		return hierarchy.ExceptionTicket{}, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return hierarchy.ExceptionTicket{}, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	const stmt = `
UPDATE exception_tickets
SET
	state = 'triaged',
	reason_code = $2,
	severity = $3,
	suggested_action = $4,
	triaged_at = NOW()
WHERE id = $1
  AND state <> 'resolved'
RETURNING id, category, context_type, context_id, reported_by_team_id, state, reason_code, severity, suggested_action, details_json, resolution_code, opened_at, triaged_at, resolved_at;`
	var ticket hierarchy.ExceptionTicket
	err = tx.QueryRowContext(ctx, stmt, input.TicketID, input.ReasonCode, input.Severity, input.SuggestedAction).Scan(
		&ticket.ID,
		&ticket.Category,
		&ticket.ContextType,
		&ticket.ContextID,
		&ticket.ReportedByTeamID,
		&ticket.State,
		&ticket.ReasonCode,
		&ticket.Severity,
		&ticket.SuggestedAction,
		&ticket.DetailsJSON,
		&ticket.ResolutionCode,
		&ticket.OpenedAt,
		&ticket.TriagedAt,
		&ticket.ResolvedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return hierarchy.ExceptionTicket{}, fmt.Errorf("%w: triage not allowed or ticket not found", hierarchy.ErrConflict)
		}
		return hierarchy.ExceptionTicket{}, mapSQLError(err)
	}

	if _, err := tx.ExecContext(
		ctx,
		`INSERT INTO exception_actions(ticket_id, action_type, actor, automated, notes, minutes_spent) VALUES ($1, 'triaged', $2, FALSE, $3, $4)`,
		ticket.ID,
		input.Actor,
		input.ReasonCode,
		input.MinutesSpent,
	); err != nil {
		return hierarchy.ExceptionTicket{}, err
	}

	if err := tx.Commit(); err != nil {
		return hierarchy.ExceptionTicket{}, err
	}
	return ticket, nil
}

func (s *HierarchyStore) ResolveException(ctx context.Context, input hierarchy.ResolveExceptionInput) (hierarchy.ExceptionTicket, error) {
	if err := hierarchy.ValidateResolveExceptionInput(input); err != nil {
		return hierarchy.ExceptionTicket{}, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return hierarchy.ExceptionTicket{}, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	const stmt = `
UPDATE exception_tickets
SET
	state = 'resolved',
	resolution_code = $2,
	resolved_at = NOW()
WHERE id = $1
  AND state <> 'resolved'
RETURNING id, category, context_type, context_id, reported_by_team_id, state, reason_code, severity, suggested_action, details_json, resolution_code, opened_at, triaged_at, resolved_at;`
	var ticket hierarchy.ExceptionTicket
	err = tx.QueryRowContext(ctx, stmt, input.TicketID, input.ResolutionCode).Scan(
		&ticket.ID,
		&ticket.Category,
		&ticket.ContextType,
		&ticket.ContextID,
		&ticket.ReportedByTeamID,
		&ticket.State,
		&ticket.ReasonCode,
		&ticket.Severity,
		&ticket.SuggestedAction,
		&ticket.DetailsJSON,
		&ticket.ResolutionCode,
		&ticket.OpenedAt,
		&ticket.TriagedAt,
		&ticket.ResolvedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return hierarchy.ExceptionTicket{}, fmt.Errorf("%w: resolve not allowed or ticket not found", hierarchy.ErrConflict)
		}
		return hierarchy.ExceptionTicket{}, mapSQLError(err)
	}

	actionType := "resolved"
	if input.Automated {
		actionType = "auto_resolved"
	}
	if _, err := tx.ExecContext(
		ctx,
		`INSERT INTO exception_actions(ticket_id, action_type, actor, automated, notes, minutes_spent) VALUES ($1, $2, $3, $4, $5, $6)`,
		ticket.ID,
		actionType,
		input.Actor,
		input.Automated,
		input.Notes,
		input.MinutesSpent,
	); err != nil {
		return hierarchy.ExceptionTicket{}, err
	}

	if err := tx.Commit(); err != nil {
		return hierarchy.ExceptionTicket{}, err
	}
	return ticket, nil
}

func (s *HierarchyStore) ListExceptionActions(ctx context.Context) ([]hierarchy.ExceptionAction, error) {
	const stmt = `
SELECT id, ticket_id, action_type, actor, automated, notes, minutes_spent, created_at
FROM exception_actions
ORDER BY id ASC;`
	rows, err := s.db.QueryContext(ctx, stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	actions := make([]hierarchy.ExceptionAction, 0)
	for rows.Next() {
		var action hierarchy.ExceptionAction
		if err := rows.Scan(
			&action.ID,
			&action.TicketID,
			&action.ActionType,
			&action.Actor,
			&action.Automated,
			&action.Notes,
			&action.MinutesSpent,
			&action.CreatedAt,
		); err != nil {
			return nil, err
		}
		actions = append(actions, action)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return actions, nil
}

func (s *HierarchyStore) GetExceptionMetrics(ctx context.Context) (hierarchy.ExceptionMetrics, error) {
	metrics := hierarchy.ExceptionMetrics{}

	const adminHoursStmt = `
SELECT COALESCE(SUM(minutes_spent), 0)::FLOAT8 / 60.0
FROM exception_actions
WHERE automated = FALSE
  AND created_at >= NOW() - INTERVAL '7 days';`
	if err := s.db.QueryRowContext(ctx, adminHoursStmt).Scan(&metrics.AdminHoursPerWeek); err != nil {
		return hierarchy.ExceptionMetrics{}, err
	}

	const touchesStmt = `
WITH total_matches AS (
	SELECT COUNT(*)::FLOAT8 AS c FROM matches
), manual_touches AS (
	SELECT COUNT(a.id)::FLOAT8 AS c
	FROM exception_actions a
	JOIN exception_tickets t ON t.id = a.ticket_id
	WHERE a.automated = FALSE
	  AND t.context_type = 'match'
)
SELECT
	CASE
		WHEN tm.c = 0 THEN 0.0
		ELSE mt.c / tm.c
	END
FROM total_matches tm, manual_touches mt;`
	if err := s.db.QueryRowContext(ctx, touchesStmt).Scan(&metrics.ManualTouchesPerFixture); err != nil {
		return hierarchy.ExceptionMetrics{}, err
	}

	const zeroTouchStmt = `
WITH total_matches AS (
	SELECT COUNT(*)::FLOAT8 AS c FROM matches
), touched_matches AS (
	SELECT COUNT(DISTINCT t.context_id)::FLOAT8 AS c
	FROM exception_actions a
	JOIN exception_tickets t ON t.id = a.ticket_id
	WHERE a.automated = FALSE
	  AND t.context_type = 'match'
)
SELECT
	CASE
		WHEN tm.c = 0 THEN 1.0
		ELSE GREATEST((tm.c - touched.c) / tm.c, 0.0)
	END
FROM total_matches tm, touched_matches touched;`
	if err := s.db.QueryRowContext(ctx, zeroTouchStmt).Scan(&metrics.ZeroTouchFixtureRate); err != nil {
		return hierarchy.ExceptionMetrics{}, err
	}

	const p50Stmt = `
SELECT COALESCE(
	EXTRACT(EPOCH FROM percentile_cont(0.5) WITHIN GROUP (ORDER BY (resolved_at - opened_at))) / 3600.0,
	0.0
)
FROM exception_tickets
WHERE state = 'resolved'
  AND resolved_at IS NOT NULL;`
	if err := s.db.QueryRowContext(ctx, p50Stmt).Scan(&metrics.TimeToCloseHoursP50); err != nil {
		return hierarchy.ExceptionMetrics{}, err
	}

	return metrics, nil
}

func (s *HierarchyStore) EvaluateSchedulingException(ctx context.Context, input hierarchy.EvaluateSchedulingExceptionInput) (hierarchy.ExceptionAutomationResult, error) {
	if err := hierarchy.ValidateEvaluateSchedulingExceptionInput(input); err != nil {
		return hierarchy.ExceptionAutomationResult{}, err
	}

	reportedByTeamID := (*int64)(nil)
	reportInput := hierarchy.ReportExceptionInput{
		Category:         "scheduling_conflict",
		ContextType:      "match",
		ContextID:        input.MatchID,
		ReportedByTeamID: reportedByTeamID,
		ReasonCode:       input.ConflictCode,
		Severity:         3,
		SuggestedAction:  "request_reschedule_confirmation",
		DetailsJSON:      []byte(`{}`),
	}
	ticket, err := s.ReportException(ctx, reportInput)
	if err != nil {
		return hierarchy.ExceptionAutomationResult{}, err
	}

	result := hierarchy.ExceptionAutomationResult{Ticket: ticket}
	if input.HomeConfirmed && input.AwayConfirmed {
		resolved, err := s.ResolveException(ctx, hierarchy.ResolveExceptionInput{
			TicketID:       ticket.ID,
			Actor:          input.Actor,
			ResolutionCode: "reschedule_confirmed",
			Notes:          "both teams confirmed schedule",
			Automated:      true,
			MinutesSpent:   0,
		})
		if err != nil {
			return hierarchy.ExceptionAutomationResult{}, err
		}
		result.Ticket = resolved
		result.AutoResolved = true
	}
	return result, nil
}

func (s *HierarchyStore) EvaluateNoShowException(ctx context.Context, input hierarchy.EvaluateNoShowExceptionInput) (hierarchy.ExceptionAutomationResult, error) {
	if err := hierarchy.ValidateEvaluateNoShowExceptionInput(input); err != nil {
		return hierarchy.ExceptionAutomationResult{}, err
	}

	reasonCode := "attendance_uncertain"
	suggestedAction := "manual_review"
	severity := int32(3)
	autoResolve := false
	resolutionCode := ""
	if input.GraceMinutes >= 15 {
		if input.HomeCheckedIn && !input.AwayCheckedIn {
			reasonCode = "away_no_show"
			suggestedAction = "forfeit_away_team"
			severity = 5
			autoResolve = true
			resolutionCode = "forfeit_away"
		} else if !input.HomeCheckedIn && input.AwayCheckedIn {
			reasonCode = "home_no_show"
			suggestedAction = "forfeit_home_team"
			severity = 5
			autoResolve = true
			resolutionCode = "forfeit_home"
		}
	}

	ticket, err := s.ReportException(ctx, hierarchy.ReportExceptionInput{
		Category:        "no_show",
		ContextType:     "match",
		ContextID:       input.MatchID,
		ReasonCode:      reasonCode,
		Severity:        severity,
		SuggestedAction: suggestedAction,
		DetailsJSON:     []byte(`{}`),
	})
	if err != nil {
		return hierarchy.ExceptionAutomationResult{}, err
	}

	result := hierarchy.ExceptionAutomationResult{Ticket: ticket}
	if autoResolve {
		resolved, err := s.ResolveException(ctx, hierarchy.ResolveExceptionInput{
			TicketID:       ticket.ID,
			Actor:          input.Actor,
			ResolutionCode: resolutionCode,
			Notes:          "rule-based no-show forfeit recommendation applied",
			Automated:      true,
			MinutesSpent:   0,
		})
		if err != nil {
			return hierarchy.ExceptionAutomationResult{}, err
		}
		result.Ticket = resolved
		result.AutoResolved = true
	}
	return result, nil
}

func (s *HierarchyStore) EvaluateReplayDisputeException(ctx context.Context, input hierarchy.EvaluateReplayDisputeExceptionInput) (hierarchy.ExceptionAutomationResult, error) {
	if err := hierarchy.ValidateEvaluateReplayDisputeExceptionInput(input); err != nil {
		return hierarchy.ExceptionAutomationResult{}, err
	}

	category := "result_dispute"
	reasonCode := "dispute_reported"
	suggestedAction := "manual_review"
	severity := int32(4)
	autoResolve := false
	resolutionCode := ""

	if input.ParseStatus == "failed" {
		category = "replay_parse_failure"
		reasonCode = "parse_failed"
		suggestedAction = "reupload_replay"
		severity = 5
	} else if input.IdentityStatus == "mismatch" {
		category = "replay_identity_mismatch"
		reasonCode = "identity_mismatch"
		suggestedAction = "verify_platform_links"
		severity = 5
	} else if !input.DisputeRaised {
		reasonCode = "validated_no_dispute"
		suggestedAction = "no_action_required"
		severity = 1
		autoResolve = true
		resolutionCode = "no_dispute_detected"
	}

	ticket, err := s.ReportException(ctx, hierarchy.ReportExceptionInput{
		Category:        category,
		ContextType:     "result_submission",
		ContextID:       input.ResultSubmissionID,
		ReasonCode:      reasonCode,
		Severity:        severity,
		SuggestedAction: suggestedAction,
		DetailsJSON:     []byte(`{}`),
	})
	if err != nil {
		return hierarchy.ExceptionAutomationResult{}, err
	}

	result := hierarchy.ExceptionAutomationResult{Ticket: ticket}
	if autoResolve {
		resolved, err := s.ResolveException(ctx, hierarchy.ResolveExceptionInput{
			TicketID:       ticket.ID,
			Actor:          input.Actor,
			ResolutionCode: resolutionCode,
			Notes:          "replay and identity checks passed without dispute",
			Automated:      true,
			MinutesSpent:   0,
		})
		if err != nil {
			return hierarchy.ExceptionAutomationResult{}, err
		}
		result.Ticket = resolved
		result.AutoResolved = true
	}
	return result, nil
}

func ensureExceptionContextExists(ctx context.Context, queryer interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}, contextType string, contextID int64) error {
	var stmt string
	switch contextType {
	case "match":
		stmt = `SELECT id FROM matches WHERE id = $1;`
	case "scrim":
		stmt = `SELECT id FROM scrims WHERE id = $1;`
	case "result_submission":
		stmt = `SELECT id FROM result_submissions WHERE id = $1;`
	case "replay_evidence":
		stmt = `SELECT id FROM replay_evidence WHERE id = $1;`
	default:
		return fmt.Errorf("%w: unsupported context type", hierarchy.ErrInvalidInput)
	}

	var id int64
	if err := queryer.QueryRowContext(ctx, stmt, contextID).Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%w: exception context not found", hierarchy.ErrDependency)
		}
		return err
	}
	return nil
}

func insertExceptionTicket(ctx context.Context, tx *sql.Tx, input hierarchy.ReportExceptionInput) (hierarchy.ExceptionTicket, error) {
	details := input.DetailsJSON
	if len(details) == 0 {
		details = []byte(`{}`)
	}

	const stmt = `
INSERT INTO exception_tickets(
	category,
	context_type,
	context_id,
	reported_by_team_id,
	reason_code,
	severity,
	suggested_action,
	details_json
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id, category, context_type, context_id, reported_by_team_id, state, reason_code, severity, suggested_action, details_json, resolution_code, opened_at, triaged_at, resolved_at;`
	var ticket hierarchy.ExceptionTicket
	err := tx.QueryRowContext(
		ctx,
		stmt,
		input.Category,
		input.ContextType,
		input.ContextID,
		input.ReportedByTeamID,
		input.ReasonCode,
		input.Severity,
		input.SuggestedAction,
		details,
	).Scan(
		&ticket.ID,
		&ticket.Category,
		&ticket.ContextType,
		&ticket.ContextID,
		&ticket.ReportedByTeamID,
		&ticket.State,
		&ticket.ReasonCode,
		&ticket.Severity,
		&ticket.SuggestedAction,
		&ticket.DetailsJSON,
		&ticket.ResolutionCode,
		&ticket.OpenedAt,
		&ticket.TriagedAt,
		&ticket.ResolvedAt,
	)
	if err != nil {
		return hierarchy.ExceptionTicket{}, mapSQLError(err)
	}
	return ticket, nil
}
