package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Dassler87/go_13_final/pkg/db"
)

// Создаём задачу
func addTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task db.Task
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		writeJson(w, map[string]string{"error": "ошибка десериализации JSON"})
		return
	}

	// Если заголовка нет - ошибка
	if task.Title == "" {
		writeJson(w, map[string]string{"error": "Не указан заголовок задачи"})
		return
	}

	now := time.Now()

	// Если дата не указана — берём сегодняшнюю
	if task.Date == "" {
		task.Date = now.Format(DateLayout)
	}

	// Проверяем формат даты
	t, err := time.Parse(DateLayout, task.Date)
	if err != nil {
		writeJson(w, map[string]string{"error": "дата представлена в неверном формате"})
		return
	}

	// Если правило повторения указано — проверяем и вычисляем следующую дату
	var next string
	if task.Repeat != "" {
		next, err = NextDate(now, task.Date, task.Repeat)
		if err != nil {
			writeJson(w, map[string]string{"error": err.Error()})
			return
		}
	}

	// Если дата меньше сегодняшней — корректируем
	if afterNow(now, t) {
		if task.Repeat == "" {
			task.Date = now.Format(DateLayout)
		} else {
			task.Date = next
		}
	}

	// Добавляем  задачу в базу
	id, err := db.AddTask(&task)
	if err != nil {
		writeJson(w, map[string]string{"error": err.Error()})
		return
	}

	writeJson(w, map[string]string{"id": fmt.Sprintf("%d", id)})
}

// Получаем задачу по id
func getTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	if id == "" {
		writeJson(w, map[string]string{"error": "Не указан идентификатор"})
		return
	}

	task, err := db.GetTask(id)
	if err != nil {
		writeJson(w, map[string]string{"error": err.Error()})
		return
	}

	writeJson(w, task)
}

// Обновляем задачу
func updateTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task db.Task
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		writeJson(w, map[string]string{"error": "ошибка десериализации JSON"})
		return
	}

	// Проверка наличия обязательного id и заголовка
	if task.ID == "" {
		writeJson(w, map[string]string{"error": "обязателен идентификатор задачи"})
		return
	}
	if task.Title == "" {
		writeJson(w, map[string]string{"error": "обязателен заголовок задачи"})
		return
	}

	now := time.Now()

	// Если дата не указана — берём сегодняшнюю
	if task.Date == "" {
		task.Date = now.Format(DateLayout)
	}

	// Проверяем формат даты
	t, err := time.Parse(DateLayout, task.Date)
	if err != nil {
		writeJson(w, map[string]string{"error": "дата представлена в неверном формате"})
		return
	}

	// Если правило повторения указано — проверяем и вычисляем следующую дату
	var next string
	if task.Repeat != "" {
		next, err = NextDate(now, task.Date, task.Repeat)
		if err != nil {
			writeJson(w, map[string]string{"error": err.Error()})
			return
		}
	}

	// Если дата меньше сегодняшней — исправляем
	if afterNow(now, t) {
		if task.Repeat == "" {
			task.Date = now.Format(DateLayout)
		} else {
			task.Date = next
		}
	}

	// Добавляем задачу в БД
	if err = db.UpdateTask(&task); err != nil {
		writeJson(w, map[string]string{"error": err.Error()})
		return
	}

	writeJson(w, struct{}{})
}

// Отмечаем задачу выполненной: удаляем одноразовую или ставим новую дату для периодической
func doneTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	if id == "" {
		writeJson(w, map[string]string{"error": "Не указан идентификатор"})
		return
	}

	task, err := db.GetTask(id)
	if err != nil {
		writeJson(w, map[string]string{"error": err.Error()})
		return
	}

	// Одноразовая задача — удаляем
	if task.Repeat == "" {
		if err := db.DeleteTask(id); err != nil {
			writeJson(w, map[string]string{"error": err.Error()})
			return
		}
		writeJson(w, struct{}{})
		return
	}

	// Периодическая — считаем следующую дату
	now := time.Now()
	next, err := NextDate(now, task.Date, task.Repeat)
	if err != nil {
		writeJson(w, map[string]string{"error": err.Error()})
		return
	}

	if err := db.UpdateDate(next, id); err != nil {
		writeJson(w, map[string]string{"error": err.Error()})
		return
	}

	writeJson(w, struct{}{})
}

// Удаляем задачу
func deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	if id == "" {
		writeJson(w, map[string]string{"error": "Не указан идентификатор"})
		return
	}

	if err := db.DeleteTask(id); err != nil {
		writeJson(w, map[string]string{"error": err.Error()})
		return
	}

	writeJson(w, struct{}{})
}
