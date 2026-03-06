package authz

import (
	"context"
	"database/sql"
)

type Permission struct {
	Role     string
	Resource string
	Action   string
}

// AuthzContext carries optional scope identifiers for the resource being acted
// on. Any nil pointer means "global" — i.e. no scope restriction applies for
// that dimension.
type AuthzContext struct {
	// FranchiseID is set when the resource is scoped to a specific franchise.
	FranchiseID *int64
	// ClubID is set when the resource is scoped to a specific club.
	ClubID *int64
	// TeamID is set when the resource is scoped to a specific team.
	TeamID *int64
}

// GlobalContext returns an AuthzContext with no scope restrictions.
func GlobalContext() AuthzContext {
	return AuthzContext{}
}

// Evaluator checks whether a set of roles is allowed to perform an action on a
// resource. AllowedInContext extends this with optional organizational scope.
type Evaluator interface {
	// Allowed checks global role permission without scope.
	Allowed(roles []string, resource, action string) bool
	// AllowedInContext checks whether the principal (identified by their roles
	// plus the scoped role assignments carried in authzRoles) can perform
	// action on resource within the given scope. globalRoles are the roles from
	// the session token / JWT; scopedRoles are role names that have been
	// verified against the DB for the specific scope in authzCtx.
	AllowedInContext(globalRoles []string, scopedRoles []string, resource, action string, authzCtx AuthzContext) bool
}

type StaticEvaluator struct {
	rolePermissions map[string][]Permission
}

func DefaultPermissions() []Permission {
	return []Permission{
		{Role: "admin", Resource: "admin.ping", Action: "read"},
		// Global admin and operator have full access to privileged resources.
		{Role: "admin", Resource: "*", Action: "*"},
		{Role: "operator", Resource: "*", Action: "*"},
		// Franchise manager can manage their franchise scope.
		{Role: "fm", Resource: ResourceRoleAssignment, Action: ActionCreate},
		{Role: "fm", Resource: ResourceRoleAssignment, Action: ActionRevoke},
		{Role: "fm", Resource: ResourceRosterMembership, Action: ActionCreate},
		{Role: "fm", Resource: ResourceRosterMembership, Action: ActionDelete},
		{Role: "fm", Resource: ResourceResultOverride, Action: ActionCreate},
		{Role: "fm", Resource: ResourceResultSubmission, Action: ActionRatify},
		{Role: "fm", Resource: ResourceResultSubmission, Action: ActionReject},
		{Role: "fm", Resource: ResourceResultSubmission, Action: ActionReset},
		{Role: "fm", Resource: ResourcePlayer, Action: ActionUpdate},
		// General manager can manage their club scope.
		{Role: "gm", Resource: ResourceRoleAssignment, Action: ActionCreate},
		{Role: "gm", Resource: ResourceRoleAssignment, Action: ActionRevoke},
		{Role: "gm", Resource: ResourceRosterMembership, Action: ActionCreate},
		{Role: "gm", Resource: ResourceRosterMembership, Action: ActionDelete},
		{Role: "gm", Resource: ResourceResultSubmission, Action: ActionRatify},
		{Role: "gm", Resource: ResourceResultSubmission, Action: ActionReject},
		// Captain can manage their team scope.
		{Role: "captain", Resource: ResourceRoleAssignment, Action: ActionCreate},
		{Role: "captain", Resource: ResourceRoleAssignment, Action: ActionRevoke},
		{Role: "captain", Resource: ResourceResultSubmission, Action: ActionRatify},
		{Role: "captain", Resource: ResourceResultSubmission, Action: ActionReject},
		{Role: "captain", Resource: ResourceScrim, Action: ActionCreate},
		// Any authenticated player can check in to a scrim and submit results.
		{Role: "player", Resource: ResourceScrim, Action: ActionUpdate},
		{Role: "player", Resource: ResourceResultSubmission, Action: ActionCreate},
		{Role: "player", Resource: ResourceReplayEvidence, Action: ActionCreate},
	}
}

func NewStaticEvaluator(permissions []Permission) StaticEvaluator {
	rolePermissions := map[string][]Permission{}
	for _, permission := range permissions {
		rolePermissions[permission.Role] = append(rolePermissions[permission.Role], permission)
	}
	return StaticEvaluator{rolePermissions: rolePermissions}
}

func (e StaticEvaluator) Allowed(roles []string, resource, action string) bool {
	for _, role := range roles {
		permissions := e.rolePermissions[role]
		for _, permission := range permissions {
			if matches(permission.Resource, resource) && matches(permission.Action, action) {
				return true
			}
		}
	}
	return false
}

// AllowedInContext combines the global role check (globalRoles) with a check
// against scopedRoles that have been validated for the specific authzCtx. If
// either set grants the permission, access is allowed.
func (e StaticEvaluator) AllowedInContext(globalRoles []string, scopedRoles []string, resource, action string, _ AuthzContext) bool {
	if e.Allowed(globalRoles, resource, action) {
		return true
	}
	return e.Allowed(scopedRoles, resource, action)
}

func matches(policyValue, requestValue string) bool {
	return policyValue == "*" || policyValue == requestValue
}

func NewDatabaseBackedEvaluator(ctx context.Context, db *sql.DB, fallback []Permission) (Evaluator, error) {
	permissions, err := LoadPermissions(ctx, db)
	if err != nil {
		return nil, err
	}
	if len(permissions) == 0 {
		permissions = fallback
	}
	evaluator := NewStaticEvaluator(permissions)
	return evaluator, nil
}
