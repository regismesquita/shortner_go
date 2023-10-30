package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCreateAlias(t *testing.T) {
	// Create a new Gin router
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.POST("/:alias", createAlias)

	// Create a sample request to use
	reqBody := strings.NewReader(`{"url":"https://example.com"}`)
	req, err := http.NewRequest(http.MethodPost, "/alias", reqBody)
	if err != nil {
		t.Fatalf("Couldn't create request: %v\n", err)
	}

	// Record the response
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetAlias(t *testing.T) {
	// Create a new Gin router
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	store.Set("alias", &URLData{URL: "https://example.com", Count: 0})
	r.GET("/:alias", getAlias)

	// Create a sample request to use
	req, err := http.NewRequest(http.MethodGet, "/alias", nil)
	if err != nil {
		t.Fatalf("Couldn't create request: %v\n", err)
	}

	// Record the response
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusMovedPermanently, w.Code)
	assert.Equal(t, "https://example.com", w.Header().Get("Location"))
	// Assert we increased the count
	assert.Equal(t, 1, store.urls["alias"].Count)
}

func TestListStats(t *testing.T) {
	// Create a new Gin router
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.GET("/stats", listStats)

	// Create a sample request to use
	req, err := http.NewRequest(http.MethodGet, "/stats", nil)
	if err != nil {
		t.Fatalf("Couldn't create request: %v\n", err)
	}

	// Record the response
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// JSON response you're expecting
	expectedJSON := `{"alias":{"URL":"https://example.com","Count":1}}`

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, expectedJSON, w.Body.String())
}
