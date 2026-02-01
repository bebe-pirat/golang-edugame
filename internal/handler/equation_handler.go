package handler

import (
	"edugame/internal"
	"edugame/internal/database"
	"edugame/internal/entity"
	"edugame/internal/generator"
	"edugame/internal/repository"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"sort"

	"github.com/gorilla/sessions"
)

var store = sessions.NewCookieStore([]byte("a-very-secret-key"))

type EquationHandler struct {
	tmpl             *template.Template
	userRepo         *repository.UserRepository
	typeRepo         *repository.TypeRepository
	userProgressRepo *repository.UserProgressRepository
	gen              *generator.Generator
}

func NewEquationHandler(userRepo *repository.UserRepository, typeRepo *repository.TypeRepository, userProgressRepo *repository.UserProgressRepository) *EquationHandler {
	tmpl := template.Must(template.ParseFiles("../../internal/templates/equation.html"))

	return &EquationHandler{
		tmpl:             tmpl,
		userRepo:         userRepo,
		typeRepo:         typeRepo,
		userProgressRepo: userProgressRepo,
		gen:              generator.NewGenerator(),
	}
}

type EquationWithID struct {
	Id int
	Eq generator.Equation
}

func NewEquationWithID(eq generator.Equation, id int) *EquationWithID {
	return &EquationWithID{
		Id: id,
		Eq: eq,
	}
}

type EquationData struct {
	Eqs   []EquationWithID
	Class int
}

func NewEquationData(list []EquationWithID, class int) *EquationData {
	return &EquationData{
		Eqs:   list,
		Class: class,
	}
}

