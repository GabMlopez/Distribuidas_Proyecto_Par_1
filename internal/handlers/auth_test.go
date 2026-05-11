package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestLoginHandler_InvalidData(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.POST("/auth/login", Login_handler)

	// Test Case 1: Malformed JSON
	t.Run("Malformed JSON", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString("{invalid json}"))
		resp := httptest.NewRecorder()
		r.ServeHTTP(resp, req)

		if resp.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", resp.Code)
		}
	})

	// Test Case 2: Attempting NoSQL Injection via JSON Object (if it were allowed)
	// Since LoginRequest.Usuario is a string, this should fail at the binding stage.
	t.Run("NoSQL Injection Attempt via Object", func(t *testing.T) {
		injectionData := map[string]interface{}{
			"usuario":     map[string]string{"$ne": "admin"},
			"contrasenia": "password",
		}
		body, _ := json.Marshal(injectionData)
		req, _ := http.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
		resp := httptest.NewRecorder()
		r.ServeHTTP(resp, req)

		// It should fail because "usuario" is expected to be a string, not a map
		if resp.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400 (Binding error), got %d", resp.Code)
		}
	})
}
