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

type Evaluator interface {
	Allowed(roles []string, resource, action string) bool
}

type StaticEvaluator struct {
	rolePermissions map[string][]Permission
}

func DefaultPermissions() []Permission {
	return []Permission{
		{Role: "admin", Resource: "admin.ping", Action: "read"},
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
