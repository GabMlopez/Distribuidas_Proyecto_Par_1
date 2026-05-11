package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"os"

	"golang.org/x/crypto/bcrypt"
)

var aesKey []byte

func init() {
	key := os.Getenv("AES_ENCRYPTION_KEY")
	if len(key) == 32 {
		aesKey = []byte(key)
	}
}

func Hash_contrasenia(contrasenia string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(contrasenia), bcrypt.DefaultCost)
	return string(bytes), err
}

func Check_contrasenia(contrasenia, hash string) error {
	if contrasenia == "" {
		return errors.New("la contraseña no puede estar vacía")
	}
	if hash == "" {
		return errors.New("el hash de contraseña no puede estar vacío")
	}

	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(contrasenia))
	if err != nil {
		return errors.New("contraseña incorrecta")
	}
	return nil
}

func EncryptMessage(plaintext string) string {
	if len(aesKey) != 32 {
		return plaintext // Fallback si no hay key
	}
	if plaintext == "" {
		return ""
	}

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return plaintext
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return plaintext
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return plaintext
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext)
}

func DecryptMessage(ciphertextBase64 string) string {
	if len(aesKey) != 32 || ciphertextBase64 == "" {
		return ciphertextBase64
	}

	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextBase64)
	if err != nil {
		// Probablemente no estaba encriptado (legacy)
		return ciphertextBase64
	}

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return ciphertextBase64
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return ciphertextBase64
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return ciphertextBase64
	}

	nonce, ciphertextBytes := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return ciphertextBase64
	}

	return string(plaintext)
}
