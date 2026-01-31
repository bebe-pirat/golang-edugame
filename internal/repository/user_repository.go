package repository

import (
	"database/sql"
	"time"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// получить пользователя тест
func (r *UserRepository) GetTestUserId() (int, error) {
	var userID int
	err := r.db.QueryRow(`
       SELECT id FROM users WHERE username = $1
    `, "тест").Scan(&userID)

	return userID, err
}

// получить имя пользователя по id
func (r *UserRepository) GetUserNameById(userId int) (string, error) {
	var username string
	err := r.db.QueryRow(`
       SELECT id FROM users WHERE id = ?
    `, userId).Scan(&username)

	return username, err
}

// получить класс пользователя тест
func (r *UserRepository) GetTestClassbyUserId(id int) (int, error) {
	var class int
	err := r.db.QueryRow(`
       SELECT class FROM users WHERE id = $1
    `, id).Scan(&class)

	return class, err
}

// Создать сессию для пользователя
func (r *UserRepository) CreateSession(userID int) (string, error) {
	var sessionID string
	err := r.db.QueryRow(`
        INSERT INTO sessions (user_id, created_at, last_activity) 
        VALUES ($1, $2, $3)
        RETURNING id::text
    `, userID, time.Now(), time.Now()).Scan(&sessionID)

	return sessionID, err
}
