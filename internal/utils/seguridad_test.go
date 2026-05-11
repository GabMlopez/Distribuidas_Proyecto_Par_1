package utils

import (
	"testing"
)

func TestHashAndCheckContrasenia(t *testing.T) {
	contrasenia := "misuperclave123"

	// Probar Hash_contrasenia
	hash, err := Hash_contrasenia(contrasenia)
	if err != nil {
		t.Fatalf("Error al hashear contraseña: %v", err)
	}

	if hash == "" {
		t.Errorf("El hash generado está vacío")
	}

	if hash == contrasenia {
		t.Errorf("El hash no debe ser igual a la contraseña en texto plano")
	}

	// Probar Check_contrasenia - Éxito
	err = Check_contrasenia(contrasenia, hash)
	if err != nil {
		t.Errorf("Check_contrasenia falló para una contraseña correcta: %v", err)
	}

	// Probar Check_contrasenia - Fallo con clave incorrecta
	err = Check_contrasenia("claveIncorrecta", hash)
	if err == nil {
		t.Errorf("Check_contrasenia debió fallar con contraseña incorrecta")
	} else if err.Error() != "contraseña incorrecta" {
		t.Errorf("Mensaje de error esperado 'contraseña incorrecta', obtenido: %v", err)
	}

	// Probar Check_contrasenia - Contraseña vacía
	err = Check_contrasenia("", hash)
	if err == nil {
		t.Errorf("Check_contrasenia debió fallar con contraseña vacía")
	} else if err.Error() != "la contraseña no puede estar vacía" {
		t.Errorf("Mensaje de error esperado 'la contraseña no puede estar vacía', obtenido: %v", err)
	}

	// Probar Check_contrasenia - Hash vacío
	err = Check_contrasenia(contrasenia, "")
	if err == nil {
		t.Errorf("Check_contrasenia debió fallar con hash vacío")
	} else if err.Error() != "el hash de contraseña no puede estar vacío" {
		t.Errorf("Mensaje de error esperado 'el hash de contraseña no puede estar vacío', obtenido: %v", err)
	}
}
