package main

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func route_Root_Index() gin.HandlerFunc {
	return func(c *gin.Context) {
		dso := c.MustGet("dso").(*DataSourceOrchestration)
		logger := dso.Logger
		logger.Debug("calling route_Root_Index()")

		session := sessions.Default(c)
		user := getUser(session)
		flashes := getFlashes(session)

		// db := c.MustGet("Database").(*gorm.DB)

		c.HTML(http.StatusOK, "root/index", struct {
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

func route_Root_Ping() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	}
}
