package main

import (
	"io"
	"log/slog"
	"os"
)

// SetupLogger initializes and configures the slog logger based on AppConfig settings
func SetupLogger(logLevel int, logFile string) *slog.Logger {
	// Map log level (1-5) to slog.Level
	var level slog.Level
	switch logLevel {
	case 1:
		level = slog.LevelError
	case 2:
		level = slog.LevelWarn
	case 4, 5: // Map both Debug and Trace to Debug (slog doesn't have Trace)
		level = slog.LevelDebug
	default: // Default to Info
		level = slog.LevelInfo
	}

	// Determine output destination
	var output io.Writer
	if logFile == "" {
		output = os.Stdout
	} else {
		logf, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			// Can't use logger yet, use stderr
			slog.Error("Error opening log file", "file", logFile, "error", err)
			os.Exit(1)
		}
		output = logf
	}

	// Create handler options
	opts := &slog.HandlerOptions{
		Level: level,
	}

	// Create JSON handler for structured logging
	handler := slog.NewJSONHandler(output, opts)
	logger := slog.New(handler)

	// Set as default logger
	slog.SetDefault(logger)

	// Log destination
	if logFile == "" {
		logger.Info("No log_file specified in 'config.toml', logging to STDOUT")
	} else {
		logger.Info("Logging to file", "file", logFile)
	}

	return logger
}
