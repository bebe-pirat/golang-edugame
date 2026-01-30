package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB

type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func LoadConfig() Config {
	// Для локальной разработки - значения по умолчанию
	// TODO: надо будет разобраться, что делать когда сайт начнет работать на хостинге
	return Config{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "1234"),
		DBName:   getEnv("DB_NAME", "edugame"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func InitDB() error {
	config := LoadConfig()

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode,
	)

	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("ошибка подключения к БД: %w", err)
	}

	// Проверяем подключение
	err = DB.Ping()
	if err != nil {
		return fmt.Errorf("ошибка ping БД: %w", err)
	}

	// Устанавливаем максимальное количество соединений
	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(25)
	DB.SetConnMaxLifetime(5 * 60) // 5 минут

	log.Println("База данных подключена успешно")
	return nil
}

func CloseDB() {
	if DB != nil {
		DB.Close()
		log.Println("Соединение с БД закрыто")
	}
}

// все верно
