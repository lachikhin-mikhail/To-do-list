package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lachikhin-mikhail/go_final_project/internal/db"
)

// verifyToken проверяет токен на подлинность, возвращает true если токен корректен
func verifyToken(signedToken string) bool {
	password := []byte(os.Getenv("TODO_PASSWORD"))

	jwtToken, err := jwt.Parse(signedToken, func(t *jwt.Token) (interface{}, error) {
		return password, nil
	})
	if err != nil {
		log.Printf("Failed to parse token: %s\n", err)
	}
	if jwtToken.Valid {
		return true
	}
	return false
}

// getAndVerifyToken проверяет cookie на наличие токена авторизации, и проверяет его подлинность.
// Возвращает ошибку, если токен не найден, или токен не прошёл проверку.
func getAndVerifyToken(r *http.Request) error {
	token, err := r.Cookie("token")

	if err != nil {
		return err
	}
	if verifyToken(token.Value) {
		return nil
	}
	return fmt.Errorf("ошибка авторизации")
}

// isID возвращает true если переданная строка содержит только символы, которые могут находится в строке ID в базе данных.
func isID(id string) bool {
	isID, _ := regexp.Match("[0-9]+", []byte(id))
	return isID
}

// writeErr пишет ошибку в response в формате JSON и статус запроса BadRequest
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

// writeEmptyJson пишет в response пустой JSON {} и статус запроса OK
func writeEmptyJson(w http.ResponseWriter) {
	okResp := map[string]string{}
	resp, err := json.Marshal(okResp)
	if err != nil {
		log.Println(err)
	}
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

// formatTask проверяет переданную задачу Task на корректность полей, а так же корректирует дату задачи.
// Возвращает отформатированную задачу или ошибку.
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
