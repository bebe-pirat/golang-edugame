package main

import (
	"edugame/internal/database"
	"edugame/internal/handler"
	middleware "edugame/internal/midlleware"
	"edugame/internal/repository"

	"encoding/gob"
	"fmt"
	"net/http"
	"os"
)

func init() {
	gob.Register(map[string]string{})
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
		port = "8800"
	}

	teacherRepo := repository.NewTeacherRepository(database.DB)
	typeRepo := repository.NewTypeRepository(database.DB)
	userRepo := repository.NewUserRepository(database.DB)
	userProgressRepo := repository.NewUserProgressRepository(database.DB)

	indexHandler := handler.NewIndexHandler()
	equationHandler := handler.NewEquationHandler(userRepo, typeRepo)
	statsHandler := handler.NewStatsHandler(userProgressRepo, userRepo)
	loginHandler := handler.NewLoginHandler(userRepo)
	registrationHandler := handler.NewRegistrationHandler(userRepo)
	homeHandler := handler.NewHomeHandler()
	teacherHandlers := handler.NewTeacherHandlers(teacherRepo)

	mux := http.NewServeMux()

	mux.Handle("/static/", http.StripPrefix("/static/",
		http.FileServer(http.Dir("../../internal/static"))))

	mux.HandleFunc("/", indexHandler.IndexHandler)
	mux.HandleFunc("/login", loginHandler.LoginPage)
	mux.HandleFunc("/auth/login", loginHandler.Login)
	mux.HandleFunc("/register", registrationHandler.RegisterPage)
	mux.HandleFunc("/auth/register", registrationHandler.Register)

	mux.Handle("/home",
		middleware.RequireRole("student")(http.HandlerFunc(homeHandler.HomePage)))

	mux.Handle("/equation",
		middleware.RequireRole("student")(http.HandlerFunc(equationHandler.EquationHandler)))

	mux.Handle("/stats",
		middleware.RequireRole("student")(http.HandlerFunc(statsHandler.StatsPage)))

	mux.Handle("/api/check",
		middleware.RequireRole("student")(http.HandlerFunc(equationHandler.CheckAnswersHandler)))

	mux.Handle("/teacher_home",
		middleware.RequireRole("teacher")(http.HandlerFunc(teacherHandlers.TeacherHome)))

	mux.Handle("/teacher/class", middleware.RequireRole("teacher")(http.HandlerFunc(teacherHandlers.ClassStatistics)))

	mux.Handle("/teacher/student",
		middleware.RequireRole("teacher")(http.HandlerFunc(teacherHandlers.StudentStatistics)))

	mux.Handle("/teacher/student/attempts",
		middleware.RequireRole("teacher")(http.HandlerFunc(teacherHandlers.StudentAttemptsByType)))

	mux.Handle("/logout",
		middleware.RequireAuth(http.HandlerFunc(loginHandler.Logout)))

	fmt.Printf("🚀 Сервер запущен на порту %s\n", port)
	fmt.Println("📌 Публичные маршруты: /, /login, /register")
	fmt.Println("🎓 Студентские маршруты: /home, /equation, /stats, /api/check")
	fmt.Println("👨‍🏫 Учительские маршруты: /teacher_home, /teacher/class, /teacher/student")

	err = http.ListenAndServe(":"+port, mux)
	if err != nil {
		fmt.Printf("❌ Ошибка запуска сервера: %v\n", err)
		os.Exit(1)
	}
}
