package repository

import (
	"database/sql"
	"edugame/internal/entity"
)

type RoleRepository struct {
	db *sql.DB
}

func NewRoleRepository(db *sql.DB) *RoleRepository {
	return &RoleRepository{db: db}
}

// GetAll получает все роли
func (r *RoleRepository) GetAll() ([]entity.Role, error) {
	query := `SELECT id, name, description, created_at FROM roles ORDER BY id`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []entity.Role
	for rows.Next() {
		var role entity.Role
		err := rows.Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt)
		if err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}

	return roles, nil
}

// GetByID получает роль по ID
func (r *RoleRepository) GetByID(id int) (*entity.Role, error) {
	query := `SELECT id, name, description, created_at FROM roles WHERE id = $1`

	var role entity.Role
	err := r.db.QueryRow(query, id).Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &role, nil
}

// GetByName получает роль по имени
func (r *RoleRepository) GetByName(name string) (*entity.Role, error) {
	query := `SELECT id, name, description, created_at FROM roles WHERE name = $1`

	var role entity.Role
	err := r.db.QueryRow(query, name).Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &role, nil
}
