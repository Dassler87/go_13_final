package db

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// Глобальная переменная для подключения к БД
var db *sql.DB

const DateLayout = "20060102"

// Схема таблицы scheduler - задачи с датой, заголовком, комментарием и правилом повтора
const schema = `
CREATE TABLE IF NOT EXISTS scheduler (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date CHAR(8) NOT NULL DEFAULT "",
    title VARCHAR(255) NOT NULL DEFAULT "",
    comment TEXT,
    repeat VARCHAR(128) NOT NULL DEFAULT ""
);
CREATE INDEX IF NOT EXISTS idx_date ON scheduler(date);
`

// Init подключает SQL-базу по пути dbFile
// Если файла нет — создаёт его и сразу делает таблицу + индекс
func Init(dbFile string) error {
	var err error
	db, err = sql.Open("sqlite", dbFile)
	if err != nil {
		return fmt.Errorf("ошибка открытия БД: %w", err)
	}

	// Обязательно проверяем реальное подключение
	if err = db.Ping(); err != nil {
		_ = db.Close()
		return fmt.Errorf("ошибка подключения к БД: %w", err)
	}

	// Создаём таблицу и индекс, если их нет
	_, execErr := db.Exec(schema)
	if execErr != nil {
		_ = db.Close()
		return fmt.Errorf("ошибка создания схемы БД: %w", execErr)
	}

	return nil
}

// Close закрывает соединение с БД.
func Close() error {
	if db != nil {
		return db.Close()
	}
	return nil
}
