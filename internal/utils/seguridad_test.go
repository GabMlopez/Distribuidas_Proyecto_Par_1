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

func TestEncryptionDecryption(t *testing.T) {
	// Configurar una clave de prueba si no existe
	originalKey := aesKey
	defer func() { aesKey = originalKey }() // Restaurar después

	aesKey = []byte("12345678901234567890123456789012")

	testCases := []struct {
		name      string
		plaintext string
	}{
		{"Normal message", "Hola mundo"},
		{"Empty message", ""},
		{"Long message", "Este es un mensaje bastante más largo para probar que el cifrado GCM funciona correctamente con diferentes longitudes de entrada sin problemas."},
		{"Special characters", "¡Hola! ¿Cómo estás? 123 @#$%^&*()"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ciphertext := EncryptMessage(tc.plaintext)
			
			if tc.plaintext != "" && ciphertext == tc.plaintext {
				t.Errorf("EncryptMessage no cifró el mensaje: %s", tc.plaintext)
			}

			decrypted := DecryptMessage(ciphertext)
			if decrypted != tc.plaintext {
				t.Errorf("DecryptMessage falló. Esperado: %s, Obtenido: %s", tc.plaintext, decrypted)
			}
		})
	}
}

func TestEncryptionFallback(t *testing.T) {
	originalKey := aesKey
	defer func() { aesKey = originalKey }()

	// Probar con clave de tamaño incorrecto
	aesKey = []byte("short")
	
	plaintext := "Mensaje de prueba"
	ciphertext := EncryptMessage(plaintext)
	
	if ciphertext != plaintext {
		t.Errorf("EncryptMessage debería devolver el texto original si la clave es inválida")
	}

	decrypted := DecryptMessage(ciphertext)
	if decrypted != plaintext {
		t.Errorf("DecryptMessage debería devolver el texto original si la clave es inválida")
	}
}

func TestDecryptInvalidInput(t *testing.T) {
	// Configurar una clave de prueba if not set
	if len(aesKey) != 32 {
		aesKey = []byte("12345678901234567890123456789012")
	}

	invalidInputs := []string{
		"not-base64!",
		"SGVsbG8=", // "Hello" in base64, but too short for GCM nonce
		"",
	}

	for _, input := range invalidInputs {
		t.Run("Input: "+input, func(t *testing.T) {
			decrypted := DecryptMessage(input)
			// Debería devolver la entrada original si no puede descifrar (fallback legacy)
			if decrypted != input {
				t.Errorf("DecryptMessage debería devolver la entrada original para fallos. Esperado: %s, Obtenido: %s", input, decrypted)
			}
		})
	}
}
