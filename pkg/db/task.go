package db

import (
	"database/sql"
	"fmt"
	"time"
)

// Структура задачи для API
type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

// Добавляет задачу в БД и возвращает IDновой записи и ошибку
func AddTask(task *Task) (int64, error) {
	query := `INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`
	res, err := db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// Возвращает список задач с лимитом и поиском
func Tasks(limit int, search string) ([]*Task, error) {
	var rows *sql.Rows
	var err error

	if search == "" {
		// Без поиска — все задачи
		rows, err = db.Query(
			"SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date LIMIT ?",
			limit,
		)
	} else {
		// Проверяем, является ли search датой в формате 02.01.2006
		t, perr := time.Parse("02.01.2006", search)
		if perr == nil {
			// Это дата — ищем точное совпадение
			rows, err = db.Query(
				"SELECT id, date, title, comment, repeat FROM scheduler WHERE date = ? ORDER BY date LIMIT ?",
				t.Format("20060102"), limit,
			)
		} else {
			// Текстовый поиск по title и comment
			rows, err = db.Query(
				"SELECT id, date, title, comment, repeat FROM scheduler WHERE title LIKE ? OR comment LIKE ? ORDER BY date LIMIT ?",
				"%"+search+"%", "%"+search+"%", limit,
			)
		}
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		var id int
		var t Task
		err = rows.Scan(&id, &t.Date, &t.Title, &t.Comment, &t.Repeat)
		if err != nil {
			return nil, err
		}
		t.ID = fmt.Sprintf("%d", id)
		tasks = append(tasks, &t)
	}

	if tasks == nil {
		tasks = []*Task{}
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return tasks, nil
}

// Получает задачу по ID. Если ID не число или задачи нет — возвращает ошибку
func GetTask(id string) (*Task, error) {
	var task Task
	var intID int
	_, err := fmt.Sscanf(id, "%d", &intID)
	if err != nil {
		return nil, fmt.Errorf("неверный формат id")
	}

	query := `SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?`
	err = db.QueryRow(query, intID).Scan(&intID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("Задача не найдена")
		}
		return nil, err
	}
	task.ID = fmt.Sprintf("%d", intID)
	return &task, nil
}

// Обновляет задачу полностью. Если задачи нет — вернёт ошибку
func UpdateTask(task *Task) error {
	var intID int
	_, err := fmt.Sscanf(task.ID, "%d", &intID)
	if err != nil {
		return fmt.Errorf("неверный формат id")
	}

	query := `UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?`
	res, err := db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat, intID)
	if err != nil {
		return err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("Задача не найдена")
	}
	return nil
}

// DeleteTask удаляет задачу по id
func DeleteTask(id string) error {
	var intID int
	_, err := fmt.Sscanf(id, "%d", &intID)
	if err != nil {
		return fmt.Errorf("неверный формат id")
	}

	query := `DELETE FROM scheduler WHERE id = ?`
	res, err := db.Exec(query, intID)
	if err != nil {
		return err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("Задача не найдена")
	}
	return nil
}

// UpdateDate обновляет только дату у задачи
func UpdateDate(next string, id string) error {
	var intID int
	_, err := fmt.Sscanf(id, "%d", &intID)
	if err != nil {
		return fmt.Errorf("неверный формат id")
	}

	query := `UPDATE scheduler SET date = ? WHERE id = ?`
	res, err := db.Exec(query, next, intID)
	if err != nil {
		return err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("Задача не найдена")
	}
	return nil
}
