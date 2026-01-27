package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

// структура для хранения заметок
type Note struct {
	ID      int    `json:"id"`
	Content string `json:"content"`
	UserID  int    `json:"user_id"`
	Time    string `json:"create_at"`
}

var db *sql.DB

func main() {
	var err error
	db, err = sql.Open("mysql", "root:NeeGan4562!?@tcp(127.0.0.1:3306)/notes_app")
	if err != nil {
		log.Fatal("ошибка подлкючения к базе данных", err)
	}
	err = db.Ping()
	if err != nil {
		log.Fatal("Error of connect", err)
	}

	fmt.Println("Подключено к MySQL")

	mux := http.NewServeMux()
	mux.HandleFunc("/api/notes/", notesHandler)
	mux.HandleFunc("/api/notes", notesHandler)

	mux.HandleFunc("/", homeHandler)
	mux.HandleFunc("/register", regHandler)
	mux.HandleFunc("/login", autoresHandler)

	// подключение стилей
	fileServer := http.FileServer(http.Dir("./pkg/ui/static/"))
	mux.Handle("/static/", http.StripPrefix("/static", fileServer))

	// запуск сервера
	log.Println("Запуск сервера на http://127.0.0.1:4000")
	err = http.ListenAndServe(":4000", mux)
	if err != nil {
		log.Fatal(err)
	}
}
