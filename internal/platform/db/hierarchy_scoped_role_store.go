package db

import (
	"context"
	"fmt"

	"github.com/kineticbot/kinetic-v3/internal/domain/hierarchy"
)

// ResolveScopedRoles returns active organization roles for one user's player
// identity in a specific game and hierarchy path. Passing the full ancestor
// path allows an FM assignment to authorize franchise descendants and a GM
// assignment to authorize club descendants without treating either as global.
func (s *HierarchyStore) ResolveScopedRoles(ctx context.Context, input hierarchy.ResolveScopedRolesInput) ([]string, error) {
	if input.UserID <= 0 || input.GameID <= 0 {
		return nil, fmt.Errorf("%w: userId and gameId must be greater than zero", hierarchy.ErrInvalidInput)
	}
	const stmt = `
SELECT DISTINCT ra.role
FROM role_assignments ra
JOIN user_players up ON up.player_id = ra.player_id
WHERE up.user_id = $1
  AND up.game_id = $2
  AND ra.is_active = TRUE
  AND (
    ($3::BIGINT IS NOT NULL AND ra.franchise_id = $3)
    OR ($4::BIGINT IS NOT NULL AND ra.club_id = $4)
    OR ($5::BIGINT IS NOT NULL AND ra.team_id = $5)
  )
ORDER BY ra.role;`
	rows, err := s.db.QueryContext(ctx, stmt, input.UserID, input.GameID, input.FranchiseID, input.ClubID, input.TeamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	roles := make([]string, 0)
	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, rows.Err()
}
