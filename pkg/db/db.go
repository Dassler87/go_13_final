package db

import (
	"database/sql"
	"fmt"
	"os"

	_ "modernc.org/sqlite"
)

// Глобальная переменная для подключения к БД
var db *sql.DB

// Схема таблицы scheduler - задачи с датой, заголовком, комментарием и правилом повтора
const schema = `
CREATE TABLE scheduler (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date CHAR(8) NOT NULL DEFAULT "",
    title VARCHAR(255) NOT NULL DEFAULT "",
    comment TEXT,
    repeat VARCHAR(128) NOT NULL DEFAULT ""
);
CREATE INDEX idx_date ON scheduler(date);
`

// Init подключает SQL-базу по пути dbFile
// Если файла нет — создаёт его и сразу делает таблицу + индекс
func Init(dbFile string) error {
	// Проверяем, существует ли файл
	_, err := os.Stat(dbFile)
	install := false
	if err != nil {
		install = true
	}

	// Открываем (или создаём) базу
	db, err = sql.Open("sqlite", dbFile)
	if err != nil {
		return fmt.Errorf("ошибка открытия БД: %w", err)
	}

	// Если файла не было — создаём таблицу и индекс
	if install {
		_, err = db.Exec(schema)
		if err != nil {
			return fmt.Errorf("ошибка создания таблицы и индекса: %w", err)
		}
	}

	return nil
}
