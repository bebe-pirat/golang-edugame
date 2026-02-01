package handler

import (
	"edugame/internal/entity"
	"edugame/internal/repository"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
)

type StatsHandler struct {
	tmpl             *template.Template
	userProgressRepo *repository.UserProgressRepository
	userRepo         *repository.UserRepository
}

func NewStatsHandler(up *repository.UserProgressRepository, u *repository.UserRepository) *StatsHandler {
	funcMap := template.FuncMap{
		// Функция для расчета процентов
		"percent": func(correct, total int) int {
			if total == 0 {
				return 0
			}
			return int(float64(correct) / float64(total) * 100)
		},
		// Функция для форматирования даты
		"formatDate": func(dateStr string) string {
			if dateStr == "" {
				return "—"
			}
			// Простое форматирование
			return dateStr[:10] // Берем только дату
		},
		// Функция для проверки, больше ли нуля
		"gt": func(a, b int) bool {
			return a > b
		},
	}
	tmpl := template.Must(template.New("stats.html").
		Funcs(funcMap).
		ParseFiles(
			"internal/templates/stats.html",
		))
	return &StatsHandler{
		tmpl:             tmpl,
		userProgressRepo: up,
		userRepo:         u,
	}
}

func (h *StatsHandler) StatsPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	session, err := store.Get(r, "app-session")
	if err != nil {
		fmt.Println("Ошибка получения сессии:", err)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	userId, ok := session.Values["user_id"].(int)
	if !ok || userId == 0 {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	stats, err := h.userProgressRepo.GetUserAllProgress(userId)
	if err != nil {
		fmt.Println("Error: ", err)
		http.Error(w, "Ошибка получения статистики", http.StatusInternalServerError)
		return
	}

	total, correct := h.GetTotalAndCorrectCount(stats)
	fmt.Println("Количество типов для пользователя: ", len(stats))

	data := map[string]interface{}{
		"Title":        "Статистика",
		"Stats":        stats,
		"TotalCount":   total,
		"CorrectCount": correct,
		"Accuracy":     float64(total) / float64(correct),
		"UserID":       userId,
		"UserName":     stats[0].Username,
	}

	h.tmpl.Execute(w, data)
}

func (h *StatsHandler) GetTotalAndCorrectCount(stats []entity.UserProgress) (int, int) {
	var total, correct int = 0, 0
	for _, value := range stats {
		total += value.AttemptsCount
		correct += value.CorrectCount
	}

	return total, correct
}

// GetStats - API endpoint для получения статистики (JSON)
func (h *StatsHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// Получаем user_id из запроса
	userIDStr := r.URL.Query().Get("user_id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil || userID == 0 {
		http.Error(w, "Неверный user_id", http.StatusBadRequest)
		return
	}

	// Получаем статистику
	stats, err := h.userProgressRepo.GetUserAllProgress(userID)
	if err != nil {
		http.Error(w, "Ошибка получения статистики", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"stats":   stats,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
