package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const DateLayout = "20060102"

// Сравнивает только даты (без времени)
func afterNow(date, now time.Time) bool {
	y1, m1, d1 := date.Date()
	y2, m2, d2 := now.Date()
	if y1 != y2 {
		return y1 > y2
	}
	if m1 != m2 {
		return m1 > m2
	}
	return d1 > d2
}

// Считает сколько дней в месяце у даты t
func daysInMonth(t time.Time) int {
	return time.Date(t.Year(), t.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

// NextDate вычисляет следующую дату по правилу
func NextDate(now time.Time, dstart string, repeat string) (string, error) {
	if repeat == "" {
		return "", fmt.Errorf("правило повторения не указано")
	}

	date, err := time.Parse(DateLayout, dstart)
	if err != nil {
		return "", fmt.Errorf("неверный формат даты: %v", err)
	}

	parts := strings.Split(repeat, " ")
	rule := parts[0]

	switch rule {
	case "d":
		if len(parts) != 2 {
			return "", fmt.Errorf("не указан интервал в днях")
		}
		days, err := strconv.Atoi(parts[1])
		if err != nil {
			return "", fmt.Errorf("неверный интервал дней: %v", err)
		}
		if days < 1 || days > 400 {
			return "", fmt.Errorf("интервал дней должен быть от 1 до 400")
		}
		for {
			date = date.AddDate(0, 0, days)
			if afterNow(date, now) {
				break
			}
		}

	case "y":
		for {
			date = date.AddDate(1, 0, 0)
			if afterNow(date, now) {
				break
			}
		}

	case "w":
		if len(parts) != 2 {
			return "", fmt.Errorf("не указаны дни недели")
		}
		var day [8]bool
		for _, s := range strings.Split(parts[1], ",") {
			n, err := strconv.Atoi(s)
			if err != nil {
				return "", fmt.Errorf("неверный день недели: %s", s)
			}
			if n < 1 || n > 7 {
				return "", fmt.Errorf("день недели должен быть от 1 до 7")
			}
			day[n] = true
		}
		for {
			date = date.AddDate(0, 0, 1)
			wd := int(date.Weekday())
			if wd == 0 {
				wd = 7
			}
			if day[wd] && afterNow(date, now) {
				break
			}
		}

	case "m":
		if len(parts) < 2 || len(parts) > 3 {
			return "", fmt.Errorf("неверный формат правила для месяцев")
		}

		var day [32]bool
		var month [13]bool
		var lastDay bool
		var secondLastDay bool

		for _, s := range strings.Split(parts[1], ",") {
			n, err := strconv.Atoi(s)
			if err != nil {
				return "", fmt.Errorf("неверный день месяца: %s", s)
			}
			switch {
			case n == -1:
				lastDay = true
			case n == -2:
				secondLastDay = true
			case n >= 1 && n <= 31:
				day[n] = true
			default:
				return "", fmt.Errorf("день месяца должен быть от 1 до 31 или -1, -2")
			}
		}

		if len(parts) == 3 {
			for _, s := range strings.Split(parts[2], ",") {
				n, err := strconv.Atoi(s)
				if err != nil {
					return "", fmt.Errorf("неверный месяц: %s", s)
				}
				if n < 1 || n > 12 {
					return "", fmt.Errorf("месяц должен быть от 1 до 12")
				}
				month[n] = true
			}
		} else {
			for i := 1; i <= 12; i++ {
				month[i] = true
			}
		}

		for {
			date = date.AddDate(0, 0, 1)
			d := date.Day()
			m := int(date.Month())

			if !month[m] {
				continue
			}

			match := day[d]
			if !match && lastDay {
				if d == daysInMonth(date) {
					match = true
				}
			}
			if !match && secondLastDay {
				if d == daysInMonth(date)-1 {
					match = true
				}
			}

			if match && afterNow(date, now) {
				break
			}
		}

	default:
		return "", fmt.Errorf("неподдерживаемый формат правила: %s", rule)
	}

	return date.Format(DateLayout), nil
}

// Обработчик считает следующую дату по правилам
func NextDateHandler(w http.ResponseWriter, r *http.Request) {
	nowStr := r.FormValue("now")
	date := r.FormValue("date")
	repeat := r.FormValue("repeat")

	var now time.Time
	if nowStr == "" {
		now = time.Now()
	} else {
		var err error
		now, err = time.Parse(DateLayout, nowStr)
		if err != nil {
			http.Error(w, "неверный формат now", http.StatusBadRequest)
			return
		}
	}

	result, err := NextDate(now, date, repeat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(result))
}
