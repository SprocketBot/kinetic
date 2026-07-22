package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/kineticbot/kinetic-v3/internal/domain/hierarchy"
)

func (s *HierarchyStore) LinkPlatformAccount(ctx context.Context, input hierarchy.LinkPlatformAccountInput) (hierarchy.PlatformAccountLink, error) {
	if err := hierarchy.ValidateLinkPlatformAccountInput(input); err != nil {
		return hierarchy.PlatformAccountLink{}, err
	}

	const stmt = `
INSERT INTO platform_account_links(subject, player_id, provider, provider_account_id, provider_account_name)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, subject, player_id, provider, provider_account_id, provider_account_name, is_active, linked_at, unlinked_at;`
	var link hierarchy.PlatformAccountLink
	err := s.db.QueryRowContext(ctx, stmt, input.Subject, input.PlayerID, input.Provider, input.ProviderAccountID, input.ProviderAccountName).Scan(
		&link.ID,
		&link.Subject,
		&link.PlayerID,
		&link.Provider,
		&link.ProviderAccountID,
		&link.ProviderAccountName,
		&link.IsActive,
		&link.LinkedAt,
		&link.UnlinkedAt,
	)
	if err != nil {
		return hierarchy.PlatformAccountLink{}, mapSQLError(err)
	}
	return link, nil
}

func (s *HierarchyStore) UnlinkPlatformAccount(ctx context.Context, input hierarchy.UnlinkPlatformAccountInput) (hierarchy.PlatformAccountLink, error) {
	if err := hierarchy.ValidateUnlinkPlatformAccountInput(input); err != nil {
		return hierarchy.PlatformAccountLink{}, err
	}

	const stmt = `
UPDATE platform_account_links
SET is_active = FALSE, unlinked_at = NOW()
WHERE subject = $1
  AND provider = $2
  AND provider_account_id = $3
  AND is_active = TRUE
RETURNING id, subject, player_id, provider, provider_account_id, provider_account_name, is_active, linked_at, unlinked_at;`
	var link hierarchy.PlatformAccountLink
	err := s.db.QueryRowContext(ctx, stmt, input.Subject, input.Provider, input.ProviderAccountID).Scan(
		&link.ID,
		&link.Subject,
		&link.PlayerID,
		&link.Provider,
		&link.ProviderAccountID,
		&link.ProviderAccountName,
		&link.IsActive,
		&link.LinkedAt,
		&link.UnlinkedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return hierarchy.PlatformAccountLink{}, fmt.Errorf("%w: active platform account link not found", hierarchy.ErrConflict)
		}
		return hierarchy.PlatformAccountLink{}, mapSQLError(err)
	}
	return link, nil
}

func (s *HierarchyStore) ListPlatformAccountLinks(ctx context.Context, subject string) ([]hierarchy.PlatformAccountLink, error) {
	const stmt = `
SELECT id, subject, player_id, provider, provider_account_id, provider_account_name, is_active, linked_at, unlinked_at
FROM platform_account_links
WHERE subject = $1
ORDER BY id DESC;`
	rows, err := s.db.QueryContext(ctx, stmt, subject)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	links := make([]hierarchy.PlatformAccountLink, 0)
	for rows.Next() {
		var link hierarchy.PlatformAccountLink
		if err := rows.Scan(
			&link.ID,
			&link.Subject,
			&link.PlayerID,
			&link.Provider,
			&link.ProviderAccountID,
			&link.ProviderAccountName,
			&link.IsActive,
			&link.LinkedAt,
			&link.UnlinkedAt,
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

func (s *HierarchyStore) ListPlatformAccountLinksByPlayerID(ctx context.Context, playerID int64) ([]hierarchy.PlatformAccountLink, error) {
	const stmt = `
SELECT id, subject, player_id, provider, provider_account_id, provider_account_name, is_active, linked_at, unlinked_at
FROM platform_account_links
WHERE player_id = $1
ORDER BY id DESC;`
	rows, err := s.db.QueryContext(ctx, stmt, playerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	links := make([]hierarchy.PlatformAccountLink, 0)
	for rows.Next() {
		var link hierarchy.PlatformAccountLink
		if err := rows.Scan(&link.ID, &link.Subject, &link.PlayerID, &link.Provider, &link.ProviderAccountID, &link.ProviderAccountName, &link.IsActive, &link.LinkedAt, &link.UnlinkedAt); err != nil {
			return nil, err
		}
		links = append(links, link)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return links, nil
}

func (s *HierarchyStore) GetEligibilityStatus(ctx context.Context, subject string) (hierarchy.EligibilityStatus, error) {
	if err := hierarchy.ValidateGetEligibilityInput(hierarchy.GetEligibilityInput{Subject: subject}); err != nil {
		return hierarchy.EligibilityStatus{}, err
	}

	const countsStmt = `
SELECT
	COUNT(*) FILTER (WHERE is_active = TRUE) AS active_links,
	COUNT(*) FILTER (WHERE is_active = FALSE) AS inactive_links
FROM platform_account_links
WHERE subject = $1;`
	var activeLinks int64
	var inactiveLinks int64
	if err := s.db.QueryRowContext(ctx, countsStmt, subject).Scan(&activeLinks, &inactiveLinks); err != nil {
		return hierarchy.EligibilityStatus{}, err
	}

	const (
		basePoints      int32 = 100
		activeLinkBonus int32 = 5
		inactivePenalty int32 = 10
		maxPoints       int32 = 200
		minPoints       int32 = 0
		thresholdPoints int32 = 40
		decayPerWeek    int32 = 10
		projectionWeeks int   = 8
	)

	points := basePoints + int32(activeLinks)*activeLinkBonus - int32(inactiveLinks)*inactivePenalty
	if points > maxPoints {
		points = maxPoints
	}
	if points < minPoints {
		points = minPoints
	}

	evaluatedAt := time.Now().UTC()
	eligibleUntil := evaluatedAt
	if points > thresholdPoints {
		weeksEligible := int((points - thresholdPoints) / decayPerWeek)
		eligibleUntil = evaluatedAt.AddDate(0, 0, weeksEligible*7)
	}

	projection := make([]hierarchy.EligibilityProjectionPoint, 0, projectionWeeks+1)
	for week := 0; week <= projectionWeeks; week++ {
		projectedPoints := points - int32(week)*decayPerWeek
		if projectedPoints < minPoints {
			projectedPoints = minPoints
		}
		projection = append(projection, hierarchy.EligibilityProjectionPoint{
			EffectiveAt: evaluatedAt.AddDate(0, 0, week*7),
			Points:      projectedPoints,
			IsEligible:  projectedPoints >= thresholdPoints,
		})
	}

	return hierarchy.EligibilityStatus{
		Subject:         subject,
		Points:          points,
		ThresholdPoints: thresholdPoints,
		DecayPerWeek:    decayPerWeek,
		EligibleUntil:   eligibleUntil,
		EvaluatedAt:     evaluatedAt,
		Projection:      projection,
	}, nil
}
