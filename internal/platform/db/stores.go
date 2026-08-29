package db

import (
	"database/sql"

	"github.com/kineticbot/kinetic-v3/internal/domain/hierarchy"
	"github.com/kineticbot/kinetic-v3/internal/platform/replayparser"
)

type Stores struct {
	HierarchyStore   hierarchy.Store
	LeagueStore      hierarchy.LeagueStore
	PlayerStore      hierarchy.PlayerStore
	IdentityStore    hierarchy.IdentityStore
	RosterStore      hierarchy.RosterStore
	RoleStore        hierarchy.RoleStore
	QueueStore       hierarchy.QueueStore
	PlatformStore    hierarchy.PlatformAccountStore
	EligibilityStore hierarchy.EligibilityStore
	ScrimStore       hierarchy.ScrimStore
	RatingStore      hierarchy.RatingStore
	ResultStore      hierarchy.ResultStore
	ReplayStore      hierarchy.ReplayStore
	ExceptionStore   hierarchy.ExceptionStore
	SchedulingStore  hierarchy.SchedulingStore
}

func NewStores(conn *sql.DB) Stores {
	return newStores(NewHierarchyStore(conn))
}

func NewStoresWithReplayParser(conn *sql.DB, parser replayparser.Parser) Stores {
	return newStores(NewHierarchyStoreWithReplayParser(conn, parser))
}

func newStores(hierarchyStore *HierarchyStore) Stores {
	return Stores{
		HierarchyStore:   hierarchyStore,
		LeagueStore:      hierarchyStore,
		PlayerStore:      hierarchyStore,
		IdentityStore:    hierarchyStore,
		RosterStore:      hierarchyStore,
		RoleStore:        hierarchyStore,
		QueueStore:       hierarchyStore,
		PlatformStore:    hierarchyStore,
		EligibilityStore: hierarchyStore,
		ScrimStore:       hierarchyStore,
		RatingStore:      hierarchyStore,
		ResultStore:      hierarchyStore,
		ReplayStore:      hierarchyStore,
		ExceptionStore:   hierarchyStore,
		SchedulingStore:  hierarchyStore,
	}
}
