// handlers/teacher_handlers.go
package handler

import (
	"edugame/internal/repository"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
)

type TeacherHandlers struct {
	teacherRepo *repository.TeacherRepository
	tmpl        *template.Template
}

func NewTeacherHandlers(teacherRepo *repository.TeacherRepository) *TeacherHandlers {
	tmpl := template.Must(template.ParseFiles("../../internal/templates/class_statisctics.html", "../../internal/templates/student_statisctics.html", "../../internal/templates/teacher_home.html", "../../internal/templates/student_attempts.html"))

	return &TeacherHandlers{
		teacherRepo: teacherRepo,
		tmpl:        tmpl,
	}
}

func (h *TeacherHandlers) TeacherOnly(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Получаем сессию
		session, err := store.Get(r, "app-session")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Проверяем роль
		role, ok := session.Values["role"].(string)
		if !ok || role != "teacher" {
			http.Error(w, "Доступ запрещен", http.StatusForbidden)
			return
		}

		next(w, r)
	}
}

// Главная страница учителя
func (h *TeacherHandlers) TeacherHome(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "app-session")
	teacherID, ok := session.Values["user_id"].(int)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Получаем классы учителя
	classes, err := h.teacherRepo.GetTeacherClasses(teacherID)
	if err != nil {
		http.Error(w, "Ошибка получения классов", http.StatusInternalServerError)
		return
	}

	// Если у учителя есть классы, показываем статистику первого класса
	var classStats map[string]interface{}
	if len(classes) > 0 {
		classStats, _ = h.teacherRepo.GetClassStatistics(classes[0].ID)
	}

	data := map[string]interface{}{
		"Classes":    classes,
		"ClassStats": classStats,
	}

	h.tmpl.ExecuteTemplate(w, "teacher_home.html", data)
}

// Статистика класса
func (h *TeacherHandlers) ClassStatistics(w http.ResponseWriter, r *http.Request) {
	classIDStr := r.URL.Query().Get("class_id")
	classID, err := strconv.Atoi(classIDStr)
	if err != nil {
		http.Error(w, "Некорректный ID класса", http.StatusBadRequest)
		return
	}

	// Получаем статистику класса
	stats, err := h.teacherRepo.GetClassStatistics(classID)
	if err != nil {
		http.Error(w, "Ошибка получения статистики", http.StatusInternalServerError)
		fmt.Println(err)
		return
	}

	// Получаем учеников класса
	students, err := h.teacherRepo.GetClassStudents(classID)
	if err != nil {
		http.Error(w, "Ошибка получения учеников", http.StatusInternalServerError)
		fmt.Println(err)
		return
	}
	data := map[string]interface{}{
		"ClassID":  classID,
		"Stats":    stats,
		"Students": students,
	}

	err = h.tmpl.ExecuteTemplate(w, "class_statisctics.html", data)
	if err != nil {
		fmt.Println(err)
	}
}

// Статистика ученика
func (h *TeacherHandlers) StudentStatistics(w http.ResponseWriter, r *http.Request) {
	studentIDStr := r.URL.Query().Get("student_id")
	studentID, err := strconv.Atoi(studentIDStr)
	if err != nil {
		http.Error(w, "Некорректный ID ученика", http.StatusBadRequest)
		fmt.Println(err)
		return
	}

	// Получаем статистику ученика
	stats, err := h.teacherRepo.GetStudentStatistics(studentID)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Ошибка получения статистики", http.StatusInternalServerError)
		return
	}

	h.tmpl.ExecuteTemplate(w, "student_statisctics.html", stats)
}

// Попытки ученика по типу уравнения
func (h *TeacherHandlers) StudentAttemptsByType(w http.ResponseWriter, r *http.Request) {
	studentIDStr := r.URL.Query().Get("student_id")
	typeIDStr := r.URL.Query().Get("type_id")

	studentID, err := strconv.Atoi(studentIDStr)
	if err != nil {
		http.Error(w, "Некорректный ID ученика", http.StatusBadRequest)
		fmt.Println(err)
		return
	}

	fmt.Println(studentID)

	typeID, err := strconv.Atoi(typeIDStr)
	if err != nil {
		http.Error(w, "Некорректный ID типа", http.StatusBadRequest)
		fmt.Println(err)
		return
	}
	fmt.Println(typeID)

	// Получаем попытки
	attempts, err := h.teacherRepo.GetStudentAttemptsByType(studentID, typeID)
	if err != nil {
		http.Error(w, "Ошибка получения попыток", http.StatusInternalServerError)
		fmt.Println(err)
		return
	}
	fmt.Println(attempts)

	// Получаем информацию об ученике
	studentStats, _ := h.teacherRepo.GetStudentStatistics(studentID)
	fmt.Println(studentStats)

	data := map[string]interface{}{
		"StudentInfo": studentStats["student_info"],
		"Attempts":    attempts,
		"TypeID":      typeID,
	}

	err = h.tmpl.ExecuteTemplate(w, "student_attempts.html", data)
	if err != nil {
		fmt.Println(err)
	}
}
