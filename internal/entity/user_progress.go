package entity

import "database/sql"

type UserProgress struct {
	Id               int
	UserId           int
	Username         string
	EquationTypeId   int
	EquationTypeName string
	Description      string
	AttemptsCount    int
	CorrectCount     int

	IsUnlocked bool

	FirstUnlockedAt sql.NullString
	LastAttemptAt   sql.NullString

	CreatedAt sql.NullString
	UpdatedAt sql.NullString
}

// тут нечего делать
