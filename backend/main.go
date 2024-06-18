package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println(err)
	}

	dbFile := os.Getenv("TODO_DBFILE")
	_, err = os.Stat(dbFile)

	var install bool
	if err != nil {
		install = true
	}

	if install {
		db, err := sqlx.Connect("sqlite3", dbFile)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer db.Close()
		installQuery, err := os.ReadFile("./backend/install.sql")
		if err != nil {
			fmt.Println(err)
			return
		}
		_, err = db.Exec(string(installQuery))
		if err != nil {
			fmt.Println(err)
			return
		}
		db.Close()
	}

	// Адрес для запуска сервера
	ip := ""
	port := os.Getenv("TODO_PORT")
	addr := fmt.Sprintf("%s:%s", ip, port)

	// Запускаем сервер
	fmt.Println("Запускаем сервер")
	err = http.ListenAndServe(addr, http.FileServer(http.Dir("web/")))
	if err != nil {
		panic(err)
	}

	fmt.Println("Завершаем работу")
}
