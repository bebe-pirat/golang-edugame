package middleware

import (
	"edugame/internal/session"
	"log/slog"
	"net/http"
	"strings"
)

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

		store := session.GetStore()
		if store == nil {
			http.Error(w, "Session store not initialized", http.StatusInternalServerError)
			return
		}

		sess, _ := store.Get(r, "app-session")

		userID, userIDOk := sess.Values["user_id"].(int)
		role, roleOk := sess.Values["role"].(string)

		if !userIDOk || userID == 0 {
			sess.Values["redirect_after_login"] = path
			sess.Save(r, w)

			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		if roleOk {
			if role == "teacher" && (path == "/home" || path == "/equation" || path == "/stats") {
				http.Redirect(w, r, "/teacher/class", http.StatusSeeOther)
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
			store := session.GetStore()
			if store == nil {
				http.Error(w, "Session store not initialized", http.StatusInternalServerError)
				slog.Info("failed to get session")
				return
			}

			sess, _ := store.Get(r, "app-session")
			role, _ := sess.Values["role"].(string)
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
