package utils

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

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
