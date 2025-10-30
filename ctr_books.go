package main

import (
	"fmt"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

# EXAMPLE ROUTES, REMOVE FOR ACTUAL USE

# This struct would normally be defined elsewhere, such as in a models file/package. Defined here for simplicity (because this whole file is disposable).
type Book struct {
	ID     string
	Title  string
	Author string
	ISBN   string
}

func getMockBooks() []Book {
	return []Book{
		{ID: "1", Title: "The Go Programming Language", Author: "Alan A. A. Donovan", ISBN: "978-0134190440"},
		{ID: "2", Title: "Learning Go", Author: "Jon Bodner", ISBN: "978-1492077213"},
		{ID: "3", Title: "Concurrency in Go", Author: "Katherine Cox-Buday", ISBN: "978-1491941294"},
	}
}

func route_Books_Index() gin.HandlerFunc {
	return func(c *gin.Context) {
		dso := c.MustGet("dso").(*DataSourceOrchestration)
		logger := dso.Logger
		logger.Debug("calling route_Books_Index()")

		session := sessions.Default(c)
		user := getUser(session)
		flashes := getFlashes(session)

		books := getMockBooks()

		c.HTML(http.StatusOK, "books/index", struct {
			AppConfig   *AppConfig
			SessionUser *SessionUser
			Flash       []string
			Books       []Book
		}{
			dso.AppConfig,
			&user,
			flashes,
			books,
		})
	}
}

func route_Books_Show() gin.HandlerFunc {
	return func(c *gin.Context) {
		dso := c.MustGet("dso").(*DataSourceOrchestration)
		logger := dso.Logger
		logger.Debug("calling route_Books_Show()")

		session := sessions.Default(c)
		user := getUser(session)
		flashes := getFlashes(session)

		id := c.Param("id")
		books := getMockBooks()

		var book *Book
		for _, b := range books {
			if b.ID == id {
				book = &b
				break
			}
		}

		if book == nil {
			logger.Error("book not found", "id", id)
			addFlash("Book not found", session)
			c.Redirect(http.StatusSeeOther, "/books")
			return
		}

		c.HTML(http.StatusOK, "books/show", struct {
			AppConfig   *AppConfig
			SessionUser *SessionUser
			Flash       []string
			Book        *Book
		}{
			dso.AppConfig,
			&user,
			flashes,
			book,
		})
	}
}

func route_Books_New() gin.HandlerFunc {
	return func(c *gin.Context) {
		dso := c.MustGet("dso").(*DataSourceOrchestration)
		logger := dso.Logger
		logger.Debug("calling route_Books_New()")

		session := sessions.Default(c)
		user := getUser(session)
		flashes := getFlashes(session)

		c.HTML(http.StatusOK, "books/new", struct {
			AppConfig   *AppConfig
			SessionUser *SessionUser
			Flash       []string
		}{
			dso.AppConfig,
			&user,
			flashes,
		})
	}
}

func route_Books_Create_POST() gin.HandlerFunc {
	return func(c *gin.Context) {
		dso := c.MustGet("dso").(*DataSourceOrchestration)
		logger := dso.Logger
		logger.Debug("calling route_Books_Create_POST()")

		session := sessions.Default(c)

		title := c.PostForm("title")
		author := c.PostForm("author")
		isbn := c.PostForm("isbn")

		if title == "" || author == "" || isbn == "" {
			logger.Error("validation failed: missing required fields")
			addFlash("All fields are required", session)
			c.Redirect(http.StatusSeeOther, "/books/new")
			return
		}

		logger.Debug("book created successfully", "title", title, "author", author, "isbn", isbn)
		addFlash(fmt.Sprintf("Book '%s' created successfully", title), session)
		c.Redirect(http.StatusSeeOther, "/books")
	}
}
