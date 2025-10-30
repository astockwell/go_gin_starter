package main

import (
	"encoding/gob"
	"fmt"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/now"
)

type SessionUser struct {
	Username  string
	DN        string
	FirstName string
	LastName  string
	Email     string

	Authenticated  bool
	AuthTime       time.Time
	AuthExpiration time.Time
	Role           uint64

	// IsOauth        bool
	// OauthSessionID string
}

// SessionUser.Role (uint64) Permission Bits:
const (
	SESSUSR__ADMIN = 1 << iota // Overall application admin
	SESSUSR__USER              // Regular user
)

// NewAuthenticatedSessionUser creates a new session user with a valid session flag/timestamp set
func NewAuthenticatedSessionUser(username string) *SessionUser {
	return &SessionUser{
		Username: username,
		// DN:       strings.ToLower(results[0].DN),

		Authenticated:  true,
		AuthTime:       time.Now(),
		AuthExpiration: AuthExpirationTime(),

		// FirstName:    results[0].GetAttributeValue("givenName"),
		// LastName:     results[0].GetAttributeValue("sn"),
	}
}

// Instantiate secure session store
func instantiateSessionStore(cfg *AppConfig) cookie.Store {
	store := cookie.NewStore(cfg.SecureCookieSigningKey, cfg.SecureCookieEncryptionKey)

	// Secure sessions
	store.Options(sessions.Options{
		Path: "/",
		// Domain: "",
		MaxAge:   cfg.SecureCookieMaxAge, // 86400 * 7
		Secure:   !cfg.SSLDisabled,
		HttpOnly: true,
	})

	// Register the SessionUser{} type to be serialized for inclusion in our sessions via `gob` encoding
	gob.Register(&SessionUser{})

	return store
}

func (s *SessionUser) SessionIsValid() bool {
	if !s.Authenticated {
		// fmt.Println("Session is not valid for", s.Username, "because s.Authenticated is false")
		return false
	}
	if time.Now().After(s.AuthExpiration) {
		// fmt.Println("Session is not valid for", s.Username, "because s.AuthExpiration has elapsed")
		return false
	}
	return true
}

func (s *SessionUser) AddRole(b uint64) {
	s.Role = s.Role | b
}

func (s *SessionUser) RemoveRole(b uint64) {
	s.Role = s.Role &^ b
}

func (s *SessionUser) IsRole(b uint64) bool {
	return ((s.Role & b) != 0)
}

// Convenience method to test for admin
func (s *SessionUser) IsAdmin() bool {
	return s.IsRole(SESSUSR__ADMIN)
}

func AuthExpirationTime() time.Time {
	return now.EndOfDay().Add(2 * time.Hour) // 2 AM
}

func addFlash(msg string, session sessions.Session) {
	session.AddFlash(msg)
	errsess := session.Save()
	if errsess != nil {
		panic(errsess.Error())
	}
}

func getFlashes(session sessions.Session) []string {
	fs := []string{}
	if flashes := session.Flashes(); len(flashes) > 0 {
		for i := 0; i < len(flashes); i++ {
			fs = append(fs, fmt.Sprintf("%v", flashes[i]))
		}
	}
	errsess := session.Save()
	if errsess != nil {
		panic(errsess.Error())
	}
	return fs
}

func getUser(session sessions.Session) SessionUser {
	// Retrieve our struct and type-assert it
	val := session.Get(gin.AuthUserKey)
	var user = &SessionUser{}
	var ok bool
	if user, ok = val.(*SessionUser); !ok {
		return SessionUser{}
	}
	return *user
}
