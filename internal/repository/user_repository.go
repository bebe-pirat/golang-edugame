// internal/repository/user_repository.go
package repository

import (
	"crypto/rand"
	"database/sql"
	"edugame/internal/entity"
	"encoding/base64"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Регистрация пользователя
func (r *UserRepository) Register(username, password, roleName, fullName, email string, schoolID *int, classID *int) (*entity.User, error) {
	// Хэшируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Получаем role_id по имени роли
	var roleID int
	err = r.db.QueryRow("SELECT id FROM roles WHERE name = $1", roleName).Scan(&roleID)
	if err != nil {
		return nil, fmt.Errorf("роль '%s' не найдена", roleName)
	}

	var userID int
	var userSchoolID sql.NullInt64
	if schoolID != nil {
		userSchoolID = sql.NullInt64{Int64: int64(*schoolID), Valid: true}
	}

	query := `
        INSERT INTO users (username, password_hash, role_id, fullname, email, school_id)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id
    `

	err = r.db.QueryRow(query, username, string(hashedPassword), roleID, fullName, email, userSchoolID).Scan(&userID)
	if err != nil {
		return nil, err
	}

	// Если ученик, добавляем в класс
	if roleName == "student" && classID != nil {
		_, err = r.db.Exec(
			"INSERT INTO student_classes (student_id, class_id) VALUES ($1, $2)",
			userID, *classID,
		)
		if err != nil {
			return nil, err
		}
	}

	return r.GetByID(userID)
}

// Аутентификация
func (r *UserRepository) Login(username, password string) (*entity.User, error) {
	var user entity.User
	var passwordHash string
	var roleID int

	query := `
        SELECT id, username, password_hash, role_id, fullname, email, school_id, created_at
        FROM users 
        WHERE username = $1
    `

	err := r.db.QueryRow(query, username).Scan(
		&user.ID, &user.Username, &passwordHash,
		&roleID, &user.FullName, &user.Email, &user.SchoolID, &user.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	user.RoleID = roleID

	// Получаем информацию о роли
	role, err := r.getRoleByID(roleID)
	if err == nil {
		user.Role = role
	}

	// Проверяем пароль
	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// Получение роли по ID
func (r *UserRepository) getRoleByID(roleID int) (*entity.Role, error) {
	query := `SELECT id, name, description, created_at FROM roles WHERE id = $1`
	var role entity.Role
	err := r.db.QueryRow(query, roleID).Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// Создание сессии
func (r *UserRepository) CreateSession(userID int) (string, error) {
	// Генерируем токен
	token, err := generateToken()
	if err != nil {
		return "", err
	}

	expiresAt := time.Now().Add(24 * time.Hour) // Сессия на 24 часа

	_, err = r.db.Exec(`
        INSERT INTO user_sessions (user_id, session_token, expires_at)
        VALUES ($1, $2, $3)
    `, userID, token, expiresAt)

	return token, err
}

// Получение пользователя по токену
func (r *UserRepository) GetBySessionToken(token string) (*entity.User, error) {
	var user entity.User
	var roleID int

	query := `
        SELECT u.id, u.username, u.role_id, u.fullname, u.email, u.school_id, u.created_at
        FROM users u
        JOIN user_sessions s ON u.id = s.user_id
        WHERE s.session_token = $1 AND s.expires_at > NOW()
    `

	err := r.db.QueryRow(query, token).Scan(
		&user.ID, &user.Username, &roleID,
		&user.FullName, &user.Email, &user.SchoolID, &user.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	user.RoleID = roleID

	// Получаем информацию о роли
	role, err := r.getRoleByID(roleID)
	if err == nil {
		user.Role = role
	}

	return &user, nil
}

// Удаление сессии (выход)
func (r *UserRepository) Logout(token string) error {
	_, err := r.db.Exec("DELETE FROM user_sessions WHERE session_token = $1", token)
	return err
}

// Получение пользователя по ID
func (r *UserRepository) GetByID(id int) (*entity.User, error) {
	var user entity.User
	var roleID int
	var schoolID sql.NullInt64

	query := `
        SELECT id, username, role_id, fullname, email, school_id, created_at
        FROM users WHERE id = $1
    `

	err := r.db.QueryRow(query, id).Scan(
		&user.ID, &user.Username, &roleID,
		&user.FullName, &user.Email, &schoolID, &user.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	user.RoleID = roleID
	if schoolID.Valid {
		sid := int(schoolID.Int64)
		user.SchoolID = &sid
	}

	// Получаем информацию о роли
	role, err := r.getRoleByID(roleID)
	if err == nil {
		user.Role = role
	}

	return &user, err
}

// GetAllUsers получает всех пользователей
func (r *UserRepository) GetAllUsers() ([]entity.User, error) {
	query := `
        SELECT id, username, role_id, fullname, email, school_id, created_at
        FROM users
        ORDER BY created_at DESC
    `

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []entity.User
	for rows.Next() {
		var user entity.User
		var roleID int
		var schoolID sql.NullInt64

		err := rows.Scan(
			&user.ID, &user.Username, &roleID,
			&user.FullName, &user.Email, &schoolID, &user.CreatedAt,
		)
		if err != nil {
			continue
		}

		user.RoleID = roleID
		if schoolID.Valid {
			sid := int(schoolID.Int64)
			user.SchoolID = &sid
		}

		// Получаем информацию о роли
		role, err := r.getRoleByID(roleID)
		if err == nil {
			user.Role = role
		}

		users = append(users, user)
	}

	return users, nil
}

// GetUserByRoleType получает пользователей по типу роли
func (r *UserRepository) GetUserByRoleType(roleName string) ([]entity.User, error) {
	query := `
        SELECT u.id, u.username, u.role_id, u.fullname, u.email, u.school_id, u.created_at
        FROM users u
        JOIN roles r ON u.role_id = r.id
        WHERE r.name = $1
        ORDER BY u.created_at DESC
    `

	rows, err := r.db.Query(query, roleName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []entity.User
	for rows.Next() {
		var user entity.User
		var roleID int
		var schoolID sql.NullInt64

		err := rows.Scan(
			&user.ID, &user.Username, &roleID,
			&user.FullName, &user.Email, &schoolID, &user.CreatedAt,
		)
		if err != nil {
			continue
		}

		user.RoleID = roleID
		if schoolID.Valid {
			sid := int(schoolID.Int64)
			user.SchoolID = &sid
		}

		// Получаем информацию о роли
		role, err := r.getRoleByID(roleID)
		if err == nil {
			user.Role = role
		}

		users = append(users, user)
	}

	return users, nil
}

// UpdateUser обновляет пользователя
func (r *UserRepository) UpdateUser(id int, username, fullName, email string, roleID int, schoolID *int) (*entity.User, error) {
	var userSchoolID sql.NullInt64
	if schoolID != nil {
		userSchoolID = sql.NullInt64{Int64: int64(*schoolID), Valid: true}
	}

	query := `
        UPDATE users
        SET username = $1, fullname = $2, email = $3, role_id = $4, school_id = $5
        WHERE id = $6
        RETURNING id, username, role_id, fullname, email, school_id, created_at
    `

	var newUser entity.User
	var retSchoolID sql.NullInt64
	err := r.db.QueryRow(query, username, fullName, email, roleID, userSchoolID, id).Scan(
		&newUser.ID, &newUser.Username, &newUser.RoleID,
		&newUser.FullName, &newUser.Email, &retSchoolID, &newUser.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	if retSchoolID.Valid {
		sid := int(retSchoolID.Int64)
		newUser.SchoolID = &sid
	}

	// Получаем информацию о роли
	role, err := r.getRoleByID(newUser.RoleID)
	if err == nil {
		newUser.Role = role
	}

	return &newUser, nil
}

// DeleteUser удаляет пользователя
func (r *UserRepository) DeleteUser(id int) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

// Получение учеников класса (для учителя)
func (r *UserRepository) GetStudentsByClass(classID int) ([]entity.User, error) {
	query := `
        SELECT u.id, u.username, u.role_id, u.fullname, u.email, u.school_id, u.created_at
        FROM users u
        JOIN student_classes sc ON u.id = sc.student_id
        JOIN roles r ON u.role_id = r.id
        WHERE r.name = 'student' AND sc.class_id = $1
        ORDER BY u.fullname
    `

	rows, err := r.db.Query(query, classID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var students []entity.User
	for rows.Next() {
		var user entity.User
		var roleID int
		var schoolID sql.NullInt64

		err := rows.Scan(
			&user.ID, &user.Username, &roleID,
			&user.FullName, &user.Email, &schoolID, &user.CreatedAt,
		)
		if err != nil {
			continue
		}

		user.RoleID = roleID
		if schoolID.Valid {
			sid := int(schoolID.Int64)
			user.SchoolID = &sid
		}

		// Получаем информацию о роли
		role, err := r.getRoleByID(roleID)
		if err == nil {
			user.Role = role
		}

		students = append(students, user)
	}

	return students, nil
}

// Получение всех классов (для админа)
func (r *UserRepository) GetAllClasses() ([]entity.Class, error) {
	rows, err := r.db.Query(`
        SELECT id, name, grade, teacher_id, school_id, created_at 
        FROM classes 
        ORDER BY grade, name
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var classes []entity.Class
	for rows.Next() {
		var class entity.Class
		var schoolID sql.NullInt64

		err := rows.Scan(
			&class.ID, &class.Name, &class.Grade,
			&class.TeacherID, &schoolID, &class.CreatedAt,
		)
		if err != nil {
			return classes, err
		}

		if schoolID.Valid {
			sid := int(schoolID.Int64)
			class.SchoolID = &sid
		}

		classes = append(classes, class)
	}

	return classes, nil
}

// Получение классов учителя
func (r *UserRepository) GetTeacherClasses(teacherID int) ([]entity.Class, error) {
	rows, err := r.db.Query(`
        SELECT id, name, grade, teacher_id, school_id, created_at 
        FROM classes 
        WHERE teacher_id = $1
        ORDER BY grade, name
    `, teacherID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var classes []entity.Class
	for rows.Next() {
		var class entity.Class
		var schoolID sql.NullInt64

		err := rows.Scan(
			&class.ID, &class.Name, &class.Grade,
			&class.TeacherID, &schoolID, &class.CreatedAt,
		)
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

// Вспомогательная функция для генерации токена
func generateToken() (string, error) {
	// Используем crypto/rand для безопасной генерации
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// internal/repository/user_repository.go
func (r *UserRepository) GetStudentClass(studentID int) (int, error) {
	var classGrade int

	// Получаем класс ученика через таблицу student_classes
	query := `
        SELECT c.grade 
        FROM classes c
        JOIN student_classes sc ON c.id = sc.class_id
        WHERE sc.student_id = $1
        LIMIT 1
    `

	err := r.db.QueryRow(query, studentID).Scan(&classGrade)
	if err != nil {
		if err == sql.ErrNoRows {
			// Ученик не привязан к классу
			return 0, fmt.Errorf("ученик не привязан к классу")
		}
		return 0, err
	}

	return classGrade, nil
}
