# Sistema de Chat Distribuido - Rama `development/backend`

## Descripción
Esta rama se enfoca en el desarrollo de la lógica del servidor (Backend), integración con bases de datos, seguridad y manejo de WebSockets. Contiene los servicios esenciales para que la comunicación en tiempo real sea posible y segura.

## Características Principales en esta Rama
* **Manejo de WebSockets:** Gestión de conexiones en tiempo real, salas y broadcasting de mensajes (usando `gorilla/websocket` y Hub pattern).
* **Seguridad Reforzada:** 
  - Restricción estricta de una sola sesión por dirección IP / Usuario en los WebSockets.
  - Implementación de cabeceras de seguridad (CSP, X-Frame-Options, nosniff).
  - Prevención de vulnerabilidades como Path Traversal e IDOR.
* **Manejo de Archivos:** Soporte para subida de imágenes y vista previa dentro del chat.
* **Bases de Datos:** Conexión y estructuración para el almacenamiento (MongoDB para persistencia y Redis para manejo de mensajes en colas/pub-sub).

## Estructura de Directorios Clave
* `/controladores`: Lógica de enrutamiento y manejo de peticiones HTTP/WS.
  * `/controladores/sockets`: Lógica específica para el Hub y los clientes WebSocket.
* `/db`: Configuración de conexión a MongoDB y Redis.
* `/middleware`: Middlewares de autenticación (JWT) y seguridad.
* `/modelos`: Modelos de datos (Usuario, Sala, Mensaje, Admin).
* `/utils`: Utilidades generales como JWT y validaciones de seguridad.

## Ejecución (Desarrollo)

### Requisitos previos
- Docker & Docker Compose (Para levantar las bases de datos)
- Go (v1.26.2+)

### Levantar Base de Datos
```bash
docker-compose up -d
```

### Ejecutar Backend (Go)
```bash
go run main.go
```
