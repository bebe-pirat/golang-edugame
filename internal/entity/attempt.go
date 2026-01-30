package entity

import "time"

type Attempt struct {
	ID             int       `json:"id"`
	UserID         int       `json:"user_id"`
	EquationTypeID int       `json:"equation_type_id"`
	EquationText   string    `json:"equation_text"`
	CorrectAnswer  string    `json:"correct_answer"`
	UserAnswer     string    `json:"user_answer"`
	IsCorrect      bool      `json:"is_correct"`
	CreatedAt      time.Time `json:"created_at"`
}

func NewAttempt(userId, equationTypeId int, equationText, correctAnswer, userAnswer string) Attempt {
	return Attempt{
		UserID:         userId,
		EquationTypeID: equationTypeId,
		EquationText:   equationText,
		CorrectAnswer:  correctAnswer,
		UserAnswer:     userAnswer,
	}
}
