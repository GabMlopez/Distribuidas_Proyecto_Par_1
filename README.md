# Sistema de Chat Distribuido - Rama `main`

## Descripción
Esta es la rama principal del proyecto, que contiene una versión estable y funcional del Sistema de Chat Distribuido. Incluye toda la integración entre la arquitectura backend, el manejo seguro de archivos, almacenamiento avanzado y el cliente web inicial.

## Características Principales en esta Rama
* **Almacenamiento Avanzado (MinIO):** Migración del almacenamiento de archivos hacia MinIO, permitiendo un manejo robusto de objetos compatible con S3.
* **Persistencia de Mensajes:** Integración con bases de datos (MongoDB) para asegurar el registro y guardado del historial de conversaciones.
* **Seguridad (Full Patch):** 
  - Prevención de inyección XSS (sanitización de inputs).
  - Remediación de vulnerabilidades IDOR y Path Traversal.
  - Implementación de cabeceras de seguridad CSP.
* **Cliente Web (Svelte):** Interfaz inicial responsiva y fluida.
* **Comunicación en Tiempo Real:** Infraestructura sólida de WebSockets para envío de mensajes multimedia.

## Estructura de Directorios Clave
* `/frontend`: Aplicación cliente desarrollada en Svelte.
* `/db`: Controladores de persistencia (MongoDB, Redis, MinIO).
* `/controladores` y `/middleware`: Lógica de negocios del chat, protección de rutas HTTP/WS y manejo de archivos.
* `docker-compose.yml`: Archivo de orquestación para la infraestructura (Bases de datos, etc.).

## Ejecución (Desarrollo / Producción)

### Requisitos previos
- Docker & Docker Compose
- Go (v1.26.2+)
- Node.js (Para ejecutar el frontend)

### Levantar Infraestructura
Es necesario tener MinIO, MongoDB y Redis funcionando:
```bash
docker-compose up -d
```

### Ejecutar Backend (Go)
```bash
# Ejecutar desde la raíz del proyecto
go run main.go
```

### Ejecutar Frontend (Svelte)
```bash
cd frontend
npm install
npm run dev
```
