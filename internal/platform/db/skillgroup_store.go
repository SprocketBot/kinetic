package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/kineticbot/kinetic-v3/internal/domain/skillgroup"
)

type SkillGroupStore struct {
	db *sql.DB
}

func NewSkillGroupStore(db *sql.DB) *SkillGroupStore {
	return &SkillGroupStore{db: db}
}

func mapSkillGroupSQLError(err error) error {
	var pgError *pgconn.PgError
	if errors.As(err, &pgError) {
		switch pgError.Code {
		case "23505":
			return fmt.Errorf("%w: %s", skillgroup.ErrConflict, pgError.ConstraintName)
		case "23503":
			return fmt.Errorf("%w: %s", skillgroup.ErrDependency, pgError.ConstraintName)
		case "23514":
			return fmt.Errorf("%w: %s", skillgroup.ErrInvalidInput, pgError.ConstraintName)
		}
	}
	return err
}

func scanSkillGroup(scanner interface {
	Scan(dest ...any) error
}) (skillgroup.SkillGroup, error) {
	var sg skillgroup.SkillGroup
	var promotionThreshold sql.NullInt32
	var demotionThreshold sql.NullInt32
	if err := scanner.Scan(
		&sg.ID,
		&sg.LeagueID,
		&sg.Name,
		&sg.DisplayOrder,
		&sg.RatingFloor,
		&sg.RatingCeiling,
		&promotionThreshold,
		&demotionThreshold,
		&sg.IsActive,
		&sg.CreatedAt,
	); err != nil {
		return skillgroup.SkillGroup{}, err
	}
	if promotionThreshold.Valid {
		v := promotionThreshold.Int32
		sg.PromotionThreshold = &v
	}
	if demotionThreshold.Valid {
		v := demotionThreshold.Int32
		sg.DemotionThreshold = &v
	}
	return sg, nil
}

func (s *SkillGroupStore) CreateSkillGroup(ctx context.Context, input skillgroup.CreateSkillGroupInput) (skillgroup.SkillGroup, error) {
	const stmt = `
INSERT INTO skill_groups(league_id, name, display_order, rating_floor, rating_ceiling, promotion_threshold, demotion_threshold)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, league_id, name, display_order, rating_floor, rating_ceiling, promotion_threshold, demotion_threshold, is_active, created_at;`

	row := s.db.QueryRowContext(ctx, stmt,
		input.LeagueID,
		input.Name,
		input.DisplayOrder,
		input.RatingFloor,
		input.RatingCeiling,
		input.PromotionThreshold,
		input.DemotionThreshold,
	)
	sg, err := scanSkillGroup(row)
	if err != nil {
		return skillgroup.SkillGroup{}, mapSkillGroupSQLError(err)
	}
	return sg, nil
}

func (s *SkillGroupStore) ListSkillGroups(ctx context.Context, leagueID int64) ([]skillgroup.SkillGroup, error) {
	const stmt = `
SELECT id, league_id, name, display_order, rating_floor, rating_ceiling,
       promotion_threshold, demotion_threshold, is_active, created_at
FROM skill_groups
WHERE league_id = $1
ORDER BY display_order ASC;`

	rows, err := s.db.QueryContext(ctx, stmt, leagueID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	groups := make([]skillgroup.SkillGroup, 0)
	for rows.Next() {
		sg, err := scanSkillGroup(rows)
		if err != nil {
			return nil, err
		}
		groups = append(groups, sg)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return groups, nil
}

func (s *SkillGroupStore) GetSkillGroup(ctx context.Context, skillGroupID int64) (skillgroup.SkillGroup, error) {
	const stmt = `
SELECT id, league_id, name, display_order, rating_floor, rating_ceiling,
       promotion_threshold, demotion_threshold, is_active, created_at
FROM skill_groups
WHERE id = $1;`

	row := s.db.QueryRowContext(ctx, stmt, skillGroupID)
	sg, err := scanSkillGroup(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return skillgroup.SkillGroup{}, fmt.Errorf("%w: skill group not found", skillgroup.ErrNotFound)
		}
		return skillgroup.SkillGroup{}, mapSkillGroupSQLError(err)
	}
	return sg, nil
}

func (s *SkillGroupStore) UpdateSkillGroup(ctx context.Context, input skillgroup.UpdateSkillGroupInput) (skillgroup.SkillGroup, error) {
	setClauses := make([]string, 0)
	args := make([]any, 0)

	// First arg slot is reserved for the WHERE id = $N at the end; build SET args starting at $1.
	argIdx := 1

	if input.RatingFloor != nil {
		setClauses = append(setClauses, fmt.Sprintf("rating_floor = $%d", argIdx))
		args = append(args, *input.RatingFloor)
		argIdx++
	}
	if input.RatingCeiling != nil {
		setClauses = append(setClauses, fmt.Sprintf("rating_ceiling = $%d", argIdx))
		args = append(args, *input.RatingCeiling)
		argIdx++
	}
	if input.PromotionThreshold != nil {
		setClauses = append(setClauses, fmt.Sprintf("promotion_threshold = $%d", argIdx))
		args = append(args, *input.PromotionThreshold)
		argIdx++
	}
	if input.DemotionThreshold != nil {
		setClauses = append(setClauses, fmt.Sprintf("demotion_threshold = $%d", argIdx))
		args = append(args, *input.DemotionThreshold)
		argIdx++
	}
	if input.IsActive != nil {
		setClauses = append(setClauses, fmt.Sprintf("is_active = $%d", argIdx))
		args = append(args, *input.IsActive)
		argIdx++
	}

	if len(setClauses) == 0 {
		return s.GetSkillGroup(ctx, input.SkillGroupID)
	}

	args = append(args, input.SkillGroupID)
	stmt := fmt.Sprintf(`
UPDATE skill_groups
SET %s
WHERE id = $%d
RETURNING id, league_id, name, display_order, rating_floor, rating_ceiling,
          promotion_threshold, demotion_threshold, is_active, created_at;`,
		strings.Join(setClauses, ", "),
		argIdx,
	)

	row := s.db.QueryRowContext(ctx, stmt, args...)
	sg, err := scanSkillGroup(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return skillgroup.SkillGroup{}, fmt.Errorf("%w: skill group not found", skillgroup.ErrNotFound)
		}
		return skillgroup.SkillGroup{}, mapSkillGroupSQLError(err)
	}
	return sg, nil
}

