package api

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

func NextDate(now time.Time, dstart string, repeat string) (string, error) {
	parsedTime, err := time.Parse(layout, dstart)
	if err != nil {
		return "", err
	}

	if len(repeat) == 0 {
		return "", errors.New("repeat is empty")
	}

	var repeatDate time.Time

	switch repeat[0] {
	case 'd':
		rules := strings.Split(repeat, " ")
		if len(rules) != 2 {
			return "", fmt.Errorf("incorrect format for param repeat %v", repeat)
		}

		num, err := strconv.Atoi(rules[1])
		if err != nil {
			return "", fmt.Errorf("param repeat %v is incorrect, err: %v", repeat, err)
		}
		if num <= 0 || num > 400 {
			return "", fmt.Errorf("incorrect day: %v in param repeat", num)
		}
		repeatDate = parsedTime.AddDate(0, 0, num)
		for {
			if repeatDate.After(now) || repeatDate.Equal(now) {
				break
			}
			repeatDate = repeatDate.AddDate(0, 0, num)
		}
	case 'y':
		if len(repeat) > 1 {
			return "", fmt.Errorf("param repeat %v is incorrect", repeat)
		}
		repeatDate = parsedTime.AddDate(1, 0, 0)
		for {
			if repeatDate.After(now) || repeatDate.Equal(now) {
				break
			}
			repeatDate = repeatDate.AddDate(1, 0, 0)
		}
	case 'w':
		rules := strings.Split(repeat, " ")
		if len(rules) != 2 {
			return "", fmt.Errorf("incorrect format for param repeat %v", repeat)
		}

		weekDaysStr := strings.Split(rules[1], ",")
		if len(weekDaysStr) > 7 {
			return "", fmt.Errorf("incorrect format for param repeat %v", repeat)
		}

		var weekDays []int
		for _, v := range weekDaysStr {
			num, err := strconv.Atoi(v)
			if err != nil {
				return "", fmt.Errorf("param repeat %v is incorrect, err: %v", repeat, err)
			}
			if num <= 0 || num > 7 {
				return "", fmt.Errorf("param repeat %v is incorrect", repeat)
			}
			if num == 7 {
				num = 0
			}
			weekDays = append(weekDays, num)
		}

		sort.Ints(weekDays)

		weekDayNow := int(now.Weekday())

		var nextDayWeek = weekDays[0]
		isWeekDelay := true

		for _, v := range weekDays {
			if weekDayNow < v {
				nextDayWeek = v
				isWeekDelay = false
			}
		}

		delay := nextDayWeek - weekDayNow
		if isWeekDelay {
			delay += 7
		}
		repeatDate = now.AddDate(0, 0, delay)
	case 'm':
		if parsedTime.After(now) {
			now = parsedTime
		}
		rules := strings.Split(repeat, " ")

		if len(rules) < 2 {
			return "", fmt.Errorf("incorrect format for param repeat %v", repeat)
		}

		monthDaysStr := strings.Split(rules[1], ",")

		var monthDays []int
		for _, v := range monthDaysStr {
			num, err := strconv.Atoi(v)
			if err != nil {
				return "", fmt.Errorf("param repeat %v is incorrect, err: %v", repeat, err)
			}
			if num < -2 || num > 31 || num == 0 {
				return "", fmt.Errorf("param repeat %v is incorrect", repeat)
			}
			monthDays = append(monthDays, num)
		}

		loc := now.Location()
		daysInMonth := func(y int, m time.Month) int {
			return time.Date(y, m+1, 0, 0, 0, 0, 0, loc).Day()
		}
		resolveDay := func(y int, m time.Month, d int) (int, bool) {
			switch {
			case d == -1:
				return daysInMonth(y, m), true
			case d == -2:
				return daysInMonth(y, m) - 1, true
			case d > 0 && d <= 31:
				if d > daysInMonth(y, m) {
					return 0, false
				}
				return d, true
			default:
				return 0, false
			}
		}

		var months []int
		if len(rules) == 3 {
			for _, v := range strings.Split(rules[2], ",") {
				num, err := strconv.Atoi(v)
				if err != nil {
					return "", fmt.Errorf("param repeat %v is incorrect, err: %v", repeat, err)
				}
				if num <= 0 || num > 12 {
					return "", fmt.Errorf("param repeat %v is incorrect", repeat)
				}
				months = append(months, num)
			}
			sort.Ints(months)
		}

		found := false
		for monthOffset := 0; monthOffset < 240; monthOffset++ {
			base := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc).AddDate(0, monthOffset, 0)
			y, m, _ := base.Date()
			if len(months) > 0 {
				okMonth := false
				for _, mm := range months {
					if int(m) == mm {
						okMonth = true
						break
					}
				}
				if !okMonth {
					continue
				}
			}
			dayNow := now.Day()
			if monthOffset > 0 {
				dayNow = 0
			}
			var resolvedDays []int
			for _, td := range monthDays {
				day, ok := resolveDay(y, m, td)
				if ok {
					resolvedDays = append(resolvedDays, day)
				}
			}
			sort.Ints(resolvedDays)
			for _, day := range resolvedDays {
				if day <= dayNow {
					continue
				}
				repeatDate = time.Date(y, m, day, 0, 0, 0, 0, loc)
				found = true
				break
			}
			if found {
				break
			}
		}
		if !found {
			return "", fmt.Errorf("param repeat %v is incorrect", repeat)
		}
	default:
		return "", fmt.Errorf("param repeat %v is incorrect", repeat)
	}

	return repeatDate.Format(layout), nil
}
