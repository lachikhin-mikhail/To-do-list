package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/lachikhin-mikhail/go_final_project/internal/db"
)

// task.go содержит обработчики запросов к api/task

// PostTaskHandler обрабатывает запрос с методом POST.
// Если пользователь авторизован и задача отправлена в корректном формате, добавляет новую задачу в базу данных.
// Возвращает JSON {"id": string} или JSON {"error": error} в случае ошибки.
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

	if err = getAndVerifyToken(r); err != nil {
		write()
		return
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

// PutTaskHandler обрабатывает запрос с методом PUT.
// Если пользователь авторизован и задача существует, и отправлена в корректном формате, обновляет поля задачи в базе данных.
// Возвращает пустой JSON {} или JSON {"error": error} в случае ошибки.
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
	if err = getAndVerifyToken(r); err != nil {
		write()
		return
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

// GetTaskHandler обрабатывает запрос с методом GET.
// Если пользователь авторизован, возвращает задачу с указанным ID.
// Возвращает JSON {"task":Task}, или JSON {"error": error} при ошибке.
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

	if err = getAndVerifyToken(r); err != nil {
		write()
		return
	}

	q := r.URL.Query()
	id := q.Get("id")

	task, err = db.GetTaskByID(id)
	if err != nil {
		log.Println(err)
	}
	write()

}

// DeleteTaskHandler обрабатывает запрос к api/task с методом DELETE.
// Если пользователь авторизован и id существует, удаляет задачу.
// При успешном выполнение возвращает пустой JSON {}. Иначе возвращает JSON {"error":error}.
func DeleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	var err error

	if err = getAndVerifyToken(r); err != nil {
		writeErr(err, w)
		return
	}

	q := r.URL.Query()
	id := q.Get("id")
	isID := isID(id)
	if !isID {
		writeErr(fmt.Errorf("некорректный формат id"), w)
		return
	}

	err = db.DeleteTask(id)
	if err != nil {
		writeErr(err, w)
		return
	}
	writeEmptyJson(w)

}
