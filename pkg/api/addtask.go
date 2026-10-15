package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Dassler87/go_13_final/pkg/db"
)

// Универсальный ответ с ошибкой
func writeError(w http.ResponseWriter, status int, message string) {
	w.WriteHeader(status)
	resp := map[string]string{"error": message}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if encErr := json.NewEncoder(w).Encode(resp); encErr != nil {
		log.Printf("error encoding JSON for status %d: %v", status, encErr)
	}
}

// Создаём задачу
func addTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "метод не поддерживается")
		return
	}

	var task db.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		// Ошибка формата JSON → 400 Bad Request
		log.Printf("JSON decode error: %v", err)

		writeError(w, http.StatusBadRequest, "некорректный JSON")
		return
	}

	// Если заголовка нет - ошибка
	if task.Title == "" {
		writeError(w, http.StatusBadRequest, "поле title обязательно")
		return
	}

	now := time.Now()

	// Если дата не указана — берём сегодняшнюю
	if task.Date == "" {
		task.Date = now.Format(db.DateLayout)
	}

	// Проверяем формат даты
	t, err := time.Parse(db.DateLayout, task.Date)
	if err != nil {
		writeError(w, http.StatusBadRequest, "дата представлена в неверном формате. Ожидайте YYYYMMDD")
		return
	}

	// Если правило повторения указано — проверяем и вычисляем следующую дату
	var next string
	if task.Repeat != "" {
		next, err = NextDate(now, task.Date, task.Repeat)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
	}

	// Если дата меньше сегодняшней — корректируем
	if afterNow(now, t) {
		if task.Repeat == "" {
			task.Date = now.Format(db.DateLayout)
		} else {
			task.Date = next
		}
	}

	// Добавляем  задачу в базу
	id, err := db.AddTask(&task)
	if err != nil {
		log.Printf("db.AddTask error: %v", err)

		writeError(w, http.StatusInternalServerError, "ошибка при сохранении задачи в базе данных")
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	resp := map[string]string{"id": fmt.Sprintf("%d", id)}
	if encErr := json.NewEncoder(w).Encode(resp); encErr != nil {
		log.Printf("error encoding JSON for 201: %v", encErr)
	}
}

// Получаем задачу по id
func getTaskHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.FormValue("id")
	if idStr == "" {
		writeError(w, http.StatusBadRequest, "не указан идентификатор")
		return
	}

	var id int
	_, err := fmt.Sscanf(idStr, "%d", &id)
	if err != nil {
		writeError(w, http.StatusBadRequest, "неверный формат id: ожидается число")
		return
	}

	task, err := db.GetTask(id)
	if err != nil {
		log.Printf("db.GetTask error: %v", err)

		if err == sql.ErrNoRows {
			writeError(w, http.StatusNotFound, "задача не найдена")
		} else {
			writeError(w, http.StatusInternalServerError, "внутренняя ошибка сервера")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if encErr := json.NewEncoder(w).Encode(task); encErr != nil {
		log.Printf("error encoding JSON for 200: %v", encErr)
	}
}

// Обновляем задачу
func updateTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "метод не поддерживается. Требуется PUT")
		return
	}

	var task db.Task
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		log.Printf("JSON decode error in updateTask: %v", err)
		writeError(w, http.StatusBadRequest, "ошибка при чтении данных задачи")
		return
	}

	// Проверка наличия обязательного id и заголовка
	if task.ID == "" {
		writeError(w, http.StatusBadRequest, "обязателен идентификатор задачи")
		return
	}
	var id int
	_, err = fmt.Sscanf(task.ID, "%d", &id)
	if err != nil {
		writeError(w, http.StatusBadRequest, "неверный формат id: ожидается число")
		return
	}

	if task.Title == "" {
		writeError(w, http.StatusBadRequest, "обязателен заголовок задачи")
		return
	}

	now := time.Now()

	// Если дата не указана — берём сегодняшнюю
	if task.Date == "" {
		task.Date = now.Format(db.DateLayout)
	}

	// Проверяем формат даты
	t, err := time.Parse(db.DateLayout, task.Date)
	if err != nil {
		writeError(w, http.StatusBadRequest, "дата представлена в неверном формате. Ожидайте YYYYMMDD")
		return
	}

	// Если правило повторения указано — проверяем и вычисляем следующую дату
	var next string
	if task.Repeat != "" {
		next, err = NextDate(now, task.Date, task.Repeat)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
	}

	// Если дата меньше сегодняшней — исправляем
	if afterNow(now, t) {
		if task.Repeat == "" {
			task.Date = now.Format(db.DateLayout)
		} else {
			task.Date = next
		}
	}

	// Добавляем задачу в БД
	if err = db.UpdateTask(&task, id); err != nil {
		log.Printf("db.UpdateTask error: %v", err)

		writeError(w, http.StatusInternalServerError, "ошибка при обновлении задачи")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Отмечаем задачу выполненной: удаляем одноразовую или ставим новую дату для периодической
func doneTaskHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.FormValue("id")
	if idStr == "" {
		writeError(w, http.StatusBadRequest, "не указан идентификатор")
		return
	}

	var intID int
	_, err := fmt.Sscanf(idStr, "%d", &intID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "неверный формат id: ожидается число")
		return
	}

	task, err := db.GetTask(intID)
	if err != nil {
		log.Printf("db.GetTask error in doneTask: %v", err)

		if err == sql.ErrNoRows {
			writeError(w, http.StatusNotFound, "задача не найдена")
		} else {
			writeError(w, http.StatusInternalServerError, "внутренняя ошибка сервера")
		}
		return
	}

	// Одноразовая задача — удаляем
	if task.Repeat == "" {
		if err := db.DeleteTask(intID); err != nil {
			log.Printf("db.DeleteTask error: %v", err)

			if err == sql.ErrNoRows {
				writeError(w, http.StatusNotFound, "задача не найдена")
			} else {
				writeError(w, http.StatusInternalServerError, "ошибка при удалении задачи")
			}
			return
		}

		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Периодическая — считаем следующую дату
	now := time.Now()
	next, err := NextDate(now, task.Date, task.Repeat)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := db.UpdateDate(next, intID); err != nil {
		log.Printf("db.UpdateDate error: %v", err)

		if err == sql.ErrNoRows {
			writeError(w, http.StatusNotFound, "задача не найдена")
		} else {
			writeError(w, http.StatusInternalServerError, "ошибка при обновлении даты задачи")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Удаляем задачу
func deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeError(w, http.StatusMethodNotAllowed, "метод не поддерживается. Используйте DELETE")
		return
	}

	idStr := r.FormValue("id")
	if idStr == "" {
		writeError(w, http.StatusBadRequest, "не указан идентификатор")
		return
	}

	var intID int
	_, err := fmt.Sscanf(idStr, "%d", &intID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "неверный формат id: ожидается число")
		return
	}

	if err := db.DeleteTask(intID); err != nil {
		log.Printf("db.DeleteTask error: %v", err)

		if err == sql.ErrNoRows {
			writeError(w, http.StatusNotFound, "задача не найдена")
		} else {
			writeError(w, http.StatusInternalServerError, "ошибка при удалении задачи")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
