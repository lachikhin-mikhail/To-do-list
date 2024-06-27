package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Загружаем переменные среды
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println(err)
	}

	// Если бд не существует, создаём
	if !dbExists() {
		installDB()
	}

	// Запуск бд
	startDB()
	defer DB.Close()

	// Адрес для запуска сервера
	ip := ""
	port := os.Getenv("TODO_PORT")
	addr := fmt.Sprintf("%s:%s", ip, port)

	// Router
	r := chi.NewRouter()

	r.Handle("/*", http.FileServer(http.Dir("./web")))

	r.Get("/api/nextdate", getNextDate)
	r.Get("/api/tasks", getTasks)
	r.Get("/api/task", getTask)
	r.Post("/api/task", postTask)

	// Запуск сервера
	err = http.ListenAndServe(addr, r)
	if err != nil {
		panic(err)
	}
	NextDate(time.Now(), "", "")

	fmt.Println("Завершаем работу")

}
