package entity

type UserProgress struct {
	Id               int
	UserId           int
	Username         string
	EquationTypeId   int
	EquationTypeName string
	Description      string
	AttemptsCount    int
	CorrectCount     int
	BestTime         int

	IsUnlocked bool

	FirstUnlockedAt string
	LastAttemptAt   string

	CreatedAt string
	UpdatedAt string
}

// тут нечего делать
