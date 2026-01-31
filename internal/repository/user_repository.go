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
func (r *UserRepository) Register(username, email, password, role, fullName string, classID int) (*entity.User, error) {
	// Хэшируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	var userID int
	query := `
        INSERT INTO users (username, email, password_hash, role, fullname)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id
    `

	err = r.db.QueryRow(query, username, email, string(hashedPassword), role, fullName).Scan(&userID)
	if err != nil {
		return nil, err
	}

	// Если ученик, добавляем в класс
	if role == "student" {
		_, err = r.db.Exec(
			"INSERT INTO student_classes (student_id, class_id) VALUES ($1, $2)",
			userID, classID,
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

	query := `
        SELECT id, username, email, password_hash, role, fullname, created_at
        FROM users 
        WHERE username = $1 OR email = $1
    `

	err := r.db.QueryRow(query, username).Scan(
		&user.ID, &user.Username, &user.Email, &passwordHash,
		&user.Role, &user.FullName, &user.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Проверяем пароль
	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
	if err != nil {
		return nil, err
	}

	return &user, nil
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

	query := `
        SELECT u.id, u.username, u.email, u.role, u.fullname, u.created_at
        FROM users u
        JOIN user_sessions s ON u.id = s.user_id
        WHERE s.session_token = $1 AND s.expires_at > NOW()
    `

	err := r.db.QueryRow(query, token).Scan(
		&user.ID, &user.Username, &user.Email, &user.Role,
		&user.FullName, &user.CreatedAt,
	)

	if err != nil {
		return nil, err
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

	query := `
        SELECT id, username, email, role, fullname, created_at
        FROM users WHERE id = $1
    `

	err := r.db.QueryRow(query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.Role,
		&user.FullName, &user.CreatedAt,
	)

	return &user, err
}

// Получение учеников класса (для учителя)
func (r *UserRepository) GetStudentsByClass(classID int) ([]entity.User, error) {
	query := `
        SELECT u.id, u.username, u.email, u.role, u.fullname, u.created_at
        FROM users u
        JOIN student_classes sc ON u.id = sc.student_id
        WHERE u.role = 'student' AND sc.class_id = $1
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
		err := rows.Scan(
			&user.ID, &user.Username, &user.Email, &user.Role,
			&user.FullName, &user.CreatedAt,
		)
		if err != nil {
			continue
		}
		students = append(students, user)
	}

	return students, nil
}

// Получение всех классов (для админа)
func (r *UserRepository) GetAllClasses() ([]entity.Class, error) {
	rows, err := r.db.Query(`
        SELECT id, name, grade, teacher_id, created_at 
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
		err := rows.Scan(
			&class.ID, &class.Name, &class.Grade,
			&class.TeacherID, &class.CreatedAt,
		)
		if err != nil {
			return classes, err
		}
		classes = append(classes, class)
	}

	return classes, nil
}

// Получение классов учителя
func (r *UserRepository) GetTeacherClasses(teacherID int) ([]entity.Class, error) {
	rows, err := r.db.Query(`
        SELECT id, name, grade, teacher_id, created_at 
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
		err := rows.Scan(
			&class.ID, &class.Name, &class.Grade,
			&class.TeacherID, &class.CreatedAt,
		)
		if err != nil {
			continue
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
