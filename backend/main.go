package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	_ "modernc.org/sqlite"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(os.Getenv("TODO_PORT"))

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

	appPath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	dbFile := filepath.Join(filepath.Dir(appPath), os.Getenv("TODO_DBFILE"))
	_, err = os.Stat(dbFile)

	var install bool
	if err != nil {
		install = true
	}

	if install == true {
		db, err := sql.Open("sqlite", dbFile)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer db.Close()
		db.Exec()
	}

	fmt.Println("Завершаем работу")
}