func (s *SkillGroupStore) RecordTransition(ctx context.Context, input skillgroup.RecordTransitionInput) (skillgroup.SkillGroupTransition, error) {
	const stmt = `
INSERT INTO skill_group_transitions(player_id, from_skill_group_id, to_skill_group_id, rating_at_transition, direction)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, player_id, from_skill_group_id, to_skill_group_id, rating_at_transition, direction, effective_at;`

	row := s.db.QueryRowContext(ctx, stmt,
		input.PlayerID,
		input.FromSkillGroupID,
		input.ToSkillGroupID,
		input.RatingAtTransition,
		input.Direction,
	)
	var t skillgroup.SkillGroupTransition
	var fromSgID sql.NullInt64
	if err := row.Scan(&t.ID, &t.PlayerID, &fromSgID, &t.ToSkillGroupID, &t.RatingAtTransition, &t.Direction, &t.EffectiveAt); err != nil {
		return skillgroup.SkillGroupTransition{}, mapSkillGroupSQLError(err)
	}
	if fromSgID.Valid {
		v := fromSgID.Int64
		t.FromSkillGroupID = &v
	}
	return t, nil
}

func (s *SkillGroupStore) ListTransitionsByPlayerID(ctx context.Context, playerID int64) ([]skillgroup.SkillGroupTransition, error) {
	const stmt = `
SELECT id, player_id, from_skill_group_id, to_skill_group_id, rating_at_transition, direction, effective_at
FROM skill_group_transitions
WHERE player_id = $1
ORDER BY effective_at DESC, id DESC;`

	rows, err := s.db.QueryContext(ctx, stmt, playerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	transitions := make([]skillgroup.SkillGroupTransition, 0)
	for rows.Next() {
		var t skillgroup.SkillGroupTransition
		var fromSgID sql.NullInt64
		if err := rows.Scan(&t.ID, &t.PlayerID, &fromSgID, &t.ToSkillGroupID, &t.RatingAtTransition, &t.Direction, &t.EffectiveAt); err != nil {
			return nil, err
		}
		if fromSgID.Valid {
			v := fromSgID.Int64
			t.FromSkillGroupID = &v
		}
		transitions = append(transitions, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return transitions, nil
}

func (s *SkillGroupStore) GetSkillGroupForRating(ctx context.Context, leagueID int64, rating int32) (skillgroup.SkillGroup, error) {
	const stmt = `
SELECT id, league_id, name, display_order, rating_floor, rating_ceiling,
       promotion_threshold, demotion_threshold, is_active, created_at
FROM skill_groups
WHERE league_id = $1
  AND is_active = TRUE
ORDER BY
    CASE WHEN rating_floor <= $2 AND $2 <= rating_ceiling THEN 0 ELSE 1 END,
    display_order DESC
LIMIT 1;`

	row := s.db.QueryRowContext(ctx, stmt, leagueID, rating)
	sg, err := scanSkillGroup(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return skillgroup.SkillGroup{}, fmt.Errorf("%w: no active skill groups found for league", skillgroup.ErrNotFound)
		}
		return skillgroup.SkillGroup{}, mapSkillGroupSQLError(err)
	}
	return sg, nil
}
