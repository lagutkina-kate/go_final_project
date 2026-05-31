package api

import (
	"encoding/json"
	"fmt"
	"go-final-project/internal/db"
	"io"
	"net/http"
	"strconv"
	"time"
)

const SearchLayout = "02.01.2006"

func (a *API) addTaskHandler(w http.ResponseWriter, r *http.Request) {
	a.logger.Println("request: addTaskHandler()")

	var task db.Task

	body, err := io.ReadAll(r.Body)
	if err != nil {
		a.WriteResponse(w, http.StatusBadRequest, "error", err.Error())
		return
	}
	defer r.Body.Close()

	err = json.Unmarshal(body, &task)
	if err != nil {
		a.WriteResponse(w, http.StatusBadRequest, "error", err.Error())
		return
	}

	a.logger.Printf("task: %+v\n", task)

	now, err := calculateDate(task.Date, task.Repeat)
	if err != nil {
		a.WriteResponse(w, http.StatusBadRequest, "error", err.Error())
		return
	}

	if task.Title == "" {
		a.WriteResponse(w, http.StatusBadRequest, "error", "bad request")
		return
	}

	id, err := a.repo.AddTask(&db.Task{
		Date:    now.Format(layout),
		Title:   task.Title,
		Comment: task.Comment,
		Repeat:  task.Repeat,
	})
	if err != nil {
		a.WriteResponse(w, http.StatusInternalServerError, "error", err.Error())
		return
	}

	response := map[string]string{
		"id": fmt.Sprintf("%v", id),
	}
	a.logger.Printf("response: %v\n", response)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (a *API) GetTasksHandler(w http.ResponseWriter, r *http.Request) {
	paramSearch := r.URL.Query().Get("search")

	a.logger.Printf("request: GetTasksHandler() paramSearch: %v\n", paramSearch)

	if paramSearch == "" {
		res, err := a.repo.GetTasks(50)
		if err != nil {
			a.WriteResponse(w, http.StatusInternalServerError, "error", err.Error())
			return
		}
		a.WriteResponse(w, http.StatusOK, "tasks", res.Tasks)
		return
	}

	parsedSearch, err := time.Parse(SearchLayout, paramSearch)
	if err != nil {
		res, err := a.repo.SearchTasks(paramSearch, 50)
		if err != nil {
			a.WriteResponse(w, http.StatusInternalServerError, "error", err.Error())
			return
		}
		a.WriteResponse(w, http.StatusOK, "tasks", res.Tasks)
		return
	}

	res, err := a.repo.SearchTasksByDate(parsedSearch.Format(layout), 50)
	if err != nil {
		a.WriteResponse(w, http.StatusInternalServerError, "error", err.Error())
		return
	}
	a.WriteResponse(w, http.StatusOK, "tasks", res.Tasks)
}

func (a *API) GetTaskByIDHandler(w http.ResponseWriter, r *http.Request) {
	paramID := r.URL.Query().Get("id")

	a.logger.Printf("request: GetTaskByIDHandler() paramID: %v\n", paramID)

	if paramID == "" {
		a.WriteResponse(w, http.StatusBadRequest, "error", "id is empty")
		return
	}

	paramIDInt, err := strconv.Atoi(paramID)
	if err != nil {
		a.WriteResponse(w, http.StatusBadRequest, "error", err.Error())
		return
	}

	res, err := a.repo.GetTaskByID(paramIDInt)
	if err != nil {
		a.WriteResponse(w, http.StatusInternalServerError, "error", err.Error())
		return
	}

	if res == nil {
		a.WriteResponse(w, http.StatusOK, "error", "taskID is not found")
		return
	}

	a.logger.Printf("response: %+v\n", res)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func (a *API) UpdateTaskHandler(w http.ResponseWriter, r *http.Request) {
	a.logger.Println("request: UpdateTaskHandler()")
	var task db.Task

	body, err := io.ReadAll(r.Body)
	if err != nil {
		a.WriteResponse(w, http.StatusBadRequest, "error", err.Error())
		return
	}
	defer r.Body.Close()

	err = json.Unmarshal(body, &task)
	if err != nil {
		a.WriteResponse(w, http.StatusBadRequest, "error", err.Error())
		return
	}

	a.logger.Printf("task: %+v\n", task)

	if task.Title == "" {
		a.WriteResponse(w, http.StatusBadRequest, "error", "titile is empty")
		return
	}

	now, err := calculateDate(task.Date, task.Repeat)
	if err != nil {
		a.WriteResponse(w, http.StatusBadRequest, "error", err.Error())
		return
	}

	err = a.repo.UpdateTaskByID(&db.Task{
		ID:      task.ID,
		Date:    now.Format(layout),
		Title:   task.Title,
		Comment: task.Comment,
		Repeat:  task.Repeat,
	})
	if err != nil {
		a.WriteResponse(w, http.StatusInternalServerError, "error", err.Error())
		return
	}

	a.logger.Printf("response: %v\n", struct{}{})

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(struct{}{})
}

func (a *API) DoneTaskHandler(w http.ResponseWriter, r *http.Request) {
	paramID := r.URL.Query().Get("id")

	a.logger.Printf("request: DoneTaskHandler() paramID: %v\n", paramID)

	if paramID == "" {
		a.WriteResponse(w, http.StatusBadRequest, "error", "id is empty")
		return
	}

	paramIDInt, err := strconv.Atoi(paramID)
	if err != nil {
		a.WriteResponse(w, http.StatusBadRequest, "error", err.Error())
		return
	}

	res, err := a.repo.GetTaskByID(paramIDInt)
	if err != nil {
		a.WriteResponse(w, http.StatusInternalServerError, "error", err.Error())
		return
	}

	if res == nil {
		a.WriteResponse(w, http.StatusOK, "error", "taskID is not found")
		return
	}

	if res.Repeat == "" {
		err = a.repo.DeleteTaskByID(paramIDInt)
		if err != nil {
			a.WriteResponse(w, http.StatusBadRequest, "error", err.Error())
			return
		}
		a.logger.Printf("response: %v\n", struct{}{})

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(struct{}{})
		return
	}

	today := time.Now().Truncate(24 * time.Hour)
	nextDate, err := NextDate(today, res.Date, res.Repeat)
	if err != nil {
		a.WriteResponse(w, http.StatusInternalServerError, "error", err.Error())
		return
	}

	parsedNextDate, err := time.Parse(layout, nextDate)
	if err != nil {
		a.WriteResponse(w, http.StatusInternalServerError, "error", err.Error())
		return
	}

	res.Date = parsedNextDate.Format(layout)

	err = a.repo.UpdateTaskByID(res)
	if err != nil {
		a.WriteResponse(w, http.StatusInternalServerError, "error", err.Error())
		return
	}

	a.logger.Printf("response: %v\n", struct{}{})
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(struct{}{})
}

func (a *API) DeleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	paramID := r.URL.Query().Get("id")

	a.logger.Printf("request: DeleteTaskHandler() paramID: %v\n", paramID)

	if paramID == "" {
		a.WriteResponse(w, http.StatusBadRequest, "error", "id is empty")
		return
	}

	paramIDInt, err := strconv.Atoi(paramID)
	if err != nil {
		a.WriteResponse(w, http.StatusBadRequest, "error", err.Error())
		return
	}

	res, err := a.repo.GetTaskByID(paramIDInt)
	if err != nil {
		a.WriteResponse(w, http.StatusInternalServerError, "error", err.Error())
		return
	}

	if res == nil {
		a.WriteResponse(w, http.StatusOK, "error", "taskID is not found")
		return
	}

	err = a.repo.DeleteTaskByID(paramIDInt)
	if err != nil {
		a.WriteResponse(w, http.StatusBadRequest, "error", err.Error())
		return
	}

	a.logger.Printf("response: %v\n", struct{}{})

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(struct{}{})
}

func calculateDate(date, repeat string) (*time.Time, error) {
	var now time.Time

	if date == "" {
		now = time.Now()
	} else {
		parsedDate, err := time.Parse(layout, date)
		if err != nil {
			return nil, err
		}

		today := time.Now().Truncate(24 * time.Hour)

		if parsedDate.Before(today) {
			if repeat == "" {
				now = today
			} else {
				nextDate, err := NextDate(today, date, repeat)
				if err != nil {
					return nil, err
				}

				parsedNextDate, err := time.Parse(layout, nextDate)
				if err != nil {
					return nil, err
				}
				now = parsedNextDate
			}
		} else {
			now = parsedDate
		}
	}

	return &now, nil
}
