package middleware

import (
	"net/http"
	"strings"

	"github.com/gorilla/sessions"
)

var store = sessions.NewCookieStore([]byte("a-very-secret-key"))

// RequireAuth - middleware для проверки авторизации
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		publicPaths := []string{
			"/",
			"/login",
			"/auth/login",
			"/register",
			"/auth/register",
			"/static/",
		}

		path := r.URL.Path

		for _, publicPath := range publicPaths {
			if path == publicPath || strings.HasPrefix(path, publicPath) {
				next.ServeHTTP(w, r)
				return
			}
		}

		session, _ := store.Get(r, "app-session")

		userID, userIDOk := session.Values["user_id"].(int)
		role, roleOk := session.Values["role"].(string)

		if !userIDOk || userID == 0 {
			session.Values["redirect_after_login"] = path
			session.Save(r, w)

			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		if roleOk {
			if role == "teacher" && (path == "/home" || path == "/equation" || path == "/stats") {
				http.Redirect(w, r, "/teacher_home", http.StatusSeeOther)
				return
			}

			if role == "student" && strings.HasPrefix(path, "/teacher") {
				http.Redirect(w, r, "/home", http.StatusSeeOther)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

// RequireRole - middleware для проверки конкретной роли
func RequireRole(requiredRole string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, _ := store.Get(r, "app-session")

			role, ok := session.Values["role"].(string)
			if !ok || role != requiredRole {
				if role == "teacher" && requiredRole == "student" {
					http.Redirect(w, r, "/teacher_home", http.StatusSeeOther)
				} else if role == "student" && requiredRole == "teacher" {
					http.Redirect(w, r, "/home", http.StatusSeeOther)
				} else {
					http.Redirect(w, r, "/login", http.StatusSeeOther)
				}
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
