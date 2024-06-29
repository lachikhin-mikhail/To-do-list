package main

import (
	"fmt"
	"net/http"
	"os"

	api "github.com/lachikhin-mikhail/go_final_project/api"
	db "github.com/lachikhin-mikhail/go_final_project/internal/db"

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
	if !db.DbExists() {
		db.InstallDB()
	}

	// Запуск бд
	db.StartDB()
	defer db.DB.Close()

	// Адрес для запуска сервера
	ip := ""
	port := os.Getenv("TODO_PORT")
	addr := fmt.Sprintf("%s:%s", ip, port)

	// Router
	r := chi.NewRouter()

	r.Handle("/*", http.FileServer(http.Dir("./web")))

	r.Get("/api/nextdate", api.GetNextDateHandler)
	r.Get("/api/tasks", api.GetTasksHandler)
	r.Get("/api/task", api.GetTaskHandler)
	r.Put("/api/task", api.PutTaskHandler)
	r.Post("/api/task", api.PostTaskHandler)
	r.Post("/api/task/done", api.PostTaskDoneHandler)
	r.Delete("/api/task", api.DeleteTaskHandler)

	// Запуск сервера
	err = http.ListenAndServe(addr, r)
	if err != nil {
		panic(err)
	}

	fmt.Println("Завершаем работу")

}
