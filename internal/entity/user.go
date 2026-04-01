package entity

import "time"

type Role struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type School struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Address   string    `json:"address"`
	Phone     string    `json:"phone"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	RoleID       int       `json:"role_id"`
	Role         *Role     `json:"role,omitempty"`
	FullName     string    `json:"full_name"`
	ClassID      *int      `json:"class_id,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

type Class struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Grade     int       `json:"grade"`
	TeacherID int       `json:"teacher_id"`
	SchoolID  *int      `json:"school_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type UserSession struct {
	ID           int       `json:"id"`
	UserID       int       `json:"user_id"`
	SessionToken string    `json:"session_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
}
