package repository

import (
	"database/sql"
	"edugame/internal/entity"
	"time"
)

type ClassRepository struct {
	db *sql.DB
}

func NewClassRepository(db *sql.DB) *ClassRepository {
	return &ClassRepository{db: db}
}

// GetAll получает все классы
func (r *ClassRepository) GetAll() ([]entity.Class, error) {
	query := `SELECT id, name, grade, teacher_id, school_id, created_at FROM classes ORDER BY grade, name`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var classes []entity.Class
	for rows.Next() {
		var class entity.Class
		var schoolID sql.NullInt64

		err := rows.Scan(&class.ID, &class.Name, &class.Grade, &class.TeacherID, &schoolID, &class.CreatedAt)
		if err != nil {
			continue
		}

		if schoolID.Valid {
			sid := int(schoolID.Int64)
			class.SchoolID = &sid
		}

		classes = append(classes, class)
	}

	return classes, nil
}

// GetByID получает класс по ID
func (r *ClassRepository) GetByID(id int) (*entity.Class, error) {
	query := `SELECT id, name, grade, teacher_id, school_id, created_at FROM classes WHERE id = $1`

	var class entity.Class
	var schoolID sql.NullInt64

	err := r.db.QueryRow(query, id).Scan(&class.ID, &class.Name, &class.Grade, &class.TeacherID, &schoolID, &class.CreatedAt)
	if err != nil {
		return nil, err
	}

	if schoolID.Valid {
		sid := int(schoolID.Int64)
		class.SchoolID = &sid
	}

	return &class, nil
}

// Create создает новый класс
func (r *ClassRepository) Create(name string, grade, teacherID int, schoolID *int) (*entity.Class, error) {
	var schoolIDNull sql.NullInt64
	if schoolID != nil {
		schoolIDNull = sql.NullInt64{Int64: int64(*schoolID), Valid: true}
	}

	query := `
		INSERT INTO classes (name, grade, teacher_id, school_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, grade, teacher_id, school_id, created_at
	`

	var class entity.Class
	var retSchoolID sql.NullInt64

	err := r.db.QueryRow(query, name, grade, teacherID, schoolIDNull).Scan(
		&class.ID, &class.Name, &class.Grade, &class.TeacherID, &retSchoolID, &class.CreatedAt)
	if err != nil {
		return nil, err
	}

	if retSchoolID.Valid {
		sid := int(retSchoolID.Int64)
		class.SchoolID = &sid
	}

	return &class, nil
}

// Update обновляет класс
func (r *ClassRepository) Update(id int, name string, grade, teacherID int, schoolID *int) (*entity.Class, error) {
	var schoolIDNull sql.NullInt64
	if schoolID != nil {
		schoolIDNull = sql.NullInt64{Int64: int64(*schoolID), Valid: true}
	}

	query := `
		UPDATE classes 
		SET name = $1, grade = $2, teacher_id = $3, school_id = $4
		WHERE id = $5
		RETURNING id, name, grade, teacher_id, school_id, created_at
	`

	var class entity.Class
	var retSchoolID sql.NullInt64

	err := r.db.QueryRow(query, name, grade, teacherID, schoolIDNull, id).Scan(
		&class.ID, &class.Name, &class.Grade, &class.TeacherID, &retSchoolID, &class.CreatedAt)
	if err != nil {
		return nil, err
	}

	if retSchoolID.Valid {
		sid := int(retSchoolID.Int64)
		class.SchoolID = &sid
	}

	return &class, nil
}

// Delete удаляет класс
func (r *ClassRepository) Delete(id int) error {
	query := `DELETE FROM classes WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

// GetByTeacherID получает классы учителя
func (r *ClassRepository) GetByTeacherID(teacherID int) ([]entity.Class, error) {
	query := `SELECT id, name, grade, teacher_id, school_id, created_at FROM classes WHERE teacher_id = $1 ORDER BY grade, name`

	rows, err := r.db.Query(query, teacherID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var classes []entity.Class
	for rows.Next() {
		var class entity.Class
		var schoolID sql.NullInt64

		err := rows.Scan(&class.ID, &class.Name, &class.Grade, &class.TeacherID, &schoolID, &class.CreatedAt)
		if err != nil {
			continue
		}

		if schoolID.Valid {
			sid := int(schoolID.Int64)
			class.SchoolID = &sid
		}

		classes = append(classes, class)
	}

	return classes, nil
}

// GetBySchoolID получает классы школы
func (r *ClassRepository) GetBySchoolID(schoolID int) ([]entity.Class, error) {
	query := `SELECT id, name, grade, teacher_id, school_id, created_at FROM classes WHERE school_id = $1 ORDER BY grade, name`

	rows, err := r.db.Query(query, schoolID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var classes []entity.Class
	for rows.Next() {
		var class entity.Class
		var retSchoolID sql.NullInt64

		err := rows.Scan(&class.ID, &class.Name, &class.Grade, &class.TeacherID, &retSchoolID, &class.CreatedAt)
		if err != nil {
			continue
		}

		if retSchoolID.Valid {
			sid := int(retSchoolID.Int64)
			class.SchoolID = &sid
		}

		classes = append(classes, class)
	}

	return classes, nil
}

// GetStudentsCount получает количество учеников в классе
func (r *ClassRepository) GetStudentsCount(classID int) (int, error) {
	query := `SELECT COUNT(*) FROM student_classes WHERE class_id = $1`

	var count int
	err := r.db.QueryRow(query, classID).Scan(&count)
	return count, err
}

// AddStudentToClass добавляет ученика в класс
func (r *ClassRepository) AddStudentToClass(studentID, classID int) error {
	query := `INSERT INTO student_classes (student_id, class_id) VALUES ($1, $2)`
	_, err := r.db.Exec(query, studentID, classID)
	return err
}

// RemoveStudentFromClass удаляет ученика из класса
func (r *ClassRepository) RemoveStudentFromClass(studentID, classID int) error {
	query := `DELETE FROM student_classes WHERE student_id = $1 AND class_id = $2`
	_, err := r.db.Exec(query, studentID, classID)
	return err
}
