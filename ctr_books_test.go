package main

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/securecookie"
)

// setupTestRouter creates a test Gin engine with routes and minimal DSO configuration
func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)

	r := gin.New()

	// Create minimal AppConfig for testing
	// SecureCookieSigningKey must be 64 bytes, encryption key 32 bytes
	appConfig := &AppConfig{
		SecureCookieSigningKey:    securecookie.GenerateRandomKey(64),
		SecureCookieEncryptionKey: securecookie.GenerateRandomKey(32),
		SecureCookieMaxAge:        3600,
		SSLDisabled:               true,
		LogLevel:                  1,
	}

	// Setup session store
	store := cookie.NewStore(appConfig.SecureCookieSigningKey, appConfig.SecureCookieEncryptionKey)
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   appConfig.SecureCookieMaxAge,
		Secure:   false,
		HttpOnly: true,
	})
	r.Use(sessions.Sessions("test-session", store))

	// Load templates for testing
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	r.SetFuncMap(customTmplFuncMap(appConfig, logger))
	r.LoadHTMLGlob("templates/**/*")

	// Create minimal DSO & inject it into middleware
	dso := &DataSourceOrchestration{
		AppConfig: appConfig,
		Logger:    logger,
		DB:        nil,
	}
	r.Use(mwDSO(dso))

	// Register routes
	register_routes(r)

	return r
}

// TestBooksShow tests the GET /books/:id route with multiple scenarios (table-driven)
func TestBooksShow(t *testing.T) {
	tests := []struct {
		name           string
		bookID         string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "ValidBookID1",
			bookID:         "1",
			expectedStatus: http.StatusOK,
			expectedBody:   "The Go Programming Language",
		},
		{
			name:           "ValidBookID2",
			bookID:         "2",
			expectedStatus: http.StatusOK,
			expectedBody:   "Learning Go",
		},
		{
			name:           "ValidBookID3",
			bookID:         "3",
			expectedStatus: http.StatusOK,
			expectedBody:   "Concurrency in Go",
		},
		{
			name:           "InvalidBookID",
			bookID:         "999",
			expectedStatus: http.StatusSeeOther,
			expectedBody:   "",
		},
		{
			name:           "NonNumericBookID",
			bookID:         "invalid",
			expectedStatus: http.StatusSeeOther,
			expectedBody:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupTestRouter()

			req, err := http.NewRequest("GET", "/books/"+tt.bookID, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Check status code
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// For valid book IDs, check that the book title appears in response
			if tt.expectedStatus == http.StatusOK && tt.expectedBody != "" {
				if !strings.Contains(w.Body.String(), tt.expectedBody) {
					t.Errorf("Expected response body to contain '%s'", tt.expectedBody)
				}
			}

			// For invalid book IDs, check redirect location
			if tt.expectedStatus == http.StatusSeeOther {
				location := w.Header().Get("Location")
				if location != "/books" {
					t.Errorf("Expected redirect to '/books', got '%s'", location)
				}
			}
		})
	}
}

// TestBooksCreate tests the POST /books route
func TestBooksCreate(t *testing.T) {
	t.Run("ValidFormSubmission", func(t *testing.T) {
		router := setupTestRouter()

		// Create form data
		form := url.Values{}
		form.Add("title", "Test Book")
		form.Add("author", "Test Author")
		form.Add("isbn", "978-1234567890")

		req, err := http.NewRequest("POST", "/books", strings.NewReader(form.Encode()))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Should redirect to /books after successful creation
		if w.Code != http.StatusSeeOther {
			t.Errorf("Expected status %d, got %d", http.StatusSeeOther, w.Code)
		}

		location := w.Header().Get("Location")
		if location != "/books" {
			t.Errorf("Expected redirect to '/books', got '%s'", location)
		}
	})

	t.Run("MissingTitle", func(t *testing.T) {
		router := setupTestRouter()

		form := url.Values{}
		form.Add("author", "Test Author")
		form.Add("isbn", "978-1234567890")

		req, err := http.NewRequest("POST", "/books", strings.NewReader(form.Encode()))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Should redirect to /books/new when validation fails
		if w.Code != http.StatusSeeOther {
			t.Errorf("Expected status %d, got %d", http.StatusSeeOther, w.Code)
		}

		location := w.Header().Get("Location")
		if location != "/books/new" {
			t.Errorf("Expected redirect to '/books/new', got '%s'", location)
		}
	})

	t.Run("MissingAuthor", func(t *testing.T) {
		router := setupTestRouter()

		form := url.Values{}
		form.Add("title", "Test Book")
		form.Add("isbn", "978-1234567890")

		req, err := http.NewRequest("POST", "/books", strings.NewReader(form.Encode()))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusSeeOther {
			t.Errorf("Expected status %d, got %d", http.StatusSeeOther, w.Code)
		}

		location := w.Header().Get("Location")
		if location != "/books/new" {
			t.Errorf("Expected redirect to '/books/new', got '%s'", location)
		}
	})

	t.Run("MissingISBN", func(t *testing.T) {
		router := setupTestRouter()

		form := url.Values{}
		form.Add("title", "Test Book")
		form.Add("author", "Test Author")

		req, err := http.NewRequest("POST", "/books", strings.NewReader(form.Encode()))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusSeeOther {
			t.Errorf("Expected status %d, got %d", http.StatusSeeOther, w.Code)
		}

		location := w.Header().Get("Location")
		if location != "/books/new" {
			t.Errorf("Expected redirect to '/books/new', got '%s'", location)
		}
	})

	t.Run("AllFieldsMissing", func(t *testing.T) {
		router := setupTestRouter()

		form := url.Values{}

		req, err := http.NewRequest("POST", "/books", strings.NewReader(form.Encode()))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusSeeOther {
			t.Errorf("Expected status %d, got %d", http.StatusSeeOther, w.Code)
		}

		location := w.Header().Get("Location")
		if location != "/books/new" {
			t.Errorf("Expected redirect to '/books/new', got '%s'", location)
		}
	})
}

// TestBooksIndex tests the GET /books route
func TestBooksIndex(t *testing.T) {
	t.Run("ReturnsBooksList", func(t *testing.T) {
		router := setupTestRouter()

		req, err := http.NewRequest("GET", "/books", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Should return 200 OK
		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		// Check that response contains book titles from mock data
		body := w.Body.String()
		expectedBooks := []string{
			"The Go Programming Language",
			"Learning Go",
			"Concurrency in Go",
		}

		for _, bookTitle := range expectedBooks {
			if !strings.Contains(body, bookTitle) {
				t.Errorf("Expected response to contain '%s'", bookTitle)
			}
		}
	})
}
