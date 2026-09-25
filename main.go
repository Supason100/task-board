package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

type Task struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Phone       string `json:"phone"`
	Category    string `json:"category"`
	Description string `json:"description"`
	Location    string `json:"location"`
	Landmark    string `json:"landmark"`
	Status      string `json:"status"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	CreatedAt   string `json:"created_at"`
}

var db *sql.DB

func main() {
	var err error
	db, err = sql.Open("sqlite3", "./tasks.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	createTable()

	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.HandleFunc("/tasks", tasksHandler)
	http.HandleFunc("/tasks/{id}/taken", markTaskTaken)
	http.HandleFunc("DELETE /tasks/{id}", deleteTask)
	http.HandleFunc("PUT /tasks/{id}", updateTask)

	log.Println("Server running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func createTable() {
	query := `
	CREATE TABLE IF NOT EXISTS tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT,
		phone TEXT,
		category TEXT,
		description TEXT,
		location TEXT,
		landmark TEXT,
		status TEXT DEFAULT 'open',
		latitude REAL,
		longitude REAL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`
	_, err := db.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}

func tasksHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		getTasks(w, r)
	case "POST":
		createTask(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func getTasks(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")

	var rows *sql.Rows
	var err error

	if category != "" {
		rows, err = db.Query("SELECT id, name, phone, category, description, location, landmark, status, latitude, longitude, created_at FROM tasks WHERE category = ? ORDER BY created_at DESC", category)
	} else {
		rows, err = db.Query("SELECT id, name, phone, category, description, location, landmark, status, latitude, longitude, created_at FROM tasks ORDER BY created_at DESC")
	}	

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		rows.Scan(&t.ID, &t.Name, &t.Phone, &t.Category, &t.Description, &t.Location, &t.Landmark, &t.Status, &t.Latitude, &t.Longitude, &t.CreatedAt)
		tasks = append(tasks, t)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

func createTask(w http.ResponseWriter, r *http.Request) {
	var t Task
	err := json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := db.Exec(
			"INSERT INTO tasks (name, phone, category, description, location, landmark, status, latitude, longitude) VALUES (?, ?, ?, ?, ?, ?, 'open', ?, ?)",
		t.Name, t.Phone, t.Category, t.Description, t.Location, t.Landmark, t.Latitude, t.Longitude,	
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id, _ := result.LastInsertId()
	t.ID = int(id)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(t)
}

func markTaskTaken(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	_, err := db.Exec("UPDATE tasks SET status = 'taken' WHERE id = ?", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "taken"})
}

func deleteTask(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	_, err := db.Exec("DELETE FROM tasks WHERE id = ?", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}

func updateTask(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var t Task
	err := json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err = db.Exec(
			"UPDATE tasks SET name = ?, phone = ?, category = ?, description = ?, location = ?, landmark = ?, latitude = ?, longitude = ? WHERE id = ?",
		t.Name, t.Phone, t.Category, t.Description, t.Location, t.Landmark, t.Latitude, t.Longitude, id,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}