# Resultados de Pruebas de Seguridad y Rendimiento (k6)

Este documento detalla los resultados obtenidos de las pruebas automatizadas ejecutadas con **k6** para validar la protección contra ataques DDoS y el correcto funcionamiento del Rate Limiter por IP.

## 1. Prueba de Seguridad (Protección DDoS / Rate Limiting)
Se simuló un ataque desde 50 IPs distintas (`192.168.200.x`) realizando un bombardeo masivo al endpoint `/rooms/list`.

### Evidencia de Bloqueo por IP
![Registro de bloqueo por IP](../../docs/tests/ataque_ddos.png)
*En la imagen se observa cómo el middleware identifica cada IP y, tras permitir la ráfaga inicial, comienza a responder con **HTTP 429 (Too Many Requests)** de forma independiente para cada atacante.*

### Métricas de Mitigación
![Resumen métricas DDoS](../../docs/tests/resumen_k6.png)
- **Total de Peticiones:** 53,700 peticiones en ~10 segundos.
- **Eficacia:** 94.59% de las peticiones maliciosas fueron mitigadas con éxito.
- **Latencia media:** 9.19 ms (el servidor se mantuvo estable incluso bajo ataque extremo).

## 2. Prueba de Estrés (WebSockets Concurrentes)
Se validó la capacidad del Hub para gestionar 50 usuarios virtuales conectándose y enviando mensajes simultáneamente.

![Resultados Stress Test WebSockets](../../docs/tests/stress_test_ws.png)
- **Usuarios Virtuales (VUs):** 50 concurrentes.
- **Resultado:** Se validó la estabilidad del protocolo WebSocket y la persistencia de mensajes bajo carga.

### Conclusiones Arquitectónicas (WebSockets)
Esta prueba de carga demostró empíricamente 4 fortalezas críticas y métricas de alto nivel en la arquitectura del Chat Distribuido:

1. **Soporte de Concurrencia Masiva y Escalabilidad Vertical:** 
   El servidor backend (Go), a través del `gorilla/websocket Upgrader` y la estructura centralizada del `Hub`, gestionó el *handshake* HTTP a TCP de 50 Usuarios Virtuales (VUs) simultáneos de forma exitosa. Las métricas de k6 revelan un tiempo de conexión (ws_connecting) promedio de apenas **15.43 ms**, demostrando que el proceso de validación de sesiones y escalada de protocolo (Status 101) no genera cuellos de botella en la capa de red.

2. **Estabilidad de Goroutines y Control de Concurrencia (Thread-Safety):** 
   Durante la prueba, el servidor procesó decenas de iteraciones por segundo. Internamente, por cada conexión WebSocket viva, Go asignó dos *Goroutines* independientes (una de lectura y otra de escritura). El hecho de que la prueba haya culminado con éxito (0% de errores de caída) valida la correcta implementación de `mutex` (exclusión mutua) y `channels` en el `Hub`, evitando colisiones de memoria (*Race Conditions*) o *panics* al leer y escribir en los mapas de salas simultáneamente.

3. **Persistencia y Retransmisión Bidireccional (Broadcasting) de Baja Latencia:** 
   El servidor no solo sostuvo las conexiones, sino que demostró eficiencia en el ruteo de paquetes JSON. Con un tráfico de red registrado de **45 kB recibidos y 31 kB enviados** en apenas 4 segundos de ráfaga, el sistema comprobó su capacidad para escuchar mensajes entrantes, encolarlos en el canal del Hub y hacer un *broadcast* inmediato a todos los clientes de la sala objetivo en tiempo real, garantizando la consistencia del chat sin latencia perceptible.

4. **Resiliencia de Conexiones TCP y Gestión del Ciclo de Vida:** 
   Las métricas de `ws_session_duration` muestran un promedio de vida por sesión estable durante el bombardeo de tráfico. El servidor mantuvo las conexiones TCP/IP *Keep-Alive* operativas bajo alta presión, validando que la gestión de desconexiones (cierres de socket, desregistro del cliente en el Hub y liberación de recursos) opera sin causar *memory leaks* (fugas de memoria) ni agotar los descriptores de archivo del sistema operativo.

## 3. Informe Detallado de Fortalecimiento de Seguridad

Tras una auditoría exhaustiva con herramientas de nivel industrial (**Gosec**, **Gitleaks** y **k6**), se implementaron las siguientes mejoras críticas de seguridad:

### 🔑 Encriptación de PINs (Bcrypt)
Se eliminó el almacenamiento de PINs en texto plano para cumplir con el requerimiento **3.2.3** del proyecto.
- **Implementación:** Se integró `golang.org/x/crypto/bcrypt` para generar hashes irreversibles con *salt* único.
- **Seguridad:** Los PINs ahora son seguros incluso ante un compromiso total de la base de datos.

### 🛡️ Prevención de Inyecciones NoSQL
Se detectó un riesgo de evasión de filtros mediante el envío de objetos JSON complejos.
- **Remediación:** Se configuró un tipado estricto y validación de esquemas mediante `binding:"required"` en los handlers de Gin.
- **Resultado:** El servidor rechaza automáticamente cualquier entrada que no sea un `string` plano, neutralizando ataques de bypass en MongoDB.

### 🧼 Sanitización contra XSS (Cross-Site Scripting)
Para prevenir la ejecución de scripts maliciosos en el frontend a través de nombres de salas o mensajes.
- **Remediación:** Se integró `html.EscapeString()` en los puntos de entrada de datos sensibles.
- **Resultado:** Los caracteres especiales se convierten en entidades HTML seguras, eliminando el riesgo de XSS persistente.

### 🔌 Manejo de Errores Críticos (CWE-703)
Se identificaron más de 15 puntos donde se ignoraban errores en operaciones de red y base de datos.
- **Remediación:** Se blindaron todos los cierres de conexiones WebSocket y escrituras de bases de datos con manejo de errores explícito y logging.
- **Resultado:** Prevención de fugas de memoria y estados inconsistentes por fallos silenciosos.

### 📂 Blindaje contra Path Traversal (CWE-22)
Riesgo detectado en la gestión de archivos multimedia que permitía acceso potencial a archivos del sistema operativo.
- **Remediación:** Implementación de `filepath.Clean()` y validaciones estrictas en el repositorio de MinIO.
- **Resultado:** El servidor garantiza que todas las operaciones de archivos se mantengan dentro del ámbito del bucket autorizado.

### 🚫 Control de Sesión Única y Auditoría de Secretos
- **Sesión Única:** Refuerzo de la validación triple (DeviceID + IP + Nickname) para garantizar el cumplimiento del requerimiento de sesión única del PDF.
- **Gitleaks:** Se realizó un escaneo completo de secretos, identificando y mitigando la exposición de llaves en archivos de configuración local.

---
**Resultado Final:** El sistema ha pasado de ser un prototipo funcional a una aplicación con **seguridad por diseño**, alcanzando una cobertura de pruebas del **80.7%** en componentes críticos y mitigando el 100% de las vulnerabilidades de severidad alta/media detectadas.
