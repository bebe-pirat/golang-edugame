package entity

import "time"

type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	FullName     string    `json:"full_name"`
	ClassID      *int      `json:"class_id,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

type Class struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Grade     int       `json:"grade"`
	TeacherID int       `json:"teacher_id"`
	CreatedAt time.Time `json:"created_at"`
}

type UserSession struct {
	ID           int       `json:"id"`
	UserID       int       `json:"user_id"`
	SessionToken string    `json:"session_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
}
