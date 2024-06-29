package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/lachikhin-mikhail/go_final_project/internal/db"
)

func isID(id string) bool {
	isID, _ := regexp.Match("[0-9]+", []byte(id))
	return isID
}

// writeErr отправляет ответ от сервера с ошибкой в формате json
func writeErr(err error, w http.ResponseWriter) {
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
}

func writeEmptyJson(w http.ResponseWriter) {
	okResp := map[string]string{}
	resp, err := json.Marshal(okResp)
	if err != nil {
		log.Println(err)
	}
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func formatTask(task db.Task) (db.Task, error) {
	var date time.Time
	format := db.Format
	var err error

	if len(task.Date) == 0 || strings.ToLower(task.Date) == "today" {
		date = time.Now()
		task.Date = date.Format(format)

	} else {
		date, err = time.Parse(format, task.Date)
		if err != nil {
			log.Println(err)
			return db.Task{}, err
		}
	}
	if isID := isID(task.ID); !isID && task.ID != "" {
		err = fmt.Errorf("некорректный формат ID")
		return db.Task{}, err
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
				return db.Task{}, err
			}
		case len(task.Repeat) == 0:
			task.Date = time.Now().Format(format)
		}

	}
	return task, nil
}

func GetNextDateHandler(w http.ResponseWriter, r *http.Request) {

	q := r.URL.Query()
	now := q.Get("now")
	date := q.Get("date")
	repeat := q.Get("repeat")

	nowDate, err := time.Parse(db.Format, now)
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

func PostTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task db.Task
	var buf bytes.Buffer
	var err error
	var id int64

	write := func() {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		if err != nil {
			writeErr(err, w)
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

	task, err = formatTask(task)
	if err != nil {
		write()
		return
	}

	id, err = db.AddTask(task)
	write()
}

func PutTaskHandler(w http.ResponseWriter, r *http.Request) {
	var updatedTask db.Task
	var buf bytes.Buffer
	var err error

	write := func() {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		if err != nil {
			writeErr(err, w)
			return
		} else {
			writeEmptyJson(w)
			return
		}

	}
	_, err = buf.ReadFrom(r.Body)
	if err != nil {
		write()
		return
	}

	if err = json.Unmarshal(buf.Bytes(), &updatedTask); err != nil {
		write()
		return
	}

	updatedTask, err = formatTask(updatedTask)
	if err != nil {
		write()
		return
	}

	err = db.PutTask(updatedTask)
	write()

}

func GetTaskHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	var task db.Task

	write := func() {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		var resp []byte
		if err != nil {
			writeErr(err, w)
			return
		} else {
			resp, err = json.Marshal(task)
		}

		if err != nil {
			log.Println(err)
		}
		w.WriteHeader(http.StatusOK)
		w.Write(resp)

	}

	q := r.URL.Query()
	id := q.Get("id")

	task, err = db.GetTaskByID(id)
	if err != nil {
		log.Println(err)
	}
	write()

}

func GetTasksHandler(w http.ResponseWriter, r *http.Request) {
	var tasks []db.Task
	var err error
	var date time.Time
	format := db.Format

	// write отправляет клиенту ответ либо ошибку, в формате json
	write := func() {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		var resp []byte
		if err != nil {
			writeErr(err, w)
			return
		} else {
			if len(tasks) == 0 {
				tasksResp := map[string][]db.Task{
					"tasks": {},
				}
				resp, err = json.Marshal(tasksResp)
			} else {
				tasksResp := map[string][]db.Task{
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
		tasks, err = db.GetTaskList()

	case isDate:
		date, err = time.Parse("02.01.2006", search)
		if err == nil {
			search = date.Format(format)
			tasks, err = db.GetTaskList(search)
			break
		}
		fallthrough

	default:
		search = fmt.Sprint("%" + search + "%")
		tasks, err = db.GetTaskList(search)

	}

	if err != nil {
		log.Println(err)
	}

	write()

}

func PostTaskDoneHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	id := q.Get("id")
	isID := isID(id)
	if !isID {
		writeErr(fmt.Errorf("некорректный формат id"), w)
		return
	}
	task, err := db.GetTaskByID(id)
	if err != nil {
		writeErr(err, w)
		return
	}
	if len(task.Repeat) == 0 {
		err = db.DeleteTask(id)
		if err != nil {
			writeErr(err, w)
			return
		}
		writeEmptyJson(w)
		return
	} else {
		nextDate, err := NextDate(time.Now(), task.Date, task.Repeat)
		if err != nil {
			writeErr(err, w)
			return
		}
		task.Date = nextDate
	}
	err = db.PutTask(task)
	if err != nil {
		writeErr(err, w)
		return
	}
	writeEmptyJson(w)

}

func DeleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	id := q.Get("id")
	isID := isID(id)
	if !isID {
		writeErr(fmt.Errorf("некорректный формат id"), w)
		return
	}

	err := db.DeleteTask(id)
	if err != nil {
		writeErr(err, w)
		return
	}
	writeEmptyJson(w)

}
