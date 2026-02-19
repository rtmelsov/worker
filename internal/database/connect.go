package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq" // Импорт драйвера PostgreSQL
	"github.com/sirupsen/logrus"

	"worker/internal/config"
)

// InitDB инициализирует подключение к базе данных
func InitDB(logEntry *logrus.Entry, cfg config.DatabaseConfig) (*sql.DB, error) {
	logEntry.Info("Connecting to database...")

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	// Проверяем соединение (ping)
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error connecting to the database: %w", err)
	}

	// Настройка пула
	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(100)

	logEntry.Info("Database connection established successfully")

	return db, nil
}

