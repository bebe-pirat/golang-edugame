package repository

import (
	"database/sql"
	"edugame/internal/entity"
	"time"
)

type SchoolRepository struct {
	db *sql.DB
}

func NewSchoolRepository(db *sql.DB) *SchoolRepository {
	return &SchoolRepository{db: db}
}

// GetAll получает все школы
func (r *SchoolRepository) GetAll() ([]entity.School, error) {
	query := `SELECT id, name, address, phone, email, created_at, updated_at FROM schools ORDER BY name`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schools []entity.School
	for rows.Next() {
		var school entity.School
		err := rows.Scan(&school.ID, &school.Name, &school.Address, &school.Phone, &school.Email, &school.CreatedAt, &school.UpdatedAt)
		if err != nil {
			return nil, err
		}
		schools = append(schools, school)
	}

	return schools, nil
}

// GetByID получает школу по ID
func (r *SchoolRepository) GetByID(id int) (*entity.School, error) {
	query := `SELECT id, name, address, phone, email, created_at, updated_at FROM schools WHERE id = $1`

	var school entity.School
	err := r.db.QueryRow(query, id).Scan(&school.ID, &school.Name, &school.Address, &school.Phone, &school.Email, &school.CreatedAt, &school.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &school, nil
}

// Create создает новую школу
func (r *SchoolRepository) Create(name, address, phone, email string) (*entity.School, error) {
	query := `
		INSERT INTO schools (name, address, phone, email)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, address, phone, email, created_at, updated_at
	`

	var school entity.School
	err := r.db.QueryRow(query, name, address, phone, email).Scan(
		&school.ID, &school.Name, &school.Address, &school.Phone, &school.Email, &school.CreatedAt, &school.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &school, nil
}

// Update обновляет школу
func (r *SchoolRepository) Update(id int, name, address, phone, email string) (*entity.School, error) {
	query := `
		UPDATE schools 
		SET name = $1, address = $2, phone = $3, email = $4, updated_at = $5
		WHERE id = $6
		RETURNING id, name, address, phone, email, created_at, updated_at
	`

	var school entity.School
	err := r.db.QueryRow(query, name, address, phone, email, time.Now(), id).Scan(
		&school.ID, &school.Name, &school.Address, &school.Phone, &school.Email, &school.CreatedAt, &school.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &school, nil
}

// Delete удаляет школу
func (r *SchoolRepository) Delete(id int) error {
	query := `DELETE FROM schools WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}
