package session

import (
	"github.com/gorilla/sessions"
)

// Store - глобальный экземпляр хранилища сессий
var Store *sessions.CookieStore

// InitStore инициализирует хранилище сессий с указанным секретным ключом
func InitStore(secretKey string) {
	Store = sessions.NewCookieStore([]byte(secretKey))
}

// GetStore возвращает текущее хранилище сессий
func GetStore() *sessions.CookieStore {
	return Store
}
