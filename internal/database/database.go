package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB(connStr string) error {
	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("ошибка подключения к БД: %v", err)
	}

	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(25)
	DB.SetConnMaxLifetime(5 * time.Minute)

	err = DB.Ping()
	if err != nil {
		return fmt.Errorf("ошибка ping БД: %v", err)
	}

	log.Println("База данных подключена успешно")
	return nil
}

func CloseDB() {
	if DB != nil {
		DB.Close()
		log.Println("Соединение с БД закрыто")
	}
}
