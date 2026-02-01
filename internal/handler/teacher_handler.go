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

// DirectorHome - главная страница директора
func (h *TeacherHandlers) DirectorHome(w http.ResponseWriter, r *http.Request) {
	fmt.Println("!")
	classes, err := h.teacherRepo.GetAllClasses()
	if err != nil {
		http.Error(w, "Ошибка получения классов", http.StatusInternalServerError)
		fmt.Printf("Ошибка GetAllClasses: %v\n", err)
		return
	}

	// 2. Получаем общую статистику по всем классам
	stats, err := h.teacherRepo.GetClassesStatistics()
	if err != nil {
		http.Error(w, "Ошибка получения статистики", http.StatusInternalServerError)
		fmt.Printf("Ошибка GetClassesStatistics: %v\n", err)
		return
	}

	// 3. Подготавливаем данные для шаблона
	data := map[string]interface{}{
		"Title":      "Панель директора",
		"Classes":    classes,
		"ClassStats": nil, // по умолчанию nil
	}

	// 4. Извлекаем общую статистику (класс с ID 0)
	if overallStats, exists := stats[0]; exists {
		data["ClassStats"] = overallStats

		// Отладочная информация
		fmt.Printf("Общая статистика получена: %d классов, %d учеников\n",
			overallStats["total_classes"],
			overallStats["total_students"])
	} else {
		fmt.Println("Общая статистика не найдена (ключ 0 отсутствует)")
		// Создаем пустую статистику
		data["ClassStats"] = map[string]interface{}{
			"total_classes":    len(classes),
			"total_students":   0,
			"total_attempts":   0,
			"total_correct":    0,
			"overall_accuracy": 0,
			"top_students":     []map[string]interface{}{},
		}
	}

	// 5. Рендерим шаблон
	fmt.Printf("Рендеринг шаблона с %d классами\n", len(classes))
	err = h.tmpl.ExecuteTemplate(w, "director_overall_stats.html", data)
	if err != nil {
		fmt.Printf("Ошибка рендеринга шаблона: %v\n", err)
		http.Error(w, "Ошибка отображения страницы", http.StatusInternalServerError)
		return
	}
}

func (h *TeacherHandlers) DirectorClassStats(w http.ResponseWriter, r *http.Request) {
	classIDStr := r.URL.Query().Get("class_id")
	classID, err := strconv.Atoi(classIDStr)
	if err != nil {
		http.Error(w, "Некорректный ID класса", http.StatusBadRequest)
		fmt.Printf("Некорректный class_id: %s\n", classIDStr)
		return
	}

	fmt.Printf("Запрос статистики класса ID: %d\n", classID)

	// Получаем статистику класса
	stats, err := h.teacherRepo.GetClassStatistics(classID)
	if err != nil {
		http.Error(w, "Ошибка получения статистики", http.StatusInternalServerError)
		fmt.Printf("Ошибка GetClassStatistics: %v\n", err)
		return
	}

	// Получаем учеников класса
	students, err := h.teacherRepo.GetClassStudents(classID)
	if err != nil {
		http.Error(w, "Ошибка получения учеников", http.StatusInternalServerError)
		fmt.Printf("Ошибка GetClassStudents: %v\n", err)
		return
	}

	fmt.Printf("Статистика класса %d: %+v\n", classID, stats)
	fmt.Printf("Учеников в классе: %d\n", len(students))

	data := map[string]interface{}{
		"Title":    "Статистика класса",
		"ClassID":  classID,
		"Stats":    stats,
		"Students": students,
	}

	err = h.tmpl.ExecuteTemplate(w, "director_class.html", data)
	if err != nil {
		fmt.Printf("Ошибка рендеринга шаблона: %v\n", err)
		http.Error(w, "Ошибка отображения страницы", http.StatusInternalServerError)
	}
}

// Статистика класса
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
		fmt.Println("Ошибка получения класса")
		return
	}

	// Получаем статистику класса
	stats, err := h.teacherRepo.GetClassStatistics(class.ID)
	if err != nil {
		http.Error(w, "Ошибка получения статистики", http.StatusInternalServerError)
		fmt.Println(err)
		return
	}

	// Получаем учеников класса
	students, err := h.teacherRepo.GetClassStudents(class.ID)
	if err != nil {
		http.Error(w, "Ошибка получения учеников", http.StatusInternalServerError)
		fmt.Println(err)
		return
	}
	data := map[string]interface{}{
		"ClassID":  class.ID,
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

// Статистика ученика
func (h *TeacherHandlers) DirectorStudentStatistics(w http.ResponseWriter, r *http.Request) {
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

	h.tmpl.ExecuteTemplate(w, "director_student.html", stats)
}

// Попытки ученика по типу уравнения
func (h *TeacherHandlers) DirectorStudentAttemptsByType(w http.ResponseWriter, r *http.Request) {
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

	err = h.tmpl.ExecuteTemplate(w, "director_student_attempts.html", data)
	if err != nil {
		fmt.Println(err)
	}
}
