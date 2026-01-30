package repository

import (
	"database/sql"
	"edugame/internal/entity"
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

// Сохранить попытку решения
func (r *UserRepository) SaveAttempt(userID int, equationTypeId int, equationText, correctAnswer, userAnswer string, isCorrect bool) error {
	_, err := r.db.Exec(`
		INSERT INTO attempts
		(user_id, equation_type_id, equation_text, correct_answer, user_answer, is_correct, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, userID, equationTypeId, equationText, correctAnswer, userAnswer, isCorrect, time.Now())

	if err != nil {
		return err
	}

	if isCorrect {
		_, err = r.db.Exec(`
			UPDATE user_progress
			SET attempts_count = attempts_count + 1,
			correct_count = correct_count + 1,
			best_time = 0,
			last_attempt_at = $1
			updated_at = $2
		`, time.Now(), time.Now())
	} else {
		_, err = r.db.Exec(`
			UPDATE user_progress
			SET attempts_count = attempts_count + 1,
			best_time = 0,
			last_attempt_at = $1
			updated_at = $2
		`, time.Now(), time.Now())
	}

	return err
}

func (r *UserRepository) GetUserProgressBySpecificEqType(userId, equationTypeId int) (entity.UserProgress, error) {
	var up entity.UserProgress

	err := r.db.QueryRow(`
		SELECT id, 
			?, 
			username, 
			?, 
			name, 
			description,
			attempts_count, 
			correct_count, 
			best_time, 
			is_unlocked, 
			first_unlocked_at, 
			last_attempt_at,
			created_at, updated_at
		FROM user_progress 
		JOIN equation_types ON user_progress.equation_type_id = equation_types.id
		JOIN users ON users.id = user_progress.user_id
	`, userId, equationTypeId).Scan(
		&up.Id,
		&up.UserId,
		&up.Username,
		&up.EquationTypeId,
		&up.EquationTypeName,
		&up.Description,
		&up.AttemptsCount,
		&up.CorrectCount,
		&up.BestTime,
		&up.IsUnlocked,
		&up.FirstUnlockedAt,
		&up.LastAttemptAt,
		&up.CreatedAt,
		&up.UpdatedAt,
	)

	return up, err
}

// получить прогресс по всем типам уравнений
func (r *UserRepository) GetUserAllProgress(userId int) ([]entity.UserProgress, error) {
	rows, err := r.db.Query(`
		SELECT id, 
			?, 
			username, 
			equation_type_id, 
			name, 
			description,
			attempts_count, 
			correct_count, 
			best_time, 
			is_unlocked, 
			first_unlocked_at, 
			last_attempt_at,
			created_at, updated_at
		FROM user_progress 
		JOIN equation_types ON user_progress.equation_type_id = equation_types.id
		JOIN users ON users.id = user_progress.user_id
	`, userId)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var ups []entity.UserProgress

	for rows.Next() {
		var up entity.UserProgress

		if err := rows.Scan(&up.Id,
			&up.UserId,
			&up.Username,
			&up.EquationTypeId,
			&up.EquationTypeName,
			&up.Description,
			&up.AttemptsCount,
			&up.CorrectCount,
			&up.BestTime,
			&up.IsUnlocked,
			&up.FirstUnlockedAt,
			&up.LastAttemptAt,
			&up.CreatedAt,
			&up.UpdatedAt); err != nil {
			return ups, err
		}

		ups = append(ups, up)
	}

	if err = rows.Err(); err != nil {
		return ups, err
	}

	return ups, err
}

// GetWeakTypesId возвращает id тех типов уравнений, в которых пользователь решил менее 70 процентов примеров правильно
func (r *TypeRepository) GetWeakTypesId(list []entity.UserProgress) ([]int, error) {
	if len(list) == 0 {
		return nil, &UserRepositoryError{"No user progress"}
	}

	weakTypesId := make([]int, 0)

	for _, value := range list {
		accuracy := float64(value.CorrectCount) / float64(value.AttemptsCount)
		if accuracy < 0.7 {
			weakTypesId = append(weakTypesId, value.EquationTypeId)
		}
	}

	return weakTypesId, nil
}

type UserRepositoryError struct {
	Message string
}

func (e *UserRepositoryError) Error() string {
	return "User repository error: " + e.Message
}
