package main

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"
)

// NextDate возвращает дату и ошибку, исходя из правил указанных в repeat.
func NextDate(now time.Time, date string, repeat string) (string, error) {
	if repeat == "" {
		return "", fmt.Errorf("пустая строка")
	}

	format := "20060102"

	startDate, err := time.Parse(format, date)
	if err != nil {
		return "", err
	}

	var nextDate string = ""

	repeat = strings.ToLower(repeat)
	prefix := string(repeat[0])
	code, _ := strings.CutPrefix(repeat, prefix)
	code = strings.TrimSpace(code)

	switch prefix {
	case "d":
		days, err := strconv.Atoi(code)
		if err != nil {
			return "", err
		}
		if days > 400 {
			return "", fmt.Errorf("слишком большой временной промежуток")
		}
		nextDateDT := startDate.AddDate(0, 0, days)
		if nextDateDT.Before(now) || nextDateDT.Equal(now) {
			// Исходя из тестов вижу что ожидается при текущем дне > чем стартовый день + разница, возвращать следующий к текущему день?.. Но в тз такого не было :')
			nextDateDT = now.AddDate(0, 0, 1)
		}
		nextDate = nextDateDT.Format(format)

	case "y":
		nextDateDT := startDate
		for nextDateDT.Before(startDate) || nextDateDT.Equal(startDate) || nextDateDT.Before(now) || nextDateDT.Equal(now) {
			nextDateDT = nextDateDT.AddDate(1, 0, 0)
		}
		nextDate = nextDateDT.Format(format)

	case "w":
		var days int
		targetWDs, err := listAtoi(strings.Split(code, ","))
		if err != nil {
			return "", err
		}
		idx := slices.IndexFunc(targetWDs, func(wd int) bool { return wd > 7 || wd < 1 })
		if idx != -1 {
			return "", fmt.Errorf("некорректный формат w")
		}

		switch {
		case startDate.After(now) || startDate.Equal(now):
			startWD := int(startDate.Weekday())
			closestWD := closestWD(startDate, targetWDs)
			days = daysBetweenWD(startWD, closestWD)
			if closestWD == -1 || days == -1 {
				return "", fmt.Errorf("ошибка вычисления дней недели")
			}
		case startDate.Before(now):
			currentWD := int(now.Weekday())

			closestWD := closestWD(now, targetWDs)
			days = daysBetweenWD(currentWD, closestWD)
			if closestWD == -1 || days == -1 {
				return "", fmt.Errorf("ошибка вычисления дней недели")
			}
		}
		nextDate = now.AddDate(0, 0, days).Format(format)

	case "m":
		// Текущая дата
		currentMonth := int(now.Month())
		currentDay := now.Day()
		currentYear := now.Year()

		// Разделяем repeat на указания
		codes := strings.Split(code, " ")
		// Если через пробел указано больше 2, или меньше 1 указания, то repeat был указан неправильно
		if len(codes) > 2 || len(codes) < 1 {
			return "", fmt.Errorf("некорректный формат repeat")
		}
		// Дни в которые должно происходить повторение
		targetDays, err := listAtoi(strings.Split(codes[0], ","))
		if err != nil {
			return "", err
		}
		// Проверяем, что эти дни не превышают максимальное количество дней в месяце
		idx := slices.IndexFunc(targetDays, func(d int) bool { return d > 31 || d < -31 })
		if idx != -1 {
			return "", fmt.Errorf("некорректный формат repeat")
		}
		// Проверяем, есть ли указания по месяцам, если есть то сохраняем их в targetMonths
		var targetMonths []int
		var monthSpecified bool
		if len(codes) > 1 {
			targetMonths, err = listAtoi(strings.Split(codes[1], ","))
			if err != nil {
				return "", err
			}
			idx := slices.IndexFunc(targetMonths, func(d int) bool { return d > 12 })
			if idx != -1 {
				return "", fmt.Errorf("некорректный формат repeat")
			}
			slices.Sort(targetMonths)
			monthSpecified = true
		}
		// Смотрим, будет ли ближайшая подходящая дата в этом месяце или году
		isNextMonth := isNextMonth(now, startDate, targetDays)
		isNextYear := isNextYear(isNextMonth, now, startDate, targetMonths)

		// Готовим переменные, из которых будет собирать следующую дату
		var nextYear, nextMonth, nextDay int
		switch {
		// В этом месяце
		case !isNextMonth && (slices.Contains(targetMonths, currentMonth) || !monthSpecified):
			nextYear = currentYear
			nextMonth = currentMonth

			targetDays = processTargetDays(nextYear, time.Month(nextMonth), targetDays)
			// Поправка на дату начала
			var idx int
			startDay := pickBigger(currentDay, startDate.Day())
			if currentDay == startDate.Day() || currentDay == startDay {
				idx = getClosestIdx(startDay, targetDays)
				if idx == -1 {
					return "", fmt.Errorf("ошибка в вычисление ближайшего дня")
				}
			} else {
				idx = getClosestIdx(startDay-1, targetDays)
				if idx == -1 {
					return "", fmt.Errorf("ошибка в вычисление ближайшего дня")
				}
			}
			nextDay = targetDays[idx]

		// В этом году, в другом месяце, есть определённый месяц
		case !isNextYear && isNextMonth && monthSpecified:
			nextYear = currentYear
			// Поправка на дату начала

			startMonth := pickBigger(currentMonth, int(startDate.Month()))
			if currentMonth < int(startDate.Month()) { // tbf??
				startMonth--
			}
			idx := getClosestIdx(startMonth, targetMonths)
			if idx == -1 {
				return "", fmt.Errorf("ошибка в вычисление ближайшего месяца")
			}
			nextMonth = targetMonths[idx]

			targetDays = processTargetDays(nextYear, time.Month(nextMonth), targetDays)
			nextDay = targetDays[0]

		// В этом году, в другом месяце, нет требований к месяцу
		case !isNextYear && isNextMonth && !monthSpecified:
			nextYear = currentYear
			nextMonth = currentMonth + 1
			if startDate.Year() == currentYear {
				nextMonth = pickBigger(nextMonth, int(startDate.Month()))
			}
			targetDays = processTargetDays(nextYear, time.Month(nextMonth), targetDays)
			if nextMonth == int(startDate.Month()) {
				nextDay = targetDays[getClosestIdx(startDate.Day()-1, targetDays)]
			} else { // Если следующий месяц больше чем начальный месяц, значит нам не нужно переживать при выборе дня, он будет позже начала
				nextDay = targetDays[0]
			}

		// В следующем году
		case isNextYear:
			if startDate.Year() > currentYear {
				nextYear = startDate.Year()
				if monthSpecified {
					if targetMonths[0] > int(startDate.Month()) {
						nextMonth = targetMonths[0]
					} else {
						idx := getClosestIdx(int(startDate.Month()), targetMonths)
						if idx == -1 {
							return "", fmt.Errorf("ошибка в вычисление ближайшего месяца")
						}
						nextMonth = targetMonths[idx]
					}
				} else {
					nextMonth = int(startDate.Month())
				}
				targetDays = processTargetDays(nextYear, time.Month(nextMonth), targetDays)
				idx := getClosestIdx(startDate.Day(), targetDays)
				if idx == -1 {
					return "", fmt.Errorf("ошибка в вычисление ближайшего дня")
				}
				nextDay = targetDays[idx]
			} else {
				nextYear = currentYear + 1
				if monthSpecified {
					nextMonth = targetMonths[0]
				} else {
					nextMonth = 1
				}
				targetDays = processTargetDays(nextYear, time.Month(nextMonth), targetDays)
				nextDay = targetDays[0]

			}
		default:
			return "", fmt.Errorf("ошибка в case m")
		}

		nextDate = time.Date(nextYear, time.Month(nextMonth), nextDay, 0, 0, 0, 0, time.UTC).Format(format)
	default:
		return "", fmt.Errorf("некорректный формат repeat")
	}

	return nextDate, nil

}
