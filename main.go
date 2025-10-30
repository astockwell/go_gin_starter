package main

import (
	"embed"
	"fmt"
	"html/template"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	_ "github.com/joho/godotenv/autoload"
)

// DO NOT REMOVE: This triggers generated code. (One day we will also need: `go:embed public`)
//
//go:embed templates/**/*
var embeddedFiles embed.FS

func main() {
	appConfig, err := NewAppConfigFromFile("config")
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	// Setup structured logger as early as possible
	logger := SetupLogger(appConfig.LogLevel, appConfig.LogFile)
	logger.Info("Logger initialized", "level", appConfig.LogLevel)

	if appConfig.CacheTemplates {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize Gin router
	r := gin.New()
	r.Use(
		// Don't log requests to root '/' or ping, as the load balancers abuse them
		gin.LoggerWithWriter(gin.DefaultWriter, "/", "/ping"),
		gin.Recovery(),
	)

	// Load web view templates conditionally based on cache setting
	// (This enables us to load changed templates from disk on page refresh during development)
	if appConfig.CacheTemplates || os.Getenv("ENVIRONMENT") == "production" {
		// Build+serve a template set from embedded templates w/ our custom function map
		tmpl := template.Must(template.New("base").Funcs(customTmplFuncMap(appConfig, logger)).ParseFS(embeddedFiles, "templates/**/*.tmpl"))
		r.SetHTMLTemplate(tmpl)
	} else {
		// Add our custom template function map
		r.SetFuncMap(customTmplFuncMap(appConfig, logger))
		// Build+serve a template set from disk
		r.LoadHTMLGlob(filepath.Join(appConfig.WorkingDir, "templates/**/*"))
	}

	// // Serve embedded "public" files (there are none at the moment)
	// publicSub, err := fs.Sub(embeddedFiles, "public")
	// if err != nil {
	// 	logger.Error("fs.Sub(embeddedFiles, \"public\") failed", "error", err)
	// 	os.Exit(1)
	// }
	// r.StaticFS("/public", http.FS(publicSub))

	sesh := instantiateSessionStore(appConfig)
	r.Use(sessions.Sessions("mysession", sesh))

	// Create DSO with logger (you would also add DB connections, etc here)
	dso := &DataSourceOrchestration{
		AppConfig: appConfig,
		Logger:    logger,
	}

	// Add DSO middleware to make it (and it's conns) available to all routes/handlers
	r.Use(mwDSO(dso))

	// Register routes
	register_routes(r)

	// Start server (blocks indefinitely)
	if appConfig.SSLDisabled {
		logger.Info(fmt.Sprintf("HTTP Web server (no TLS) listening on http://localhost%s", appConfig.HostPort), "host_port", appConfig.HostPort)
		if err := r.Run(appConfig.HostPort); err != nil {
			logger.Error("Server failed", "error", err)
			os.Exit(1)
		}
	} else {
		logger.Info("Starting TLS server", "host_port", appConfig.HostPort, "cert", appConfig.SSLCertFile, "key", appConfig.SSLKeyFile)
		if err := r.RunTLS(appConfig.HostPort, appConfig.SSLCertFile, appConfig.SSLKeyFile); err != nil {
			logger.Error("TLS server failed", "error", err)
			os.Exit(1)
		}
	}
}
