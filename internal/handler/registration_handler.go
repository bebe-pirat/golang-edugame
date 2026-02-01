package handler

import (
	"edugame/internal/repository"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"time"
)

type RegistrationHandler struct {
	userRepo *repository.UserRepository
	tmpl     *template.Template
}

func NewRegistrationHandler(userRepo *repository.UserRepository) *RegistrationHandler {
	tmpl := template.Must(template.ParseFiles(
		"../../internal/templates/register.html",
	))

	return &RegistrationHandler{
		userRepo: userRepo,
		tmpl:     tmpl,
	}
}

// RegisterPage - показывает страницу регистрации (GET запрос)
func (h *RegistrationHandler) RegisterPage(w http.ResponseWriter, r *http.Request) {
	// Получаем список классов для выпадающего списка
	classes, err := h.userRepo.GetAllClasses()
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(len(classes))

	data := map[string]interface{}{
		"Title":   "Регистрация",
		"Error":   "",
		"Classes": classes,
		"Form":    map[string]string{},
	}

	h.tmpl.Execute(w, data)
}

// Register - обрабатывает отправку формы (POST запрос)
func (h *RegistrationHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.RegisterPage(w, r)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Ошибка обработки формы", http.StatusBadRequest)
		return
	}

	// Получаем значения из формы
	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")
	fullName := r.FormValue("full_name")
	role := r.FormValue("role")

	// Для учеников получаем класс
	var classID int
	if role == "student" {
		if classStr := r.FormValue("class_id"); classStr != "" {
			id, _ := strconv.Atoi(classStr)
			classID = id
		}
	}

	// Регистрируем пользователя
	user, err := h.userRepo.Register(username, email, password, role, fullName, classID)
	if err != nil {
		// Если ошибка - показываем форму снова, но с сохраненными данными
		classes, _ := h.userRepo.GetAllClasses()
		data := map[string]interface{}{
			"Title":   "Регистрация",
			"Error":   "Ошибка регистрации: " + err.Error(),
			"Classes": classes,
			"Form": map[string]string{
				"username":  username,
				"email":     email,
				"full_name": fullName,
				"role":      role,
			},
		}
		h.tmpl.Execute(w, data)
		return
	}

	sessionToken, err := h.userRepo.CreateSession(user.ID)
	if err != nil {
		// Если не удалось создать сессию, все равно редиректим на вход
		// Но лучше показать ошибку
		http.Redirect(w, r, "/login?error=session_error", http.StatusSeeOther)
		return
	}

	// Устанавливаем cookie с сессией
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    sessionToken,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour), // Сессия на 24 часа
		HttpOnly: true,                           // Защита от XSS
		// Secure: true, // Раскомментировать для HTTPS
	})

	session, _ := store.Get(r, "app-session")
	session.Values["user_id"] = user.ID
	session.Values["username"] = user.Username
	session.Values["role"] = user.Role
	session.Save(r, w)

	// Редирект в зависимости от роли пользователя
	switch user.Role {
	case "student":
		http.Redirect(w, r, "/home", http.StatusSeeOther)
	case "teacher":
		http.Redirect(w, r, "/teacher_home", http.StatusSeeOther)
	// case "admin":
	// 	http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
	default:
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}