func (e *EquationHandler) EquationHandler(w http.ResponseWriter, r *http.Request) {
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

	user, err := e.userRepo.GetByID(userId)
	if err != nil {
		fmt.Println("Ошибка получения пользователя:", err)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if user.Role != "student" {
		http.Error(w, "Доступ запрещен. Только для учеников", http.StatusForbidden)
		return
	}

	class, err := e.userRepo.GetStudentClass(userId)
	if err != nil {
		fmt.Println("Ошибка получения класс: ", err)
		return
	}
	listTypes, err := e.typeRepo.GetListTypes(class)
	if err != nil {
		fmt.Println("Ошибка получения типов уравнений:", err)
		http.Error(w, "Ошибка загрузки уравнений", http.StatusInternalServerError)
		return
	}

	typeStats, err := e.userProgressRepo.GetUserTypeStatistics(userId)
	if err != nil {
		fmt.Println("Ошибка получения статистики:", err)
		// Продолжаем без адаптивной генерации
	}

	fmt.Printf("Пользователь: %s (ID: %d, Класс: %d)\n", user.Username, userId, class)
	fmt.Printf("Типы уравнений для %d класса: %d\n", class, len(listTypes))

	listEquations, err := e.generateAdaptiveEquations(listTypes, typeStats, userId)
	if err != nil {
		fmt.Println("Ошибка генерации уравнений:", err)
		http.Error(w, "Ошибка генерации уравнений", http.StatusInternalServerError)
		return
	}

	fmt.Printf("Сгенерировано %d уравнений:\n", len(listEquations))
	for i, eq := range listEquations {
		fmt.Printf("  %d: %s (ответ: %s)\n", i+1, eq.Eq.Text, eq.Eq.CorrectAnswer)
	}

	session, _ = store.Get(r, "equations-session")
	correctAnswers := make(map[int]string)
	for i, eq := range listEquations {
		correctAnswers[i] = eq.Eq.CorrectAnswer
	}
	session.Values["correct_answers"] = correctAnswers
	if err := session.Save(r, w); err != nil {
		fmt.Println("Ошибка сохранения верных ответов в сессию")
		fmt.Println("Error: ", err)
		return
	}

	equationData := NewEquationData(listEquations, listEquations[0].Eq.Class)

	e.tmpl.Execute(w, equationData)
}

// generateAdaptiveEquations - адаптивная генерация уравнений
func (e *EquationHandler) generateAdaptiveEquations(
	types []generator.EquationType,
	typeStats map[int]repository.TypeStat,
	userId int,
) ([]EquationWithID, error) {
	const totalEquations = internal.CountEqs

	// 1. Классифицируем типы уравнений по успеваемости
	var weakTypes []generator.EquationType   // < 70% правильных
	var mediumTypes []generator.EquationType // 70-90%
	var strongTypes []generator.EquationType // > 90%
	var newTypes []generator.EquationType    // нет статистики

	for _, t := range types {
		stat, exists := typeStats[t.ID]

		if !exists || stat.Attempts == 0 {
			// Нет статистики - новый тип
			newTypes = append(newTypes, t)
			continue
		}

		accuracy := float64(stat.Correct) / float64(stat.Attempts) * 100

		if accuracy < 70 {
			weakTypes = append(weakTypes, t)
		} else if accuracy < 90 {
			mediumTypes = append(mediumTypes, t)
		} else {
			strongTypes = append(strongTypes, t)
		}

		fmt.Printf("Тип %d: попыток=%d, верно=%d, точность=%.1f%%\n",
			t.ID, stat.Attempts, stat.Correct, accuracy)
	}

	// 2. Рассчитываем количество уравнений каждого типа
	//    Базовая формула: слабые × 2, средние × 1, сильные × 0.5, новые × 1.5

	// Веса для разных категорий
	weights := map[string]float64{
		"weak":   2.0, // В 2 раза чаще
		"medium": 1.0, // Стандартная частота
		"strong": 0.5, // В 2 раза реже
		"new":    1.5, // Чаще новых для освоения
	}

	// Рассчитываем общий вес
	totalWeight := 0.0
	totalWeight += float64(len(weakTypes)) * weights["weak"]
	totalWeight += float64(len(mediumTypes)) * weights["medium"]
	totalWeight += float64(len(strongTypes)) * weights["strong"]
	totalWeight += float64(len(newTypes)) * weights["new"]

	// Распределяем уравнения
	equations := make([]EquationWithID, 0, totalEquations)
	gen := generator.NewGenerator()

	// 3. Генерируем уравнения в соответствии с распределением
	equationIndex := 0

	// Список всех типов с их весами для случайного выбора
	type weightedType struct {
		Type   generator.EquationType
		Weight float64
		Count  int
	}

	weightedTypes := make([]weightedType, 0)

	// Добавляем типы с весами
	for _, t := range weakTypes {
		weightedTypes = append(weightedTypes, weightedType{
			Type:   t,
			Weight: weights["weak"],
			Count:  0,
		})
	}

	for _, t := range mediumTypes {
		weightedTypes = append(weightedTypes, weightedType{
			Type:   t,
			Weight: weights["medium"],
			Count:  0,
		})
	}

	for _, t := range strongTypes {
		weightedTypes = append(weightedTypes, weightedType{
			Type:   t,
			Weight: weights["strong"],
			Count:  0,
		})
	}

	for _, t := range newTypes {
		weightedTypes = append(weightedTypes, weightedType{
			Type:   t,
			Weight: weights["new"],
			Count:  0,
		})
	}

	// 4. Алгоритм выбора типа уравнения
	for equationIndex < totalEquations {
		// Сортируем по количеству уже сгенерированных уравнений
		sort.Slice(weightedTypes, func(i, j int) bool {
			// Сначала выбираем те, которые генерировались меньше чем ожидалось
			expectedI := float64(equationIndex+1) * weightedTypes[i].Weight / totalWeight
			expectedJ := float64(equationIndex+1) * weightedTypes[j].Weight / totalWeight
			ratioI := float64(weightedTypes[i].Count) / expectedI
			ratioJ := float64(weightedTypes[j].Count) / expectedJ
			return ratioI < ratioJ
		})

		// Выбираем тип из топ-3 кандидатов для разнообразия
		candidates := weightedTypes
		if len(weightedTypes) > 3 {
			candidates = weightedTypes[:3]
		}

		// Выбираем случайный тип из кандидатов
		selectedIdx := 0
		if len(candidates) > 1 {
			selectedIdx = gen.GetRandSource().Intn(len(candidates))
		}

		selectedType := candidates[selectedIdx].Type

		// Генерируем уравнение
		eq, err := gen.GenerateEquation(selectedType)
		if err != nil {
			return nil, err
		}

		equations = append(equations, *NewEquationWithID(eq, equationIndex))

		// Обновляем счетчик
		for i := range weightedTypes {
			if weightedTypes[i].Type.ID == selectedType.ID {
				weightedTypes[i].Count++
				break
			}
		}

		equationIndex++
	}

	// 5. Перемешиваем уравнения для разнообразия
	shuffledEquations := make([]EquationWithID, len(equations))
	perm := gen.GetRandSource().Perm(len(equations))
	for i, v := range perm {
		equations[v].Id = i
		shuffledEquations[i] = equations[v]
	}

	// 6. Логируем распределение
	fmt.Printf("\n=== АДАПТИВНАЯ ГЕНЕРАЦИЯ ===\n")
	fmt.Printf("Слабые типы (<70%%): %d\n", len(weakTypes))
	fmt.Printf("Средние типы (70-90%%): %d\n", len(mediumTypes))
	fmt.Printf("Сильные типы (>90%%): %d\n", len(strongTypes))
	fmt.Printf("Новые типы: %d\n", len(newTypes))
	fmt.Printf("Итоговое распределение:\n")

	for _, wt := range weightedTypes {
		percentage := float64(wt.Count) / float64(totalEquations) * 100
		stat, exists := typeStats[wt.Type.ID]

		if exists && stat.Attempts > 0 {
			accuracy := float64(stat.Correct) / float64(stat.Attempts) * 100
			fmt.Printf("  Тип %d: %d уравнений (%.1f%%) - точность: %.1f%%\n",
				wt.Type.ID, wt.Count, percentage, accuracy)
		} else {
			fmt.Printf("  Тип %d: %d уравнений (%.1f%%) - новый\n",
				wt.Type.ID, wt.Count, percentage)
		}
	}

	return shuffledEquations, nil
}

func generateListOfEquations(types []generator.EquationType) ([]EquationWithID, error) {
	eqs := make([]EquationWithID, internal.CountEqs)
	typesCount := len(types)
	gen := generator.NewGenerator()

	for i := 1; i <= internal.CountEqs; i++ {
		eq, err := gen.GenerateEquation(types[i%typesCount])

		if err != nil {
			return nil, err
		}

		eqs[i-1] = *NewEquationWithID(eq, i-1)
	}

	return eqs, nil
}

func (e *EquationHandler) CheckAnswersHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "equations-session")
	correctAnswers, ok := session.Values["correct_answers"].(map[int]string)

	if !ok {
		http.Error(w, "Сессия не найдена", http.StatusBadRequest)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Answers []struct {
			EquationID     int    `json:"equation_id"`
			UserAnswer     string `json:"user_answer"`
			EquationText   string `json:"equation_text"`
			EquationTypeId int    `json:"equation_type_id"`
		} `json:"answers"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	results := make([]map[string]interface{}, len(request.Answers))
	correctCount := 0
	attempts := make([]entity.Attempt, 0)
	userId, err := getUserIdFromSession(r)

	if err != nil {
		fmt.Println("ошибка получения id")
	}

	for i, answer := range request.Answers {
		correctAnswer, exists := correctAnswers[answer.EquationID]
		isCorrect := exists && answer.UserAnswer == correctAnswer

		feedback := "❌ Неправильно. Правильный ответ:" + correctAnswer

		if isCorrect {
			correctCount++
			feedback = "✅ Правильно!"
		}

		results[i] = map[string]interface{}{
			"equation_id":    answer.EquationID,
			"is_correct":     isCorrect,
			"correct_answer": correctAnswer,
			"feedback":       feedback,
		}

		fmt.Println(answer.EquationTypeId, answer.EquationText)
		attempts = append(attempts, entity.NewAttempt(userId, answer.EquationTypeId, answer.EquationText, correctAnswer, answer.UserAnswer))
	}

	go func() {
		attemptRepo := repository.NewAttemptRepository(database.DB)

		for _, a := range attempts {
			err := attemptRepo.SaveAttempt(a)
			if err != nil {
				fmt.Println("Error:", err)
				break
			}
		}
	}()

	response := map[string]interface{}{
		"total":            len(request.Answers),
		"correct":          correctCount,
		"results":          results,
		"overall_feedback": fmt.Sprintf("Правильно %d из %d", correctCount, len(request.Answers)),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func getUserIdFromSession(r *http.Request) (int, error) {
	session, err := store.Get(r, "app-session")
	if err != nil {
		return 0, err
	}

	userId, ok := session.Values["user_id"].(int)
	if !ok {
		return 0, fmt.Errorf("user_id not found in session")
	}

	return userId, nil
}
