package authz

import (
	"context"
	"database/sql"
)

func LoadPermissions(ctx context.Context, db *sql.DB) ([]Permission, error) {
	rows, err := db.QueryContext(ctx, `SELECT role_name, resource, action FROM policies`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	permissions := make([]Permission, 0)
	for rows.Next() {
		var permission Permission
		if err := rows.Scan(&permission.Role, &permission.Resource, &permission.Action); err != nil {
			return nil, err
		}
		permissions = append(permissions, permission)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return permissions, nil
}
