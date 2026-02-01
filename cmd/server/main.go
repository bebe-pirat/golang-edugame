package main

import (
	"edugame/internal/database"
	"edugame/internal/handler"
	middleware "edugame/internal/midlleware"
	"edugame/internal/repository"
	"log"
	"time"

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
		port = "3000"
	}

	teacherRepo := repository.NewTeacherRepository(database.DB)
	typeRepo := repository.NewTypeRepository(database.DB)
	userRepo := repository.NewUserRepository(database.DB)
	userProgressRepo := repository.NewUserProgressRepository(database.DB)

	indexHandler := handler.NewIndexHandler()
	equationHandler := handler.NewEquationHandler(userRepo, typeRepo, userProgressRepo)
	statsHandler := handler.NewStatsHandler(userProgressRepo, userRepo)
	loginHandler := handler.NewLoginHandler(userRepo)
	registrationHandler := handler.NewRegistrationHandler(userRepo)
	homeHandler := handler.NewHomeHandler()
	teacherHandlers := handler.NewTeacherHandlers(teacherRepo)

	mux := http.NewServeMux()

	mux.Handle("/static/", http.StripPrefix("/static/",
		http.FileServer(http.Dir("internal/static"))))

	mux.HandleFunc("/", indexHandler.IndexHandler)
	mux.HandleFunc("/login", loginHandler.LoginPage)
	mux.HandleFunc("/auth/login", loginHandler.Login)
	mux.HandleFunc("/register", registrationHandler.RegisterPage)
	mux.HandleFunc("/auth/register", registrationHandler.Register)

	mux.Handle("/home",
		middleware.RequireRoles([]string{"student"})(http.HandlerFunc(homeHandler.HomePage)))

	mux.Handle("/equation",
		middleware.RequireRoles([]string{"student"})(http.HandlerFunc(equationHandler.EquationHandler)))

	mux.Handle("/stats",
		middleware.RequireRoles([]string{"student"})(http.HandlerFunc(statsHandler.StatsPage)))

	mux.Handle("/api/check",
		middleware.RequireRoles([]string{"student"})(http.HandlerFunc(equationHandler.CheckAnswersHandler)))

	mux.Handle("/director",
		middleware.RequireRoles([]string{"director"})(http.HandlerFunc(teacherHandlers.DirectorHome)))

	mux.Handle("/director/class",
		middleware.RequireRoles([]string{"director"})(http.HandlerFunc(teacherHandlers.DirectorClassStats)))

	mux.Handle("/director/student",
		middleware.RequireRoles([]string{"director"})(http.HandlerFunc(teacherHandlers.DirectorStudentStatistics)))

	mux.Handle("/director/student/attempts",
		middleware.RequireRoles([]string{"director"})(http.HandlerFunc(teacherHandlers.DirectorStudentAttemptsByType)))

	mux.Handle("/teacher/class", middleware.RequireRoles([]string{"teacher"})(http.HandlerFunc(teacherHandlers.ClassStatistics)))

	mux.Handle("/teacher/student",
		middleware.RequireRoles([]string{"teacher"})(http.HandlerFunc(teacherHandlers.StudentStatistics)))

	mux.Handle("/teacher/student/attempts",
		middleware.RequireRoles([]string{"teacher"})(http.HandlerFunc(teacherHandlers.StudentAttemptsByType)))

	mux.Handle("/logout",
		middleware.RequireAuth(http.HandlerFunc(loginHandler.Logout)))

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Printf("Сервер запущен на порту %s\n", port)
	log.Println("Публичные маршруты: /, /login, /register")
	log.Println("Студентские маршруты: /home, /equation, /stats, /api/check")
	log.Println("Учительские маршруты: /teacher_home, /teacher/class, /teacher/student")

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
