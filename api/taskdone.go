package api

import (
	"fmt"
	"net/http"
	"time"

	nd "github.com/lachikhin-mikhail/go_final_project/internal/nextdate"
)

// PostTaskDoneHandler обрабатывает запросы к /api/task/done с методом POST.
// Если пользователь авторизован, удаляет задачи не имеющих правил повторения repeat, или обновляет дату выполнения задач, имеющих правило repeat.
// Возвращает пустой JSON {} в случае успеха, или JSON {"error": error} при возникновение ошибки.
func PostTaskDoneHandler(w http.ResponseWriter, r *http.Request) {
	var err error

	q := r.URL.Query()
	id := q.Get("id")
	isID := isID(id)
	if !isID {
		writeErr(fmt.Errorf("некорректный формат id"), w)
		return
	}
	task, err := dbh.GetTaskByID(id)
	if err != nil {
		writeErr(err, w)
		return
	}
	if len(task.Repeat) == 0 {
		err = dbh.DeleteTask(id)
		if err != nil {
			writeErr(err, w)
			return
		}
		writeEmptyJson(w)
		return
	} else {
		nextDate, err := nd.NextDate(time.Now(), task.Date, task.Repeat)
		if err != nil {
			writeErr(err, w)
			return
		}
		task.Date = nextDate
	}
	err = dbh.PutTask(task)
	if err != nil {
		writeErr(err, w)
		return
	}
	writeEmptyJson(w)

}
