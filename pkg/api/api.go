package api

import (
	"net/http"
)

// Настраиваем маршруты API
func Init() {
	http.HandleFunc("/api/nextdate", NextDateHandler)
	http.HandleFunc("/api/task", taskHandler)
	http.HandleFunc("/api/task/done", doneTaskHandler)
	http.HandleFunc("/api/tasks", tasksHandler)
}

// Обрабатывает /api/task: выбирает нужный хендлер по HTTP-методу (GET/POST/PUT/DELETE)
func taskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getTaskHandler(w, r)
	case http.MethodPost:
		addTaskHandler(w, r)
	case http.MethodPut:
		updateTaskHandler(w, r)
	case http.MethodDelete:
		deleteTaskHandler(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "метод не поддерживается")
	}
}
