package main

import (
	"net/http"
	"notes/auth"

	"golang.org/x/crypto/bcrypt"
)

// функция для хэширования пароля
func hashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
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

		HashError := bcrypt.CompareHashAndPassword(
			[]byte(PasswordFrombd),
			[]byte(password),
		)

		if HashError == nil {
			auth.SetUserID(w, r, UserId)
		} else {
			// http.Error(w, "Неверный логин или пароль", http.StatusUnauthorized)
			http.Redirect(w, r, "/login#error1", http.StatusSeeOther)
		}

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

	// внесения данных в бд
	hashedPassword, _ := hashPassword(password)
	_, err := app.db.Exec("INSERT INTO users (login, password) VALUES (?, ?)", email, hashedPassword)

	if err != nil {
		app.errorLog.Println("REGISTER ERROR:", err)
		// http.Error(w, "Пользователь с таким email уже существует", http.StatusInternalServerError)
		http.Redirect(w, r, "/register#error2", http.StatusSeeOther)
	}

	// если все хорошо, то пользователь создан и перех к авторизации
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// выход из аккаунта
func (app *application) exitSession(w http.ResponseWriter, r *http.Request) {
	auth.ClearSessions(w, r)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
