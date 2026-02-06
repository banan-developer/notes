package main

import (
	"encoding/json"
	"net/http"
	"notes/auth"
	"strconv"
	"strings"
)

// домашний хэндлер
func (app *application) homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, "./pkg/ui/html/index.html")
}

// авторизация пользователя
func (app *application) autoresHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		http.ServeFile(w, r, "./pkg/ui/html/auto.html")
		return
	}

	// получения значения в input через поля
	if r.Method == http.MethodPost {
		email := r.FormValue("email")
		password := r.FormValue("password")

		if email == "" || password == "" {
			http.Error(w, "email and password required", http.StatusBadRequest)
			return
		}

		var UserId int
		var PasswordFrombd string

		// получаем данные из бд а потом сравниваем пароль из базы данных и написанным в input
		rows, err := app.db.Query(
			"SELECT id, password FROM users WHERE login = ?",
			email,
		)
		if err != nil {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}
		defer rows.Close()

		if !rows.Next() {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}

		err = rows.Scan(&UserId, &PasswordFrombd)
		if err != nil {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}

		if password != PasswordFrombd {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}

		auth.SetUserID(w, r, UserId)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

}

// регистрация
func (app *application) regHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		http.ServeFile(w, r, "./pkg/ui/html/registration.html")
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// получение значения в input через поля
	email := r.FormValue("email")
	password := r.FormValue("password")

	if email == "" || password == "" {
		http.Error(w, "email and password required", http.StatusBadRequest)
		return
	}
	// внесения данных в бд
	_, err := app.db.Exec("INSERT INTO users (login, password) VALUES (?, ?)", email, password)

	if err != nil {
		app.errorLog.Println("REGISTER ERROR:", err)
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}

	// если все хорошо, то пользователь создан и перех к авторизации
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// выход из сессии
func (app *application) exitSession(w http.ResponseWriter, r *http.Request) {
	auth.СlearSessions(w, r)
	http.ServeFile(w, r, "./pkg/ui/html/index.html")
}

// главный хэндлер(get,post и delete в одном хэндлере)
func (app *application) notesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		app.getNote(w, r)
	case http.MethodPost:
		app.createNote(w, r)
	case http.MethodDelete:
		app.deleteNote(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// функция для get запросов
func (app *application) getNote(w http.ResponseWriter, r *http.Request) {
	UserID, _ := auth.GetUserId(r)
	rows, err := app.db.Query(
		"SELECT id, content, user_id FROM notes WHERE user_id = ?", UserID,
	)

	if err != nil {
		app.errorLog.Println("DB QUERY ERROR:", err)
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
func (app *application) createNote(w http.ResponseWriter, r *http.Request) {
	var note Note
	// получаем заметку
	err := json.NewDecoder(r.Body).Decode(&note)

	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// получение айди пользователя
	UserID, ok := auth.GetUserId(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	note.UserID = UserID

	result, err := app.db.Exec("INSERT INTO notes (content, user_id) VALUES (?, ?)", note.Content, note.UserID)

	if err != nil {
		app.errorLog.Println("DB INSERT ERROR:", err)
		app.errorLog.Println("MYSQL INSERT ERROR:", err)
		return
	}

	// забирем id из result
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
func (app *application) deleteNote(w http.ResponseWriter, r *http.Request) {

	// вытаскиваем id из url
	idStr := strings.TrimPrefix(r.URL.Path, "/api/notes/")

	// превращаем айди в число
	id, err := strconv.Atoi(idStr)
	if err != nil {
		app.errorLog.Println("DB DELETE ERROR:", err)
		http.Error(w, "Invalid note id", http.StatusBadRequest)
		return
	}

	UserId, _ := auth.GetUserId(r)

	_, err = app.db.Exec("DELETE FROM notes WHERE id = ? AND user_id = ?",
		id, UserId,
	)

	if err != nil {
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
