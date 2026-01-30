package main

import (
	"edugame/internal/database"
	"edugame/internal/generator"
	"edugame/internal/repository"
	"fmt"
	"html/template"
	"net/http"
	"os"
)

var tplIndex = template.Must(template.ParseFiles("../../internal/templates/index.html"))
var tplEquation = template.Must(template.ParseFiles("../../internal/templates/equation.html"))

const countEqs = 10

func main() {
	err := database.InitDB()
	if err != nil {
		fmt.Printf("Ошибка инициализации БД: %v\n", err)
		return
	}
	defer database.CloseDB()

	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/", indexHandler)
	mux.HandleFunc("/equation", equationHandler)

	fmt.Println("Сервер запустился, дура")
	err = http.ListenAndServe(":"+port, mux)
	if err != nil {
		fmt.Printf("Ошибка запуска сервера: %v\n", err)
		os.Exit(1)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tplIndex.Execute(w, nil)
}

type EquationData struct {
	Eqs   []generator.Equation
	Class int
}

func NewEquationData(list []generator.Equation, class int) *EquationData {
	return &EquationData{
		Eqs:   list,
		Class: class,
	}
}

func equationHandler(w http.ResponseWriter, r *http.Request) {
	typeRepo := repository.NewTypeRepository(database.DB)
	userRepo := repository.NewUserRepository(database.DB)

	userId, err := userRepo.GetTestUserId()
	if err != nil {
		fmt.Println(err)
		return
	}

	sessionId, err := userRepo.CreateSession(userId)
	if err != nil {
		fmt.Println(err, sessionId)
		return
	}

	class, err := userRepo.GetTestClassbyUserId(userId)
	if err != nil {
		fmt.Println(err, sessionId)
		return
	}
	listTypes, err := typeRepo.GetListTypes(class)
	if err != nil {
		fmt.Println(err, sessionId)
		return
	}

	fmt.Println("Успешно достаны типы уравнений для ", class, "класса")
	fmt.Println("Количество уравнений: ", len(listTypes))

    listEquations, err := generateListOfEquations(listTypes)
    if err != nil {
        fmt.Println("Ошибка генерации уравнений:", err)
        http.Error(w, "Ошибка генерации уравнений", http.StatusInternalServerError)
        return
    }
    
    fmt.Printf("Сгенерировано %d уравнений:\n", len(listEquations))
    for i, eq := range listEquations {
        fmt.Printf("  %d: %s (ответ: %s)\n", i+1, eq.Text, eq.CorrectAnswer)
    }

	equationData := NewEquationData(listEquations, listEquations[0].Class)

	tplEquation.Execute(w, equationData)
}

func generateListOfEquations(types []generator.EquationType) ([]generator.Equation, error) {
	eqs := make([]generator.Equation, countEqs)
	typesCount := len(types)
	gen := generator.NewGenerator()

	for i := 1; i <= countEqs; i++ {
		eq, err := gen.GenerateEquation(types[i%typesCount])

		if err != nil {
			return nil, err
		}

		eqs[i-1] = eq
	}

	return eqs, nil
}
