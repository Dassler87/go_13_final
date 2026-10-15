package server

import (
	"fmt"
	"net/http"
	"os"

	"github.com/Dassler87/go_13_final/pkg/api"
)

// Запускает HTTP-сервер
// Если порт не задан — берёт 7540
func Run(port string) error {
	webDir := "web"

	if port == "" {
		port = "7540"
	}
	port = ":" + port

	// Регистрируем все API-маршруты
	api.Init()

	_, err := os.Stat(webDir)
	if err == nil {
		// Если папка web есть — раздаём файлы оттуда по /
		http.Handle("/", http.FileServer(http.Dir(webDir)))
	} else {
		fmt.Printf("Папка %s не найдена, статика не подключена\n", webDir)
	}

	fmt.Printf("Сервер запущен на http://localhost%s\n", port)
	return http.ListenAndServe(port, nil)
}
