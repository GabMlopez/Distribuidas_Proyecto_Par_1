package utils

import (
	"os"
	"testing"
)

func TestJWTGeneracionYVerificacion(t *testing.T) {
	// Set test secrets directly since init() might fail if .env is missing
	os.Setenv("JWT_SECRET_ADMIN", "test_admin_secret_123")
	os.Setenv("JWT_SECRET_USER", "test_user_secret_456")
	
	// manually set the vars since init() already ran
	jwtSecretAdmin = []byte("test_admin_secret_123")
	jwtSecretUser = []byte("test_user_secret_456")

	t.Run("Test Admin JWT", func(t *testing.T) {
		adminID := "admin123"
		usuario := "admin_test"

		tokenStr, err := GenerarTokenAdmin(adminID, usuario)
		if err != nil {
			t.Fatalf("Fallo al generar token admin: %v", err)
		}

		if tokenStr == "" {
			t.Errorf("El token de admin generado está vacío")
		}

		_, claims, err := VerificarTokenAdmin(tokenStr)
		if err != nil {
			t.Fatalf("Fallo al verificar token admin válido: %v", err)
		}

		if claims["admin_id"] != adminID {
			t.Errorf("Esperado admin_id %s, obtenido %v", adminID, claims["admin_id"])
		}
		if claims["usuario"] != usuario {
			t.Errorf("Esperado usuario %s, obtenido %v", usuario, claims["usuario"])
		}
	})

	t.Run("Test User JWT", func(t *testing.T) {
		usuario := "user_test"

		tokenStr, err := GenerarTokenUser(usuario)
		if err != nil {
			t.Fatalf("Fallo al generar token user: %v", err)
		}

		if tokenStr == "" {
			t.Errorf("El token de user generado está vacío")
		}

		_, claims, err := VerificarTokenUser(tokenStr)
		if err != nil {
			t.Fatalf("Fallo al verificar token user válido: %v", err)
		}

		if claims["usuario"] != usuario {
			t.Errorf("Esperado usuario %s, obtenido %v", usuario, claims["usuario"])
		}
	})

	t.Run("Test JWT Inválido Admin", func(t *testing.T) {
		// Valid user token passed to Admin verification
		tokenStr, _ := GenerarTokenUser("user_test")
		_, _, err := VerificarTokenAdmin(tokenStr)
		if err == nil {
			t.Errorf("VerificarTokenAdmin debió fallar con un token firmado con un secreto distinto")
		}
	})

	t.Run("Test JWT Inválido User", func(t *testing.T) {
		// Valid admin token passed to User verification
		tokenStr, _ := GenerarTokenAdmin("id", "admin_test")
		_, _, err := VerificarTokenUser(tokenStr)
		if err == nil {
			t.Errorf("VerificarTokenUser debió fallar con un token firmado con un secreto distinto")
		}
	})

	t.Run("Test Get Secrets", func(t *testing.T) {
		secAdmin := GetJWTSecretAdmin()
		if string(secAdmin) != "test_admin_secret_123" {
			t.Errorf("Secreto admin incorrecto: %s", string(secAdmin))
		}
		secUser := GetJWTSecretUser()
		if string(secUser) != "test_user_secret_456" {
			t.Errorf("Secreto user incorrecto: %s", string(secUser))
		}
	})
}
