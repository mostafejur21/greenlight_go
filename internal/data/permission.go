package data

import (
	"context"
	"database/sql"
	"time"

	"github.com/lib/pq"
)

// Permission slice like ("movie:read" and "movie:write")
type Permissions []string

// check if the Permissions slice contain a specific permission code
func (p Permissions) Include(code string) bool {
	for i := range p {
		if code == p[i] {
			return true
		}
	}
	return false
}

type PermissionModel struct {
	DB *sql.DB
}

// Add the provided permission codes for a specific user.
func (m PermissionModel) AddForUser(userID int64, codes ...string) error {
	query := `
    INSERT INTO users_permissions
    SELECT $1, permissions.id FROM permissions WHERE permissions.code = ANY($2)`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, userID, pq.Array(codes))
	return err

}

// The GetAllForUser() method will return all permission codes for a specific user in a Permissions slice.
func (m PermissionModel) GetAllForUser(userId int64) (Permissions, error) {
	query := `
    SELECT permissions.code
    FROM permissions
    INNER JOIN users_permissions ON user_permissions.permission_id = permission.id
    INNER JOIN users ON users_permissions.user_id = users.id
    WHERE users.id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, userId)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var permissions Permissions
	for rows.Next() {
		var permission string
		err := rows.Scan(&permission)

		if err != nil {
			return nil, err
		}
		permissions = append(permissions, permission)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return permissions, nil
}
