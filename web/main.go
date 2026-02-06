package main

import (
	"database/sql"
	"log"
	"net/http"
	"notes/auth"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

// структура для хранения зависимостей (логирование и бд)
type application struct {
	errorLog *log.Logger
	infoLog  *log.Logger
	db       *sql.DB
}

// структура для хранения заметок
type Note struct {
	ID      int    `json:"id"`
	Content string `json:"content"`
	UserID  int    `json:"user_id"`
	Time    string `json:"create_at"`
}

func main() {
	// создание файла для отлавливания ошибок
	f, err := os.OpenFile("info.log", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	infoLog := log.New(f, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(f, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	db, err := sql.Open("mysql", "root:ТУТ ПАРОЛЬ ОТ ВАШЕЙ БД?@tcp(127.0.0.1:3306)/notes_app")
	if err != nil {
		errorLog.Fatal("ошибка подлкючения к базе данных", err)
	}
	err = db.Ping()
	if err != nil {
		errorLog.Fatal("Error of connect", err)
	}

	app := &application{
		errorLog: errorLog,
		infoLog:  infoLog,
		db:       db,
	}

	infoLog.Println("Подключено к MySQL")

	mux := http.NewServeMux()
	mux.Handle("/api/notes/", auth.RequireAuth(http.HandlerFunc(app.notesHandler)))
	mux.Handle("/api/notes", auth.RequireAuth(http.HandlerFunc(app.notesHandler)))

	mux.HandleFunc("/", app.homeHandler)
	mux.HandleFunc("/register", app.regHandler)
	mux.HandleFunc("/login", app.autoresHandler)
	mux.HandleFunc("/exit", app.exitSession)

	// подключение стилей
	fileServer := http.FileServer(http.Dir("./pkg/ui/static/"))
	mux.Handle("/static/", http.StripPrefix("/static", fileServer))

	// запуск сервера
	infoLog.Printf("Запуск сервера на http://127.0.0.1:4000")
	err = http.ListenAndServe(":4000", mux)
	app.errorLog.Fatal(err)
}
