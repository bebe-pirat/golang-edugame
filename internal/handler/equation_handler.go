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

	"github.com/gorilla/sessions"
)

var store = sessions.NewCookieStore([]byte("a-very-secret-key"))

type EquationHandler struct {
	tmpl     *template.Template
	userRepo *repository.UserRepository
	typeRepo *repository.TypeRepository
	gen      *generator.Generator
}

func NewEquationHandler(userRepo *repository.UserRepository, typeRepo *repository.TypeRepository) *EquationHandler {
	tmpl := template.Must(template.ParseFiles("../../internal/templates/equation.html"))

	return &EquationHandler{
		tmpl:     tmpl,
		userRepo: userRepo,
		typeRepo: typeRepo,
		gen:      generator.NewGenerator(),
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

    fmt.Printf("Пользователь: %s (ID: %d, Класс: %d)\n", user.Username, userId, class)
    fmt.Printf("Типы уравнений для %d класса: %d\n", class, len(listTypes))    

	listEquations, err := generateListOfEquations(listTypes)
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
