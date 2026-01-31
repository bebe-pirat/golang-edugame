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
	loginHandler := handler.NewLoginHandler(userRepo)
	registrationHandler := handler.NewRegistrationHandler(userRepo)
	homeHandler := handler.NewHomeHandler()

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("../../internal/static"))))
	mux.HandleFunc("/", indexHandler.IndexHandler)
	mux.HandleFunc("/login", loginHandler.LoginPage)  // GET
	mux.HandleFunc("/auth/login", loginHandler.Login) // POST
	mux.HandleFunc("/logout", loginHandler.Logout)    // GET
	mux.HandleFunc("/register", registrationHandler.RegisterPage)
	mux.HandleFunc("/home", homeHandler.HomePage)                  // GET запрос
	mux.HandleFunc("/auth/register", registrationHandler.Register) // POST запрос
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


// сделать protected