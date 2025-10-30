package main

import (
	"bytes"
	"fmt"
	"html/template"
	"log/slog"
	"time"

	"github.com/Masterminds/sprig/v3"
	"github.com/davecgh/go-spew/spew"
	"github.com/dustin/go-humanize"
	"github.com/gin-gonic/gin"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	mdhtml "github.com/yuin/goldmark/renderer/html"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gorm.io/gorm"
)

const datetimeFormat_MonthYear = "Jan 2006" // e.g., "Feb 2024"

// DataSourceOrchestration holds references to various data sources used in the application
// It can be used as a middleware to inject these data sources into the Gin context and make them available to route handlers.
// This approach is far preferred over using global variables.
// It also reduces churn when adding new data sources to the application and needing to pass them down.
type DataSourceOrchestration struct {
	AppConfig *AppConfig
	DB        *gorm.DB
	Logger    *slog.Logger
}

// mwAppConfig adds the AppConfig object as a middleware for the Gin context
// NOTE: This is an example of an alternative pattern for direct middleware access.
// Currently, AppConfig is accessible via the DSO (DataSourceOrchestration) pattern,
// which is the preferred approach. This function is kept as a reference but is not used.
// Uncomment and register in routes.go if you need direct AppConfig middleware access.
/*
func mwAppConfig(cfg *AppConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("AppConfig", cfg)
		c.Next()
	}
}
*/

// mwDSO adds the DataSourceOrchestration object as a middleware for the Gin context
func mwDSO(dso *DataSourceOrchestration) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("dso", dso)
		c.Next()
	}
}

// mwDatabase adds the Gorm DB object as a middleware for the Gin context
// NOTE: This is an example of an alternative pattern for direct middleware access.
// Currently, the database is accessible via the DSO (DataSourceOrchestration) pattern,
// which is the preferred approach. This function is kept as a reference but is not used.
// Uncomment and register in routes.go if you need direct database middleware access.
/*
func mwDatabase(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("Database", db)
		c.Next()
	}
}
*/

// customTmplFuncMap returns a template.FuncMap with custom functions for templates
func customTmplFuncMap(cfg *AppConfig, logger *slog.Logger) template.FuncMap {
	fm := sprig.FuncMap()
	fm["fdateyear"] = func(t *time.Time) string {
		if t == nil {
			return "-"
		}
		return t.Local().Format("2006")
	}
	fm["fdatemonthyear"] = func(t *time.Time) string {
		if t == nil {
			return "-"
		}
		return t.Local().Format(datetimeFormat_MonthYear)
	}
	fm["fdate"] = func(t time.Time) string {
		return t.Local().Format("01-02-2006")
	}
	fm["fdateutc"] = func(t time.Time) string {
		return t.UTC().Format("01-02-2006")
	}
	fm["fdatetime"] = func(t time.Time) string {
		return t.Local().Format("01-02-2006 03:04 PM")
	}
	fm["to_days"] = func(d time.Duration) int {
		return int(d.Hours() / 24)
	}
	fm["overdue"] = func(t time.Time) bool {
		return t.Before(time.Now())
	}
	fm["spew"] = spew.Sdump
	fm["truncate"] = func(n int, s string) string {
		// n := 50
		rs := []rune(s)
		if len(rs) <= n {
			return s
		}
		return string(rs[:n]) + "..."
	}
	fm["percent"] = func(num, denom int) string {
		return fmt.Sprintf("%0.1f", float64(num)/float64(denom)*100)
	}
	fm["booltoyn"] = func(b bool) string {
		if b {
			return "Yes"
		}
		return "No"
	}
	fm["to_title"] = func(s string) string {
		return cases.Title(language.English).String(s)
	}
	fm["markdown"] = func(s string) string {
		md := goldmark.New(
			goldmark.WithExtensions(extension.GFM),
			goldmark.WithParserOptions(
				parser.WithAutoHeadingID(),
			),
			goldmark.WithRendererOptions(
				mdhtml.WithHardWraps(),
				mdhtml.WithXHTML(),
			),
		)
		var buf bytes.Buffer
		if err := md.Convert([]byte(s), &buf); err != nil {
			logger.Error("error converting markdown", "error", err)
			return "error converting markdown"
		}
		return buf.String()
	}
	fm["unescapeHTML"] = func(s string) template.HTML {
		return template.HTML(s)
	}
	fm["unescapeURL"] = func(s string) template.URL {
		return template.URL(s)
	}
	fm["int_commafy"] = func(i int) string {
		return humanize.Comma(int64(i))
	}
	return fm
}
