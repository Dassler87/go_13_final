package main

import (
	"log"

	"os"

	"github.com/Dassler87/go_13_final/pkg/db"

	"github.com/Dassler87/go_13_final/pkg/server"
)

func main() {

	// Определяем путь к БД из TODO_DBFILE, если нет — ставим scheduler.db
	dbFile := os.Getenv("TODO_DBFILE")
	if dbFile == "" {
		dbFile = "/app/data/scheduler.db"

	}
	log.Printf("Инициализация БД по пути: %s", dbFile)

	// Инициализируем БД если ее нет
	if err := db.Init(dbFile); err != nil {
		log.Fatalf("не удалось инициализировать БД: %v", err)
	}

	// Используем defer, чтобы гарантировать закрытие соединения с БД
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("ошибка при закрытии БД: %v", err)
		}
	}()

	// Запускаем сервер, Получаем порт из TODO_PORT или по умолчанию 7540

	port := os.Getenv("TODO_PORT")

	// Запускаем HTTP-сервер
	if err := server.Run(port); err != nil {
		log.Fatal(err)
	}
}
