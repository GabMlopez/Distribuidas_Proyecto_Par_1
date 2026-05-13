# Guía de Pruebas Automatizadas con k6

Este directorio contiene los scripts de pruebas de carga y seguridad diseñados para validar la resiliencia del Chat Distribuido. 

> **Requisito Previo:** Asegúrate de que tu backend de Go esté corriendo y que conozcas la IP local en la que está escuchando (usualmente configurada en tu `.env` o mostrada por `run_local.bat`, ej. `192.168.100.72`).

Para ejecutar cualquiera de las pruebas, debes usar tu terminal (recomendado WSL Ubuntu si estás en Windows) y estar situado en la raíz del proyecto.

---

## 1. Prueba de Umbral de Rate Limiting (`threshold_test.js`)
**¿Qué demuestra?** Valida la "inteligencia" del Rate Limiter. Comprueba que el sistema permite tráfico normal (tiempos razonables) pero bloquea inmediatamente cuando las mismas IPs intentan realizar un ataque DDoS (tiempos de denegación).

**Comando de ejecución:**
```bash
k6 run -e API_IP=192.168.100.72 tests/k6/threshold_test.js
```

---

## 2. Prueba de Seguridad Pura (`security_test.js`)
**¿Qué demuestra?** Ejecuta un ataque de fuerza bruta masivo simulando una Botnet (50 IPs falsificadas atacando en paralelo a la máxima velocidad posible) para comprobar el bloqueo perimetral devolviendo `HTTP 429`.

**Comando de ejecución:**
```bash
k6 run -e API_IP=192.168.100.72 tests/k6/security_test.js
```

---

## 3. Prueba de Estrés de WebSockets (`stress_test.js`)
**¿Qué demuestra?** Somete al sistema a una carga máxima de comunicaciones en tiempo real. 50 Usuarios Virtuales establecen el *handshake* de WebSockets, mantienen las conexiones TCP vivas y envían ráfagas de mensajes simultáneamente para comprobar el manejo de concurrencia y la estabilidad de las Goroutines (Thread-Safety).

**Comando de ejecución:**
```bash
k6 run -e API_IP=192.168.100.72 tests/k6/stress_test.js
```

---

### Notas Importantes:
- **Cambio de IP:** Si tu red cambia de IP, simplemente actualiza el valor de `API_IP=` en los comandos anteriores.
- Si algún script falla con el error *101 Switching Protocols* o te da error 403, reinicia tu backend en Go cerrando la ventana CMD completamente y volviendo a ejecutar `.\run_local.bat` para asegurar que las excepciones de pruebas (`isK6Test`) estén compiladas en la memoria.
