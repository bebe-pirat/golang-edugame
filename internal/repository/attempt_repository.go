package repository

import (
	"database/sql"
	"edugame/internal/entity"
	"time"
)

type AttemptRepository struct {
	db *sql.DB
}

func NewAttemptRepository(db *sql.DB) *AttemptRepository {
	return &AttemptRepository{db: db}
}

// Сохранить попытку решения
func (a *AttemptRepository) SaveAttempt(attempt entity.Attempt) error {
	_, err := a.db.Exec(`
		INSERT INTO attempts
		(user_id, equation_type_id, equation_text, correct_answer, user_answer, is_correct, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, attempt.UserID, attempt.EquationTypeID, attempt.EquationText, attempt.CorrectAnswer, attempt.UserAnswer, attempt.IsCorrect, time.Now())

	if err != nil {
		return err
	}

	if attempt.IsCorrect {
		_, err = a.db.Exec(`
			UPDATE user_progress
			SET attempts_count = attempts_count + 1,
			correct_count = correct_count + 1,
			last_attempt_at = $1,
			updated_at = $2
			WHERE user_id = $3 AND equation_type_id = $4
		`, time.Now(), time.Now(), attempt.UserID, attempt.EquationTypeID)
	} else {
		_, err = a.db.Exec(`
			UPDATE user_progress
			SET attempts_count = attempts_count + 1,
			last_attempt_at = $1,
			updated_at = $2
			WHERE user_id = $3 AND equation_type_id = $4
		`, time.Now(), time.Now(), attempt.UserID, attempt.EquationTypeID)
	}

	return err
}
