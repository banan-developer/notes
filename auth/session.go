package auth

import (
	"net/http"

	"github.com/gorilla/sessions"
)

// создание сессии
var Store = sessions.NewCookieStore([]byte("key!!!!"))

const sessionName = "notes-session"
const userIDKey = "user_id"

// кидаем айди пользователя в куки
func SetUserID(w http.ResponseWriter, r *http.Request, userID int) error {
	sessions, _ := Store.Get(r, sessionName)
	sessions.Values[userIDKey] = userID
	return sessions.Save(r, w)

}

// получем айди
func GetUserId(r *http.Request) (int, bool) {
	sessions, err := Store.Get(r, sessionName)
	if err != nil {
		return 0, false
	}
	id, ok := sessions.Values[userIDKey].(int)
	return id, ok
}

// выход из сессии
func СlearSessions(w http.ResponseWriter, r *http.Request) error {
	sessions, err := Store.Get(r, sessionName)
	if err != nil {
		return err
	}
	sessions.Options.MaxAge = -1

	return sessions.Save(r, w)
}
