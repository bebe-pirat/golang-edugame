// handlers/teacher_handlers.go
package handler

import (
	"edugame/internal/repository"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

type TeacherHandlers struct {
	teacherRepo *repository.TeacherRepository
	tmpl        *template.Template
}

func NewTeacherHandlers(teacherRepo *repository.TeacherRepository) *TeacherHandlers {
	tmpl := template.Must(template.ParseFiles(
		"internal/templates/class_statisctics.html",
		"internal/templates/student_statisctics.html",
		"internal/templates/director_overall_stats.html",
		"internal/templates/student_attempts.html",
		"internal/templates/director_student.html",
		"internal/templates/director_student_attempts.html",
		"internal/templates/director_class.html"))

	return &TeacherHandlers{
		teacherRepo: teacherRepo,
		tmpl:        tmpl,
	}
}

func (h *TeacherHandlers) TeacherOnly(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := store.Get(r, "app-session")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		role, ok := session.Values["role"].(string)
		if !ok || role != "teacher" {
			http.Error(w, "Доступ запрещен", http.StatusForbidden)
			return
		}

		next(w, r)
	}
}

func (h *TeacherHandlers) DirectorHome(w http.ResponseWriter, r *http.Request) {
	log.Println("!")
	classes, err := h.teacherRepo.GetAllClasses()
	if err != nil {
		http.Error(w, "Ошибка получения классов", http.StatusInternalServerError)
		log.Printf("Ошибка GetAllClasses: %v\n", err)
		return
	}

	stats, err := h.teacherRepo.GetClassesStatistics()
	if err != nil {
		http.Error(w, "Ошибка получения статистики", http.StatusInternalServerError)
		log.Printf("Ошибка GetClassesStatistics: %v\n", err)
		return
	}

	data := map[string]interface{}{
		"Title":      "Панель директора",
		"Classes":    classes,
		"ClassStats": nil, // по умолчанию nil
	}

	if overallStats, exists := stats[0]; exists {
		data["ClassStats"] = overallStats

		log.Printf("Общая статистика получена: %d классов, %d учеников\n",
			overallStats["total_classes"],
			overallStats["total_students"])
	} else {
		log.Println("Общая статистика не найдена (ключ 0 отсутствует)")
		data["ClassStats"] = map[string]interface{}{
			"total_classes":    len(classes),
			"total_students":   0,
			"total_attempts":   0,
			"total_correct":    0,
			"overall_accuracy": 0,
			"top_students":     []map[string]interface{}{},
		}
	}

	log.Printf("Рендеринг шаблона с %d классами\n", len(classes))
	err = h.tmpl.ExecuteTemplate(w, "director_overall_stats.html", data)
	if err != nil {
		log.Printf("Ошибка рендеринга шаблона: %v\n", err)
		http.Error(w, "Ошибка отображения страницы", http.StatusInternalServerError)
		return
	}
}

func (h *TeacherHandlers) DirectorClassStats(w http.ResponseWriter, r *http.Request) {
	classIDStr := r.URL.Query().Get("class_id")
	classID, err := strconv.Atoi(classIDStr)
	if err != nil {
		http.Error(w, "Некорректный ID класса", http.StatusBadRequest)
		log.Printf("Некорректный class_id: %s\n", classIDStr)
		return
	}

	log.Printf("Запрос статистики класса ID: %d\n", classID)

	stats, err := h.teacherRepo.GetClassStatistics(classID)
	if err != nil {
		http.Error(w, "Ошибка получения статистики", http.StatusInternalServerError)
		log.Printf("Ошибка GetClassStatistics: %v\n", err)
		return
	}

	students, err := h.teacherRepo.GetClassStudents(classID)
	if err != nil {
		http.Error(w, "Ошибка получения учеников", http.StatusInternalServerError)
		log.Printf("Ошибка GetClassStudents: %v\n", err)
		return
	}

	log.Printf("Статистика класса %d: %+v\n", classID, stats)
	log.Printf("Учеников в классе: %d\n", len(students))

	data := map[string]interface{}{
		"Title":    "Статистика класса",
		"ClassID":  classID,
		"Stats":    stats,
		"Students": students,
	}

	err = h.tmpl.ExecuteTemplate(w, "director_class.html", data)
	if err != nil {
		log.Printf("Ошибка рендеринга шаблона: %v\n", err)
		http.Error(w, "Ошибка отображения страницы", http.StatusInternalServerError)
	}
}

