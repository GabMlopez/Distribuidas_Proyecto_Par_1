package utils

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret []byte

func init() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "tu-clave-secreta-por-defecto-cambiar-en-produccion"
	}
	jwtSecret = []byte(secret)
}

// GenerarTokenAdmin crea token para administrador
func GenerarTokenAdmin(adminID, usuario string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"admin_id": adminID,
		"usuario":  usuario,
		"tipo":     "admin",
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
		"iat":      time.Now().Unix(),
	})
	return token.SignedString(jwtSecret)
}

// GenerarTokenUser crea token para usuario normal
func GenerarTokenUser(usuarioID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"usuario_id": usuarioID,
		"tipo":       "user",
		"exp":        time.Now().Add(time.Hour * 24).Unix(),
		"iat":        time.Now().Unix(),
	})
	return token.SignedString(jwtSecret)
}

// VerificarToken valida cualquier token
func VerificarToken(tokenString string) (*jwt.Token, jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return token, claims, nil
	}

	return nil, nil, jwt.ErrTokenInvalidClaims
}

// VerificarTokenAdmin valida un token de administrador
func VerificarTokenAdmin(tokenString string) (*jwt.Token, jwt.MapClaims, error) {
	return VerificarToken(tokenString)
}

// VerificarTokenUser valida un token de usuario normal
func VerificarTokenUser(tokenString string) (*jwt.Token, jwt.MapClaims, error) {
	return VerificarToken(tokenString)
}
