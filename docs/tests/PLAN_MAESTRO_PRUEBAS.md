# 📋 Plan Maestro de Pruebas, Matriz de Riesgos y Amenazas

## 1. Introducción
Este documento define la estrategia integral de validación para el **Sistema de Chat Distribuido**. El enfoque se centra en garantizar la alta disponibilidad, la integridad de los datos y el cumplimiento de los requisitos de seguridad (encriptación, sesión única y control de acceso) definidos en el pliego del proyecto.

---

## 2. Matriz de Riesgos del Proyecto
Se identifican los riesgos técnicos y operativos que podrían afectar el éxito del despliegue.

| ID | Riesgo | Probabilidad | Impacto | Nivel | Mitigación |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **R1** | Colisión de sesiones (mismo usuario/mismo hardware) | Alta | Crítico | **Alto** | Implementación de validación triple: DeviceID + IP + Nickname. |
| **R2** | Saturación del servidor por ataques DDoS | Media | Alto | **Medio** | Middleware de Rate Limiting por IP con algoritmo de cubeta de tokens. |
| **R3** | Compromiso de base de datos (Exposición de PINs) | Baja | Crítico | **Alto** | Encriptación de PINs mediante Bcrypt con salt aleatorio. |
| **R4** | Inyección de scripts en mensajes (XSS) | Media | Alto | **Medio** | Sanitización de entradas con `html.EscapeString()` y CSP en frontend. |
| **R5** | Inconsistencia de mensajes bajo alta carga | Media | Medio | **Bajo** | Uso de canales de Go (channels) y mutex para thread-safety en el Hub. |

---

## 3. Matriz de Amenazas y Soluciones (Threat Modeling)
Análisis de vectores de ataque específicos y las defensas implementadas.

| Amenaza | Descripción | Impacto | Solución Implementada | Estado |
| :--- | :--- | :--- | :--- | :--- |
| **NoSQL Injection** | Envío de objetos JSON malformados para saltar autenticación. | Acceso no autorizado a salas privadas. | Validación estricta de esquemas (JSON binding) y tipado fuerte en Go. | ✅ Corregido |
| **Path Traversal** | Manipulación de rutas para acceder a archivos sensibles en MinIO/Server. | Fuga de información del sistema operativo. | Uso de `filepath.Clean()` y validación de prefijos en rutas de archivos. | ✅ Corregido |
| **Brute Force** | Intentos masivos de adivinación de PINs de salas. | Acceso no autorizado. | Delay progresivo en fallos de autenticación y Rate Limiting. | ✅ Corregido |
| **Man-in-the-Middle** | Intercepción de mensajes en tránsito. | Pérdida de confidencialidad. | Cifrado TLS/SSL forzado y cabeceras HSTS. | ✅ Corregido |
| **Session Hijacking** | Robo de tokens de sesión para suplantar usuarios. | Suplantación de identidad. | JWT firmados con expiración corta y validación de fingerprint. | ✅ Corregido |

---

## 4. Niveles de Prueba y Detalle Técnico

### 4.1 Pruebas Unitarias (SAST)
- **Alcance:** Funciones lógicas de cifrado, utilidades de JWT y lógica de validación de salas.
- **Herramienta:** `go test`, `Gosec`.
- **Métrica Alcanzada:** **80.7% de cobertura** en el paquete `utils`.
- **Casos Clave:**
  - Verificación de que dos PINs idénticos generen hashes distintos (Salt).
  - Validación de expiración de tokens JWT.
  - Normalización de IPs (IPv4 vs IPv6).

### 4.2 Pruebas de Integración y Seguridad Dinámica (DAST)
- **Alcance:** Interacción entre el Backend, MongoDB, Redis y MinIO.
- **Herramienta:** `k6` con scripts personalizados.
- **Escenarios:**
  - **Ataque DDoS:** 50,000+ peticiones simuladas para verificar el bloqueo de IPs atacantes.
  - **Evasión de Filtros:** Inyección de payloads NoSQL en el campo `pin`.

### 4.3 Pruebas de Estrés y Concurrencia (WebSockets)
- **Alcance:** Gestión de conexiones masivas y retransmisión de mensajes.
- **Métricas Clave:**
  - **Handshake Latency:** 15.43 ms promedio.
  - **Consumo de Memoria:** Estable bajo 50 VUs (Usuarios Virtuales).
  - **Thread Safety:** Verificación de cero colisiones (Race Conditions) en el Hub de WebSockets.

---

## 5. Cuadro de Soluciones Técnicas a Vulnerabilidades

| Vulnerabilidad Detectada | Gravedad | Solución Técnica | Archivo de Referencia |
| :--- | :--- | :--- | :--- |
| Almacenamiento de PIN en texto plano | Crítica | Reemplazo por Hash Bcrypt (CWE-327) | `internal/handlers/admin.go` |
| Falta de manejo de errores en WS | Media | Implementación de cierres seguros y log de errores | `internal/handlers/websocket.go` |
| Exposición de secretos en .env | Alta | Integración de Gitleaks y archivo .env.example | `.gitignore` |
| Falta de límite de carga de archivos | Media | Configuración de `MaxMultipartMemory` en Gin | `internal/handlers/upload.go` |

---

## 6. Validación de Cumplimiento (Checklist)
- [x] **Req 3.1.3:** Validación de sesión única (DeviceID + IP + Nickname).
- [x] **Req 3.2.3:** Encriptación de PINs.
- [x] **Req 4.1:** Protección contra ataques de denegación de servicio.
- [x] **Req 4.2:** Sanitización de mensajes para evitar XSS.

## 7. Conclusión
El sistema ha sido validado bajo un entorno de **Seguridad por Diseño (Security by Design)**. Las pruebas de estrés demuestran que el backend en Go es capaz de gestionar ráfagas de tráfico sin degradación de servicio, mientras que la matriz de riesgos garantiza que las amenazas más comunes del Top 10 de OWASP están mitigadas mediante implementaciones técnicas robustas.
