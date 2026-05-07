package utils

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

var jwtSecretAdmin []byte
var jwtSecretUser []byte

func init() {
	godotenv.Load()
	// Cargar secret key desde variable de entorno (UNA SOLA VEZ)
	secretAdmin := os.Getenv("JWT_SECRET_ADMIN")
	if secretAdmin == "" {
		println("Secret not found admin")
		return
	}
	jwtSecretAdmin = []byte(secretAdmin)

	secretUser := os.Getenv("JWT_SECRET_USER")
	if secretUser == "" {
		println("Secret not found user")
		return
	}
	jwtSecretUser = []byte(secretUser)
	jwtSecretAdmin = []byte(secretAdmin)
}

func GetJWTSecretUser() []byte {
	return jwtSecretUser
}

// GetJWTSecret retorna la clave secreta JWT
func GetJWTSecretAdmin() []byte {
	return jwtSecretAdmin
}

// GenerarToken crea un nuevo token JWT
func GenerarTokenAdmin(adminID, usuario string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"admin_id": adminID,
		"usuario":  usuario,
		"exp":      time.Now().Add(time.Hour * 24).Unix(), // 24 horas
		"iat":      time.Now().Unix(),
	})

	return token.SignedString(jwtSecretAdmin)
}

// VerificarToken valida y parsea un token JWT
func VerificarTokenAdmin(tokenString string) (*jwt.Token, jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Verificar método de firma
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return jwtSecretAdmin, nil
	})

	if err != nil {
		return nil, nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return token, claims, nil
	}

	return nil, nil, jwt.ErrTokenInvalidClaims
}

// GenerarToken crea un nuevo token JWT
func GenerarTokenUser(usuario string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"usuario": usuario,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
		"iat":     time.Now().Unix(),
	})

	return token.SignedString(jwtSecretUser)
}

// VerificarToken valida y parsea un token JWT
func VerificarTokenUser(tokenString string) (*jwt.Token, jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Verificar método de firma
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return jwtSecretUser, nil
	})

	if err != nil {
		return nil, nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return token, claims, nil
	}

	return nil, nil, jwt.ErrTokenInvalidClaims
}