func (h *TeacherHandlers) ClassStatistics(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "app-session")
	teacherID, ok := session.Values["user_id"].(int)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	class, err := h.teacherRepo.GetTeacherClass(teacherID)
	if err != nil {
		http.Error(w, "Ошибка получения класса", http.StatusInternalServerError)
		log.Println("Ошибка получения класса")
		return
	}

	stats, err := h.teacherRepo.GetClassStatistics(class.ID)
	if err != nil {
		http.Error(w, "Ошибка получения статистики", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	students, err := h.teacherRepo.GetClassStudents(class.ID)
	if err != nil {
		http.Error(w, "Ошибка получения учеников", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	dailyResults, err := h.teacherRepo.GetDailyClassResults(class.ID, 0)
	if err != nil {
		log.Printf("Ошибка получения статистики недели:", err)
		return
	}

	data := map[string]interface{}{
		"ClassID":      class.ID,
		"Stats":        stats,
		"Students":     students,
		"DailyResults": dailyResults,
	}

	err = h.tmpl.ExecuteTemplate(w, "class_statisctics.html", data)
	if err != nil {
		log.Println(err)
	}
}

func (h *TeacherHandlers) StudentStatistics(w http.ResponseWriter, r *http.Request) {
	studentIDStr := r.URL.Query().Get("student_id")
	studentID, err := strconv.Atoi(studentIDStr)
	if err != nil {
		http.Error(w, "Некорректный ID ученика", http.StatusBadRequest)
		log.Println(err)
		return
	}

	stats, err := h.teacherRepo.GetStudentStatistics(studentID)
	if err != nil {
		log.Println(err)
		http.Error(w, "Ошибка получения статистики", http.StatusInternalServerError)
		return
	}

	h.tmpl.ExecuteTemplate(w, "student_statisctics.html", stats)
}

func (h *TeacherHandlers) StudentAttemptsByType(w http.ResponseWriter, r *http.Request) {
	studentIDStr := r.URL.Query().Get("student_id")
	typeIDStr := r.URL.Query().Get("type_id")

	studentID, err := strconv.Atoi(studentIDStr)
	if err != nil {
		http.Error(w, "Некорректный ID ученика", http.StatusBadRequest)
		log.Println(err)
		return
	}

	log.Println(studentID)

	typeID, err := strconv.Atoi(typeIDStr)
	if err != nil {
		http.Error(w, "Некорректный ID типа", http.StatusBadRequest)
		log.Println(err)
		return
	}
	log.Println(typeID)

	attempts, err := h.teacherRepo.GetStudentAttemptsByType(studentID, typeID)
	if err != nil {
		http.Error(w, "Ошибка получения попыток", http.StatusInternalServerError)
		log.Println(err)
		return
	}
	log.Println(attempts)

	studentStats, _ := h.teacherRepo.GetStudentStatistics(studentID)
	log.Println(studentStats)

	data := map[string]interface{}{
		"StudentInfo": studentStats["student_info"],
		"Attempts":    attempts,
		"TypeID":      typeID,
	}

	err = h.tmpl.ExecuteTemplate(w, "student_attempts.html", data)
	if err != nil {
		log.Println(err)
	}
}

func (h *TeacherHandlers) DirectorStudentStatistics(w http.ResponseWriter, r *http.Request) {
	studentIDStr := r.URL.Query().Get("student_id")
	studentID, err := strconv.Atoi(studentIDStr)
	if err != nil {
		http.Error(w, "Некорректный ID ученика", http.StatusBadRequest)
		log.Println(err)
		return
	}

	stats, err := h.teacherRepo.GetStudentStatistics(studentID)
	if err != nil {
		log.Println(err)
		http.Error(w, "Ошибка получения статистики", http.StatusInternalServerError)
		return
	}

	h.tmpl.ExecuteTemplate(w, "director_student.html", stats)
}

func (h *TeacherHandlers) DirectorStudentAttemptsByType(w http.ResponseWriter, r *http.Request) {
	studentIDStr := r.URL.Query().Get("student_id")
	typeIDStr := r.URL.Query().Get("type_id")

	studentID, err := strconv.Atoi(studentIDStr)
	if err != nil {
		http.Error(w, "Некорректный ID ученика", http.StatusBadRequest)
		log.Println(err)
		return
	}

	log.Println(studentID)

	typeID, err := strconv.Atoi(typeIDStr)
	if err != nil {
		http.Error(w, "Некорректный ID типа", http.StatusBadRequest)
		log.Println(err)
		return
	}
	log.Println(typeID)

	attempts, err := h.teacherRepo.GetStudentAttemptsByType(studentID, typeID)
	if err != nil {
		http.Error(w, "Ошибка получения попыток", http.StatusInternalServerError)
		log.Println(err)
		return
	}
	log.Println(attempts)

	studentStats, _ := h.teacherRepo.GetStudentStatistics(studentID)
	log.Println(studentStats)

	data := map[string]interface{}{
		"StudentInfo": studentStats["student_info"],
		"Attempts":    attempts,
		"TypeID":      typeID,
	}

	err = h.tmpl.ExecuteTemplate(w, "director_student_attempts.html", data)
	if err != nil {
		log.Println(err)
	}
}
