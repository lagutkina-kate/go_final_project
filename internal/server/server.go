package server

import (
	"go-final-project/internal/api"
	"go-final-project/internal/db"
	"log"
	"net/http"
	"time"
)

type SchedulerServer struct {
	Logger *log.Logger
	Server *http.Server
	DB     *db.SchedulerRepo
}

func NewServer(log *log.Logger, port string, repo *db.SchedulerRepo, password string) *SchedulerServer {
	router := http.NewServeMux()
	a := api.NewAPI(repo, log, password)

	router.Handle("/", http.FileServer(http.Dir("./web")))
	router.HandleFunc("/api/nextdate", a.NextDateHandler)
	router.HandleFunc("/api/signin", a.SigninHandler)

	router.HandleFunc("/api/task", a.Auth(a.TaskHandler))
	router.HandleFunc("/api/tasks", a.Auth(a.TasksHandler))
	router.HandleFunc("/api/task/done", a.Auth(a.DoneTaskHandler))

	return &SchedulerServer{
		Logger: log,
		Server: &http.Server{
			Addr:         port,
			Handler:      router,
			ErrorLog:     log,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  15 * time.Second},
		DB: repo}
}
