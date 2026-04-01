package main

import (
	"context"
	"edugame/internal/database"
	"edugame/internal/handler"
	middleware "edugame/internal/midlleware"
	"edugame/internal/repository"
	"edugame/internal/session"
	"errors"
	"log"
	"log/slog"
	"os/signal"
	"syscall"
	"time"

	"encoding/gob"
	"fmt"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func init() {
	gob.Register(map[string]string{})
	gob.Register(map[int]string{})
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	err := godotenv.Load()
	if err != nil {
		logger.Info("unable to load .env")
	}

	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		slog.Error("failed to load connection string")
		return
	}

	err = database.InitDB(connStr)
	if err != nil {
		fmt.Printf("Ошибка инициализации БД: %v\n", err)
		return
	}
	defer database.CloseDB()

	secretKey := os.Getenv("SESSION_SECRET_KEY")
	if secretKey == "" {
		log.Fatal("SESSION_SECRET_KEY is required")
	}

	session.InitStore(secretKey)
	store := session.GetStore()

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	teacherRepo := repository.NewTeacherRepository(database.DB)
	typeRepo := repository.NewTypeRepository(database.DB)
	userRepo := repository.NewUserRepository(database.DB)
	userProgressRepo := repository.NewUserProgressRepository(database.DB)
	schoolRepo := repository.NewSchoolRepository(database.DB)
	classRepo := repository.NewClassRepository(database.DB)
	roleRepo := repository.NewRoleRepository(database.DB)

	indexHandler := handler.NewIndexHandler()
	equationHandler := handler.NewEquationHandler(userRepo, typeRepo, userProgressRepo, store)
	statsHandler := handler.NewStatsHandler(userProgressRepo, userRepo, store)
	loginHandler := handler.NewLoginHandler(userRepo, store)
	registrationHandler := handler.NewRegistrationHandler(userRepo, store)
	homeHandler := handler.NewHomeHandler()
	teacherHandlers := handler.NewTeacherHandlers(teacherRepo, store)
	adminHandler := handler.NewAdminHandler(schoolRepo, classRepo, userRepo, roleRepo, typeRepo)

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

	// Админ-панель маршруты
	mux.Handle("/admin",
		middleware.RequireRoles([]string{"admin"})(http.HandlerFunc(adminHandler.Dashboard)))
	mux.Handle("/admin/dashboard",
		middleware.RequireRoles([]string{"admin"})(http.HandlerFunc(adminHandler.Dashboard)))

	// Школы
	mux.Handle("/admin/schools",
		middleware.RequireRoles([]string{"admin"})(http.HandlerFunc(adminHandler.Schools)))
	mux.Handle("/admin/schools/new",
		middleware.RequireRoles([]string{"admin"})(http.HandlerFunc(adminHandler.SchoolForm)))
	mux.Handle("/admin/schools/edit",
		middleware.RequireRoles([]string{"admin"})(http.HandlerFunc(adminHandler.SchoolForm)))
	mux.Handle("/admin/schools/create",
		middleware.RequireRoles([]string{"admin"})(http.HandlerFunc(adminHandler.SchoolCreate)))
	mux.Handle("/admin/schools/update",
		middleware.RequireRoles([]string{"admin"})(http.HandlerFunc(adminHandler.SchoolUpdate)))
	mux.Handle("/admin/schools/delete",
		middleware.RequireRoles([]string{"admin"})(http.HandlerFunc(adminHandler.SchoolDelete)))

	// Классы
	mux.Handle("/admin/classes",
		middleware.RequireRoles([]string{"admin"})(http.HandlerFunc(adminHandler.Classes)))
	mux.Handle("/admin/classes/new",
		middleware.RequireRoles([]string{"admin"})(http.HandlerFunc(adminHandler.ClassForm)))
	mux.Handle("/admin/classes/edit",
		middleware.RequireRoles([]string{"admin"})(http.HandlerFunc(adminHandler.ClassForm)))
	mux.Handle("/admin/classes/create",
		middleware.RequireRoles([]string{"admin"})(http.HandlerFunc(adminHandler.ClassCreate)))
	mux.Handle("/admin/classes/update",
		middleware.RequireRoles([]string{"admin"})(http.HandlerFunc(adminHandler.ClassUpdate)))
	mux.Handle("/admin/classes/delete",
		middleware.RequireRoles([]string{"admin"})(http.HandlerFunc(adminHandler.ClassDelete)))

	// Пользователи
	mux.Handle("/admin/users",
		middleware.RequireRoles([]string{"admin"})(http.HandlerFunc(adminHandler.Users)))
	mux.Handle("/admin/users/new",
		middleware.RequireRoles([]string{"admin"})(http.HandlerFunc(adminHandler.UserForm)))
	mux.Handle("/admin/users/edit",
		middleware.RequireRoles([]string{"admin"})(http.HandlerFunc(adminHandler.UserForm)))
	mux.Handle("/admin/users/create",
		middleware.RequireRoles([]string{"admin"})(http.HandlerFunc(adminHandler.UserCreate)))
	mux.Handle("/admin/users/update",
		middleware.RequireRoles([]string{"admin"})(http.HandlerFunc(adminHandler.UserUpdate)))
	mux.Handle("/admin/users/delete",
		middleware.RequireRoles([]string{"admin"})(http.HandlerFunc(adminHandler.UserDelete)))

	// Типы уравнений
	mux.Handle("/admin/equation-types",
		middleware.RequireRoles([]string{"admin"})(http.HandlerFunc(adminHandler.EquationTypes)))
	mux.Handle("/admin/equation-types/new",
		middleware.RequireRoles([]string{"admin"})(http.HandlerFunc(adminHandler.EquationTypeForm)))
	mux.Handle("/admin/equation-types/edit",
		middleware.RequireRoles([]string{"admin"})(http.HandlerFunc(adminHandler.EquationTypeForm)))
	mux.Handle("/admin/equation-types/create",
		middleware.RequireRoles([]string{"admin"})(http.HandlerFunc(adminHandler.EquationTypeCreate)))
	mux.Handle("/admin/equation-types/update",
		middleware.RequireRoles([]string{"admin"})(http.HandlerFunc(adminHandler.EquationTypeUpdate)))
	mux.Handle("/admin/equation-types/delete",
		middleware.RequireRoles([]string{"admin"})(http.HandlerFunc(adminHandler.EquationTypeDelete)))

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		slog.Info("server started", "port", port)

		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server didn't started", "error", err)
			return
		}
	}()

	signalChan := make(chan os.Signal, 1)

	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	slog.Info("server is shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("server forced to shutdown", "error", err)
	}

	slog.Info("server exiting")
}
