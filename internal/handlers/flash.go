package handlers

import (
	"net/http"

	"github.com/gorilla/sessions"
)

const flashKey = "flash_message"

func setFlash(w http.ResponseWriter, r *http.Request, store *sessions.CookieStore, sessionName, msg string) {
	session, _ := store.Get(r, sessionName)
	session.Values[flashKey] = msg
	_ = session.Save(r, w)
}

func popFlash(w http.ResponseWriter, r *http.Request, store *sessions.CookieStore, sessionName string) string {
	session, _ := store.Get(r, sessionName)
	v, ok := session.Values[flashKey]
	if !ok {
		return ""
	}

	msg, _ := v.(string)
	delete(session.Values, flashKey)
	_ = session.Save(r, w)
	return msg
}
