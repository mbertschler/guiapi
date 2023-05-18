package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type SessionStorage interface {
	GetSession(id string) (*Session, error)
	SetSession(s *Session) error
}

const sessionCookie = "session"

func sessionFromContext(c *gin.Context) *Session {
	sess, ok := c.Keys["session"].(*Session)
	if ok && sess != nil {
		return sess
	}
	if !ok {
		log.Printf("bad session %#v", c.Keys["session"])
	}
	if sess == nil {
		sess = &Session{}
	}
	return sess
}

func (db *DB) sessionMiddleware(c *gin.Context) {
	cookie, err := c.Request.Cookie(sessionCookie)
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
		http.SetCookie(c.Writer, &http.Cookie{
			Name:     sessionCookie,
			Value:    sess.ID,
			Path:     "/",
			HttpOnly: true,
			Expires:  time.Now().Add(30 * 24 * time.Hour),
		})
	}
	c.Keys = map[string]interface{}{
		"session": sess,
	}
	c.Next()
	err = db.SetSession(sess)
	if err != nil {
		log.Println("sessionMiddleware.SetSession:", err)
	}
}
