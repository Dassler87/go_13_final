package api

import (
	"net/http"

	"github.com/Dassler87/go_13_final/pkg/db"
)

// Структура ответа со списком задач
type TasksResp struct {
	Tasks []*db.Task `json:"tasks"`
}

// Возвращает список задач, максимум 50
func tasksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJson(w, map[string]string{"error": "метод не поддерживается"})
		return
	}

	search := r.FormValue("search")
	tasks, err := db.Tasks(50, search)
	if err != nil {
		writeJson(w, map[string]string{"error": err.Error()})
		return
	}
	writeJson(w, TasksResp{
		Tasks: tasks,
	})
}
