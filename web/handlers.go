package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

// домашний хэндлер
func homeHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./pkg/ui/html/index.html")

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
	// просто отправляем на фронт все заметки в json
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(notes)
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

	note.ID = nextID
	nextID++

	// добавляем ее в наш "бд"
	notes = append(notes, note)

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

	// удаление заметки из массива
	for i, note := range notes {
		if note.ID == id {
			notes = append(notes[:i], notes[i+1:]...)

			w.WriteHeader(http.StatusNoContent)
			return
		}
	}
	http.Error(w, "Note not found", http.StatusNotFound)
}
