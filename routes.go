package main

import (
	"github.com/gin-gonic/gin"
)

func register_routes(r *gin.Engine) {
	// Serve the homepage
	r.GET("/", route_Root_Index())
	r.GET("/ping", route_Root_Ping())

	// Books routes
	r.GET("/books", route_Books_Index())
	r.GET("/books/:id", route_Books_Show())
	r.GET("/books/new", route_Books_New())
	r.POST("/books", route_Books_Create_POST())
}
