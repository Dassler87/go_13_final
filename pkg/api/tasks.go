package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/Dassler87/go_13_final/pkg/db"
)

// Лимит задач по умолчанию
const DefaultLimitTasks = 50

// Структура ответа со списком задач
type TasksResp struct {
	Tasks []*db.Task `json:"tasks"`
}

// Возвращает список задач
func tasksHandler(w http.ResponseWriter, r *http.Request) {
	// Проверка метода: только GET
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "метод не поддерживается. Требуется GET")
		return
	}

	search := r.FormValue("search")
	tasks, err := db.Tasks(DefaultLimitTasks, search)
	if err != nil {
		log.Printf("db.Tasks error: %v", err)
		writeError(w, http.StatusInternalServerError, "не удалось получить список задач")
		return
	}
	//Успешный ответ 200 OK
	resp := TasksResp{
		Tasks: tasks,
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("error encoding JSON for 200: %v", err)
	}
}
