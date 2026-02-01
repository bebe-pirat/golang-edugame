package handler

import (
	"edugame/internal/repository"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"
)

type LoginHandler struct {
	userRepo *repository.UserRepository
	tmpl     *template.Template
}

func NewLoginHandler(userRepo *repository.UserRepository) *LoginHandler {
	tmpl := template.Must(template.ParseFiles(
		"internal/templates/login.html",
	))
	return &LoginHandler{
		userRepo: userRepo,
		tmpl:     tmpl,
	}
}

func (h *LoginHandler) LoginPage(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "app-session")
	if userID, ok := session.Values["user_id"].(int); ok && userID > 0 {
		role, ok := session.Values["role"].(string)
		if !ok {
			role = "student"
		}

		switch role {
		case "teacher":
			http.Redirect(w, r, "/teacher/class", http.StatusSeeOther)
		case "student":
			http.Redirect(w, r, "/home", http.StatusSeeOther)
		case "director":
			http.Redirect(w, r, "/director", http.StatusSeeOther)
		default:
			http.Redirect(w, r, "/index", http.StatusSeeOther)
		}
		return
	}

	data := map[string]interface{}{
		"Title":   "Вход в систему",
		"Error":   r.URL.Query().Get("error"),
		"Message": r.URL.Query().Get("message"),
		"Form": map[string]string{
			"username": r.URL.Query().Get("username"),
		},
	}

	h.tmpl.Execute(w, data)
}

func (h *LoginHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.LoginPage(w, r)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Ошибка обработки формы", http.StatusBadRequest)
		return
	}

	username := strings.TrimSpace(r.FormValue("username"))
	password := r.FormValue("password")

	if username == "" || password == "" {
		http.Redirect(w, r, "/login?error=empty_fields&username="+username, http.StatusSeeOther)
		return
	}

	user, err := h.userRepo.Login(username, password)
	if err != nil {
		fmt.Printf("Ошибка входа для пользователя %s: %v\n", username, err)
		http.Redirect(w, r, "/login?error=invalid_credentials&username="+username, http.StatusSeeOther)
		return
	}

	sessionToken, err := h.userRepo.CreateSession(user.ID)
	if err != nil {
		fmt.Printf("Ошибка создания сессии для пользователя %d: %v\n", user.ID, err)
		http.Redirect(w, r, "/login?error=session_error", http.StatusSeeOther)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    sessionToken,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	session, _ := store.Get(r, "app-session")
	session.Values["user_id"] = user.ID
	session.Values["username"] = user.Username
	session.Values["role"] = user.Role
	session.Values["full_name"] = user.FullName

	if err := session.Save(r, w); err != nil {
		fmt.Printf("Ошибка сохранения сессии Gorilla: %v\n", err)
	}

	fmt.Printf("Успешный вход: %s (ID: %d, Роль: %s)\n",
		user.Username, user.ID, user.Role)

	switch user.Role {
	case "student":
		fmt.Println("hello")
		http.Redirect(w, r, "/home", http.StatusSeeOther)
	case "teacher":
		fmt.Println("1ghfdhf")
		http.Redirect(w, r, "/teacher/class", http.StatusSeeOther)
	case "director":
		http.Redirect(w, r, "/director", http.StatusSeeOther)
	default:
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

// Logout - выход из системы
func (h *LoginHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Удаляем сессию из БД
	cookie, err := r.Cookie("session_token")
	if err == nil {
		h.userRepo.Logout(cookie.Value)
	}

	// Удаляем cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	// Удаляем сессию Gorilla
	session, _ := store.Get(r, "app-session")
	session.Options.MaxAge = -1 // Удаляем сессию
	session.Save(r, w)

	http.Redirect(w, r, "/login?message=Вы+вышли+из+системы", http.StatusSeeOther)
}
