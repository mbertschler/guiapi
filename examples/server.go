package main

import (
	"log"
	"net/http"
	"time"
)

const sessionCookie = "session"

func (db *DB) Session(w http.ResponseWriter, r *http.Request) *Session {
	cookie, err := r.Cookie(sessionCookie)
	if err != nil && err != http.ErrNoCookie {
		log.Println("sessionMiddleware.Cookie:", err)
	}
	id := ""
	if cookie != nil {
		id = cookie.Value
	}
	sess, err := db.GetSession(id)
	if err != nil {
		log.Println("sessionMiddleware.GetSession:", err)
	}
	if sess.New {
		http.SetCookie(w, &http.Cookie{
			Name:     sessionCookie,
			Value:    sess.ID,
			Path:     "/",
			HttpOnly: true,
			Expires:  time.Now().Add(30 * 24 * time.Hour),
		})
	}
	err = db.SetSession(sess)
	if err != nil {
		log.Println("sessionMiddleware.SetSession:", err)
	}
	return sess
}
