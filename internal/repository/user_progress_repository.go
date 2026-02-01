package repository

import (
	"database/sql"
	"edugame/internal/entity"
	"fmt"
)

type UserProgressRepository struct {
	db *sql.DB
}

func NewUserProgressRepository(db *sql.DB) *UserProgressRepository {
	return &UserProgressRepository{db: db}
}

type TypeStat struct {
	TypeID      int
	Attempts    int
	Correct     int
	LastAttempt sql.NullTime
}

// получить прогресс определенного пользователя по определенному типу уравнений
func (r *UserProgressRepository) GetUserProgressBySpecificEqType(userId, equationTypeId int) (entity.UserProgress, error) {
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
		&up.IsUnlocked,
		&up.FirstUnlockedAt,
		&up.LastAttemptAt,
		&up.CreatedAt,
		&up.UpdatedAt,
	)

	return up, err
}

func (r *UserProgressRepository) GetUserAllProgress(userId int) ([]entity.UserProgress, error) {
	// Используем правильный многострочный литерал
	query := `
SELECT 
    user_progress.id,                     -- 1 id
    user_progress.user_id,                -- 2 userId 
    users.username,                       -- 3 username
    user_progress.equation_type_id,       -- 4 equation_type_id
    equation_types.name,                  -- 5 equation_type_name
    equation_types.description,           -- 6 description
    user_progress.attempts_count,         -- 7 attempts_count
    user_progress.correct_count,          -- 8 correct_count
    user_progress.is_unlocked,            -- 9 is_unlocked
    user_progress.first_unlocked_at,      -- 10 first_unlocked_at
    user_progress.last_attempt_at,        -- 11 last_attempt_at
    user_progress.created_at,             -- 12 created_at
    user_progress.updated_at              -- 13 updated_at
FROM user_progress 
JOIN equation_types ON user_progress.equation_type_id = equation_types.id
JOIN users ON users.id = user_progress.user_id
WHERE user_progress.user_id = $1
`

	rows, err := r.db.Query(query, userId)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ups []entity.UserProgress

	for rows.Next() {
		var up entity.UserProgress

		err := rows.Scan(
			&up.Id,
			&up.UserId,
			&up.Username,
			&up.EquationTypeId,
			&up.EquationTypeName,
			&up.Description,
			&up.AttemptsCount,
			&up.CorrectCount,
			&up.IsUnlocked,
			&up.FirstUnlockedAt,
			&up.LastAttemptAt,
			&up.CreatedAt,
			&up.UpdatedAt,
		)

		if err != nil {
			return ups, err
		}

		ups = append(ups, up)
	}

	if err = rows.Err(); err != nil {
		return ups, err
	}

	return ups, nil
}

// GetWeakTypesId возвращает id тех типов уравнений, в которых пользователь решил менее 70 процентов примеров правильно
func (r *TypeRepository) GetWeakTypesId(list []entity.UserProgress) ([]int, error) {
	if len(list) == 0 {
		return nil, &UserProgressRepositoryError{"No user progress"}
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

type UserProgressRepositoryError struct {
	Message string
}

func (e *UserProgressRepositoryError) Error() string {
	return "User repository error: " + e.Message
}

func (r *UserProgressRepository) GetUserTypeStatistics(userID int) (map[int]TypeStat, error) {
	query := `
    SELECT 
        et.id as type_id,
        COALESCE(up.attempts_count, 0) as attempts_count,
        COALESCE(up.correct_count, 0) as correct_count,
        up.last_attempt_at
    FROM equation_types et
    LEFT JOIN user_progress up ON et.id = up.equation_type_id AND up.user_id = $1
    WHERE et.class = (
        SELECT c.grade 
        FROM classes c
        JOIN student_classes sc ON c.id = sc.class_id
        WHERE sc.student_id = $2
        LIMIT 1
    )
    ORDER BY et.id
`

	rows, err := r.db.Query(query, userID, userID)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer rows.Close()

	stats := make(map[int]TypeStat)
	for rows.Next() {
		var stat TypeStat
		err := rows.Scan(&stat.TypeID, &stat.Attempts, &stat.Correct, &stat.LastAttempt)
		fmt.Println(err)
		if err != nil {
			continue
		}
		stats[stat.TypeID] = stat
	}

	return stats, nil
}
