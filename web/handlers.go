package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// домашний хэндлер
func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, "./pkg/ui/html/index.html")
}

func autoresHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./pkg/ui/html/auto.html")
}

func regHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./pkg/ui/html/registration.html")
}

// главный хэндлер(get,post и delete в одном хэндлере)
func notesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getNote(w, r)
	case http.MethodPost:
		createNote(w, r)
	case http.MethodDelete:
		deleteNote(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// функция для get запросов
func getNote(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(
		"SELECT id, content, user_id FROM notes WHERE user_id = 1",
	)

	if err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}

	defer rows.Close()

	var notes []Note

	for rows.Next() {
		var note Note
		rows.Scan(&note.ID, &note.Content, &note.UserID)
		notes = append(notes, note)
	}

	if notes == nil {
		notes = []Note{}
	}

	// просто отправляем на фронт все заметки в json
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(notes)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// функция для post запросов
func createNote(w http.ResponseWriter, r *http.Request) {
	var note Note
	// получаем заметку
	err := json.NewDecoder(r.Body).Decode(&note)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	note.UserID = 1

	result, err := db.Exec("INSERT INTO notes (content, user_id, created_at) VALUES (?, ?)", note.Content, note.UserID)

	if err != nil {
		log.Println("MYSQL INSERT ERROR:", err)
		return
	}

	id, _ := result.LastInsertId()
	note.ID = int(id)

	// отправляем обратно на фронт
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	err = json.NewEncoder(w).Encode(note)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// функция для delete запросов
func deleteNote(w http.ResponseWriter, r *http.Request) {

	// вытаскиваем id из url
	idStr := strings.TrimPrefix(r.URL.Path, "/api/notes/")

	// превращаем айди в число
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid note id", http.StatusBadRequest)
		return
	}

	_, err = db.Exec("DELETE FROM notes WHERE id = ? AND user_id = 1",
		id,
	)

	if err != nil {
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
