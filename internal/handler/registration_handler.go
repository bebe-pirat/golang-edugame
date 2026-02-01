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
		"internal/templates/register.html",
	))

	return &RegistrationHandler{
		userRepo: userRepo,
		tmpl:     tmpl,
	}
}

func (h *RegistrationHandler) RegisterPage(w http.ResponseWriter, r *http.Request) {
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

func (h *RegistrationHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.RegisterPage(w, r)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Ошибка обработки формы", http.StatusBadRequest)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")
	fullName := r.FormValue("full_name")
	role := r.FormValue("role")

	var classID int
	if role == "student" {
		if classStr := r.FormValue("class_id"); classStr != "" {
			id, _ := strconv.Atoi(classStr)
			classID = id
		}
	}

	user, err := h.userRepo.Register(username, password, role, fullName, classID)
	if err != nil {
		classes, _ := h.userRepo.GetAllClasses()
		data := map[string]interface{}{
			"Title":   "Регистрация",
			"Error":   "Ошибка регистрации: " + err.Error(),
			"Classes": classes,
			"Form": map[string]string{
				"username":  username,
				"full_name": fullName,
				"role":      role,
			},
		}
		h.tmpl.Execute(w, data)
		return
	}

	sessionToken, err := h.userRepo.CreateSession(user.ID)
	if err != nil {
		http.Redirect(w, r, "/login?error=session_error", http.StatusSeeOther)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    sessionToken,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,                         
	})

	session, _ := store.Get(r, "app-session")
	session.Values["user_id"] = user.ID
	session.Values["username"] = user.Username
	session.Values["role"] = user.Role
	session.Save(r, w)

	switch user.Role {
	case "student":
		http.Redirect(w, r, "/home", http.StatusSeeOther)
	case "teacher":
		http.Redirect(w, r, "/teacher_home", http.StatusSeeOther)
	default:
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}
