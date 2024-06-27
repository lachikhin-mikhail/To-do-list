package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
)

func getNextDate(w http.ResponseWriter, r *http.Request) {

	q := r.URL.Query()
	now := q.Get("now")
	date := q.Get("date")
	repeat := q.Get("repeat")

	nowDate, err := time.Parse(Format, now)
	if err != nil {
		fmt.Println(err)
		return
	}

	nextDate, err := NextDate(nowDate, date, repeat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	resp := []byte(nextDate)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func postTask(w http.ResponseWriter, r *http.Request) {
	var task Task
	var buf bytes.Buffer
	var err error
	var id int64
	var date time.Time
	format := Format

	write := func() {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		if err != nil {
			log.Println(err)
			errResp := map[string]string{
				"error": err.Error(),
			}
			resp, err := json.Marshal(errResp)
			if err != nil {
				log.Println(err)
			}
			w.WriteHeader(http.StatusBadRequest)
			w.Write(resp)
			return
		} else {
			idResp := map[string]int64{
				"id": id,
			}
			resp, err := json.Marshal(idResp)
			if err != nil {
				log.Println(err)
			}
			w.WriteHeader(http.StatusCreated)
			w.Write(resp)
			return
		}

	}

	_, err = buf.ReadFrom(r.Body)
	if err != nil {
		write()
		return
	}

	if err = json.Unmarshal(buf.Bytes(), &task); err != nil {
		write()
		return
	}
	if len(task.Date) == 0 || strings.ToLower(task.Date) == "today" {
		date = time.Now()
		task.Date = date.Format(format)

	} else {
		date, err = time.Parse(Format, task.Date)
		if err != nil {
			log.Println(err)
			write()
			return
		}
	}

	// Даты с временем приведённым к 00:00:00
	dateTrunc := date.Truncate(time.Hour * 24)
	nowTrunc := time.Now().Truncate(time.Hour * 24)

	if dateTrunc.Before(nowTrunc) {
		switch {
		case len(task.Repeat) > 0:
			task.Date, err = NextDate(time.Now(), task.Date, task.Repeat)
			if err != nil {
				log.Println(err)
				write()
				return
			}
		case len(task.Repeat) == 0:
			task.Date = time.Now().Format(format)

		}

	}

	id, err = AddTask(task)
	write()
}

func getTask(w http.ResponseWriter, r *http.Request) {

}

func getTasks(w http.ResponseWriter, r *http.Request) {
	var tasks []Task
	var err error
	var date time.Time
	format := Format

	// write отправляет клиенту ответ либо ошибку, в формате json
	write := func() {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		var resp []byte
		if err != nil {
			log.Println(err)
			errResp := map[string]string{
				"error": err.Error(),
			}
			resp, err := json.Marshal(errResp)
			if err != nil {
				log.Println(err)
			}
			w.WriteHeader(http.StatusBadRequest)
			w.Write(resp)
			return
		} else {
			if len(tasks) == 0 {
				tasksResp := map[string][]Task{
					"tasks": {},
				}
				resp, err = json.Marshal(tasksResp)
			} else {
				tasksResp := map[string][]Task{
					"tasks": tasks,
				}
				resp, err = json.Marshal(tasksResp)

			}

			if err != nil {
				log.Println(err)
			}
			w.WriteHeader(http.StatusCreated)
			w.Write(resp)
			return
		}

	}

	q := r.URL.Query()
	search := q.Get("search")
	isDate, _ := regexp.Match("[0-9]{2}.[0-9]{2}.[0-9]{4}", []byte(search))

	switch {
	case len(search) == 0:
		tasks, err = GetTaskList()

	case isDate:
		date, err = time.Parse("02.01.2006", search)
		if err == nil {
			search = date.Format(format)
			tasks, err = GetTaskList(search)
			break
		}
		fallthrough

	default:
		search = fmt.Sprint("%" + search + "%")
		tasks, err = GetTaskList(search)

	}

	if err != nil {
		log.Println(err)
	}

	write()

}
