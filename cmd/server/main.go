package main

import (
	"edugame/internal/database"
	"edugame/internal/handler"
	"edugame/internal/repository"

	"encoding/gob"
	"fmt"
	"net/http"
	"os"
)

func init() {
	gob.Register(map[int]string{})
}

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

	typeRepo := repository.NewTypeRepository(database.DB)
	userRepo := repository.NewUserRepository(database.DB)
	userProgressRepo := repository.NewUserProgressRepository(database.DB)

	indexHandler := handler.NewIndexHandler()
	equationHandler := handler.NewEquationHandler(userRepo, typeRepo)
	statsHandler := handler.NewStatsHandler(userProgressRepo, userRepo)

	mux.HandleFunc("/", indexHandler.IndexHandler)
	mux.HandleFunc("/equation", equationHandler.EquationHandler)
	mux.HandleFunc("/stats", statsHandler.StatsPage)
	mux.HandleFunc("/api/check", equationHandler.CheckAnswersHandler)

	fmt.Println("Сервер запустился, дура")
	err = http.ListenAndServe(":"+port, mux)
	if err != nil {
		fmt.Printf("Ошибка запуска сервера: %v\n", err)
		os.Exit(1)
	}
}
