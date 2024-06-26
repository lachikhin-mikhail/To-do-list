package main

import (
	"database/sql"
	"log"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func dbExists() bool {
	dbFile := os.Getenv("TODO_DBFILE")
	_, err := os.Stat(dbFile)
	var exists bool
	if err == nil {
		exists = true
	}
	return exists

}

func startDB() {
	dbFile := os.Getenv("TODO_DBFILE")
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Fatal(err)
		return
	}
	db.SetMaxIdleConns(2)
	db.SetMaxOpenConns(5)
	db.SetConnMaxIdleTime(time.Minute * 5)
	db.SetConnMaxLifetime(time.Hour)
	DB = db
}

func installDB() {
	dbFile := os.Getenv("TODO_DBFILE")
	db, err := sqlx.Connect("sqlite3", dbFile)
	if err != nil {
		log.Println(err)
		return
	}
	defer db.Close()
	installQuery, err := os.ReadFile("backend/install.sql")
	if err != nil {
		log.Println(err)
		return
	}
	_, err = db.Exec(string(installQuery))
	if err != nil {
		log.Println(err)
		return
	}
	db.Close()
}

func AddTask(task Task) (int64, error) {
	db := DB
	var id int64
	res, err := db.Exec("INSERT INTO scheduler (date, title, comment, repeat) VALUES (:date, :title, :comment, :repeat)",
		sql.Named("date", task.Date), sql.Named("title", task.Title),
		sql.Named("comment", task.Comment), sql.Named("repeat", task.Repeat))
	if err == nil {
		id, _ = res.LastInsertId()
	}
	return id, err
}

func GetTaskList() ([]Task, error) {
	db := DB
	var rowsLimit int = 15
	var tasks []Task

	rows, err := db.Query("SELECT * FROM scheduler ORDER BY id LIMIT :limit", sql.Named("limit", rowsLimit))
	if err != nil {
		return []Task{}, err
	}
	defer rows.Close()

	for rows.Next() {
		task := Task{}

		err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			log.Println(err)
			return []Task{}, err
		}
		tasks = append(tasks, task)

	}
	return tasks, nil
}
