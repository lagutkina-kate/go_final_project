package db

import (
	_ "modernc.org/sqlite"

	"database/sql"
	"fmt"
	"os"
)

type SchedulerRepo struct {
	db *sql.DB
}

type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

type Tasks struct {
	Tasks []*Task `json:"tasks"`
}

func NewSchedulerRepo(dbFile string) (*SchedulerRepo, error) {
	_, err := os.Stat(dbFile)

	var install bool
	if err != nil {
		install = true
	}

	var db *sql.DB
	db, err = sql.Open("sqlite", "scheduler.db")
	if err != nil {
		return nil, err
	}

	sr := &SchedulerRepo{db: db}
	if install {
		sr.Init()
	}
	return sr, nil
}

func (s *SchedulerRepo) Init() error {
	_, err := s.db.Exec(`
	CREATE TABLE scheduler (
    	"id" INTEGER PRIMARY KEY,
   		"date"  CHAR(8) NOT NULL DEFAULT "",
    	"title" VARCHAR(100) NOT NULL DEFAULT "",
		"comment" TEXT NOT NULL DEFAULT "",
		"repeat" VARCHAR(128) NOT NULL DEFAULT "");`)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(`CREATE INDEX scheduler_dates ON scheduler (date);`)
	if err != nil {
		return err
	}

	return nil
}

func (s *SchedulerRepo) AddTask(t *Task) (int64, error) {
	result, err := s.db.Exec(
		"INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?);",
		t.Date, t.Title, t.Comment, t.Repeat)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (s *SchedulerRepo) GetTasks(num int) (*Tasks, error) {
	rows, err := s.db.Query(
		"SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date LIMIT ?;", num)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := &Tasks{
		Tasks: make([]*Task, 0),
	}

	for rows.Next() {
		var task Task
		err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			return nil, err
		}
		tasks.Tasks = append(tasks.Tasks, &task)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (s *SchedulerRepo) SearchTasks(data string, num int) (*Tasks, error) {
	searchPattern := "%" + data + "%"
	rows, err := s.db.Query(
		"SELECT * FROM scheduler WHERE title LIKE ? OR comment LIKE ? ORDER BY date LIMIT ?;", searchPattern, searchPattern, num)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := &Tasks{
		Tasks: make([]*Task, 0),
	}

	for rows.Next() {
		var task Task
		err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			return nil, err
		}
		tasks.Tasks = append(tasks.Tasks, &task)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (s *SchedulerRepo) SearchTasksByDate(data string, num int) (*Tasks, error) {
	rows, err := s.db.Query(
		"SELECT * FROM scheduler WHERE date = ? LIMIT ?;", data, num)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := &Tasks{
		Tasks: make([]*Task, 0),
	}

	for rows.Next() {
		var task Task
		err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			return nil, err
		}
		tasks.Tasks = append(tasks.Tasks, &task)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (s *SchedulerRepo) GetTaskByID(id int) (*Task, error) {
	row := s.db.QueryRow(
		"SELECT * FROM scheduler WHERE id = ?;", id)

	var task Task

	err := row.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &task, nil
}

func (s *SchedulerRepo) UpdateTaskByID(t *Task) (error) {
	row, err := s.db.Exec(
		"UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?;", 
		t.Date, t.Title, t.Comment, t.Repeat, t.ID)
	if err != nil {
		return err
	}

	rowsAffected, err := row.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("task not found")
	}
	
	return nil
}

func (s *SchedulerRepo) DeleteTaskByID(id int) (error) {
	row, err := s.db.Exec(
		"DELETE FROM scheduler WHERE id = ?;", id)
	if err != nil {
		return err
	}
	
	rowsAffected, err := row.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("task not found")
	}
	
	return nil
}