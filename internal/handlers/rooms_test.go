package controladores

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGetRealIP(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name          string
		xForwardedFor string
		xRealIP       string
		clientIP      string
		expected      string
	}{
		{
			name:          "X-Forwarded-For present",
			xForwardedFor: "192.168.1.1, 10.0.0.1",
			xRealIP:       "",
			clientIP:      "127.0.0.1",
			expected:      "192.168.1.1",
		},
		{
			name:          "X-Real-IP present",
			xForwardedFor: "",
			xRealIP:       "10.0.0.2",
			clientIP:      "127.0.0.1",
			expected:      "10.0.0.2",
		},
		{
			name:          "Fallback to ClientIP",
			xForwardedFor: "",
			xRealIP:       "",
			clientIP:      "192.168.1.100",
			expected:      "192.168.1.100",
		},
		{
			name:          "Localhost IPv6 fallback",
			xForwardedFor: "",
			xRealIP:       "",
			clientIP:      "::1",
			expected:      "127.0.0.1",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			
			// Mocking request
			c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
			if tc.xForwardedFor != "" {
				c.Request.Header.Set("X-Forwarded-For", tc.xForwardedFor)
			}
			if tc.xRealIP != "" {
				c.Request.Header.Set("X-Real-IP", tc.xRealIP)
			}
			if tc.clientIP == "::1" {
				c.Request.RemoteAddr = "[" + tc.clientIP + "]:1234"
			} else {
				c.Request.RemoteAddr = tc.clientIP + ":1234"
			}

			result := getRealIP(c)
			if result != tc.expected {
				t.Errorf("Expected IP %s, got %s", tc.expected, result)
			}
		})
	}
}

func TestGenerateUserID(t *testing.T) {
	id1 := generateUserID()
	id2 := generateUserID()

	if len(id1) != 16 {
		t.Errorf("Expected length 16, got %d", len(id1))
	}
	if id1 == id2 {
		t.Errorf("Expected different IDs, got %s and %s", id1, id2)
	}
}

func TestCapacityError(t *testing.T) {
	err := &CapacityError{Message: "Sala llena"}
	if err.Error() != "Sala llena" {
		t.Errorf("Expected error message 'Sala llena', got '%s'", err.Error())
	}
}

func TestNicknameError(t *testing.T) {
	err := &NicknameError{Message: "Nickname duplicado"}
	if err.Error() != "Nickname duplicado" {
		t.Errorf("Expected error message 'Nickname duplicado', got '%s'", err.Error())
	}
}
