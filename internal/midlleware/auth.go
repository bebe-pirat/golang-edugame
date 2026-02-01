package middleware

import (
	"net/http"
	"strings"

	"github.com/gorilla/sessions"
)

var store = sessions.NewCookieStore([]byte("a-very-secret-key"))

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

// RequireRoles - middleware для проверки нескольких ролей
func RequireRoles(allowedRoles []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, _ := store.Get(r, "app-session")

			role, _ := session.Values["role"].(string)
			success := false
			for _, value := range allowedRoles {
				if role == value {
					success = true
				}
			}

			if !success {
				http.Redirect(w, r, "/index", http.StatusSeeOther)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
