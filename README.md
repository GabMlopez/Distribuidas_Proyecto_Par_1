# 💬 Sistema de Chat Distribuido en Tiempo Real

Sistema de mensajería distribuida con soporte para salas de texto y multimedia, control de sesiones único por dispositivo (IP), almacenamiento de archivos en la nube con MinIO, y autenticación basada en JWT.

---

## 🏗️ Arquitectura

### Vista General (Contenedores)
```text
┌─────────────────────────────┐
│  Frontend (Next.js 15)       │  http://localhost:3000
│  React 19 + TailwindCSS      │
└────────────┬────────────────┘
             │  HTTP REST + WebSocket
┌────────────▼────────────────┐
│  Backend (Go + Gin)          │  http://localhost:8080
│  MongoDB + Redis + MinIO     │
└────────────┬────────────────┘
             │
┌────────────▼────────────────┐
│  Infraestructura (Docker)    │
│  MongoDB  · Redis · MinIO    │
└─────────────────────────────┘
```

### Diagrama de Flujo del Sistema
```mermaid
graph TD
    %% Clients
    User[Web Client / Browser]

    %% Frontend
    subgraph Frontend [Frontend - Vercel / Next.js]
        NextApp[Next.js App UI]
    end

    %% Backend
    subgraph Backend [Backend - Render / Go]
        API[Go API - Gin Framework]
        WSHub[WebSocket Hub]
    end

    %% Storage & Databases
    subgraph Databases [Data & Storage]
        Mongo[(MongoDB)]
        Redis[(Redis Pub/Sub)]
        MinIO[(MinIO Object Storage)]
    end

    %% Connections
    User -- HTTP / REST --> NextApp
    NextApp -- REST API --> API
    NextApp -- WebSocket --> WSHub

    API -- CRUD Operations --> Mongo
    API -- Upload / Download --> MinIO
    
    WSHub -- Publish / Subscribe --> Redis
    Redis -- Sync Messages --> WSHub
    WSHub -- Store Messages --> Mongo
```

---

## ✨ Características Principales

### 🔒 Seguridad y Control de Sesiones
- **Sesión única por dispositivo (IP):** El servidor bloquea con HTTP 409 cualquier intento de conexión desde una IP que ya tiene una sesión activa, incluso desde modo incógnito o distintas pestañas.
- **JWT:** Autenticación basada en tokens para acceso a salas y subida de archivos.
- **Cabeceras de seguridad:** CSP, X-Frame-Options y X-Content-Type-Options configurados.
- **Sanitización de inputs:** Prevención de ataques XSS e IDOR en todos los endpoints sensibles.
- **Path Traversal:** Mitigado en la gestión de archivos usando `filepath.Base()`.

### 💬 Mensajería en Tiempo Real
- **WebSockets** para mensajes de texto e indicadores de escritura.
- **Salas de texto y multimedia** con PIN de acceso opcional.
- **Desconexión limpia:** Al cerrar la ventana/pestaña, el estado `activo` del usuario se libera automáticamente en la base de datos, permitiendo reconectarse desde el mismo dispositivo.
- **Broadcast** a todos los miembros de la sala activos.

### 📁 Almacenamiento de Archivos (MinIO)
- Subida segura de imágenes, documentos, audio, video y archivos comprimidos.
- Máximo de 100 MB por archivo.
- Los archivos se sirven directamente desde MinIO con `Content-Disposition` correcto.
- Vista previa de imágenes en el chat.

### 🖥️ Frontend (Next.js)
- Interfaz moderna con modo oscuro, glassmorphism y animaciones.
- Diseño completamente responsivo.
- Gestión de salas: crear, editar, eliminar y unirse con PIN.
- Sala de chat con indicador de escritura, panel lateral de usuarios y soporte de multimedia.

---

## 📁 Estructura del Proyecto

```
/
├── frontend-next/          # Frontend (Next.js 15 + React 19 + TailwindCSS)
│   └── src/
│       ├── app/            # Páginas: / (login), /home, /room/[id]
│       ├── components/     # Componentes reutilizables (RoomCard, ChatMessage, etc.)
│       ├── context/        # UserContext (estado global del usuario y deviceId)
│       └── types/          # Tipos TypeScript compartidos
│
├── controladores/          # Handlers HTTP del backend (Go)
│   ├── salas_controller.go       # CRUD de salas + validación de sesión única por IP
│   ├── web_socket_controller.go  # Upgrade HTTP → WebSocket
│   ├── upload_controller.go      # Subida/descarga de archivos vía MinIO
│   ├── admin_controller.go       # Endpoints de administración
│   └── sockets/
│       ├── hub.go                # Gestión del Hub de WebSockets + cleanup al desconectar
│       └── cliente.go            # Estructura del cliente WebSocket
│
├── modelos/                # Structs de MongoDB (Usuario, Sala, Mensaje)
├── db/
│   ├── mongo.go            # Conexión a MongoDB
│   ├── redis.go            # Conexión a Redis
│   └── minio.go            # Cliente MinIO + inicialización del bucket
│
├── utils/
│   └── jwt.go              # Generación y validación de tokens JWT
│
├── main.go                 # Entry point + configuración de rutas Gin
└── docker-compose.yml      # Infraestructura: MongoDB, Redis, MinIO
```

---

## 🚀 Instalación y Ejecución

### Prerrequisitos
- [Go 1.21+](https://go.dev/dl/)
- [Node.js 20+](https://nodejs.org/)
- [Docker](https://www.docker.com/) (corriendo en WSL si usas Windows)

### 1. Levantar la Infraestructura (Docker)

```bash
# Desde WSL o Linux:
docker compose up -d
```

Esto levanta:
| Servicio  | Puerto Host | Descripción            |
|-----------|-------------|------------------------|
| MongoDB   | `27018`     | Base de datos principal |
| Redis     | `6380`      | Pub/Sub y caché        |
| MinIO     | `9191`      | Almacenamiento S3      |
| MinIO UI  | `9292`      | Consola web de MinIO   |

> **Credenciales MinIO por defecto:** user: `admin` / pass: `password123`

### 2. Backend (Go)

```bash
# En la raíz del proyecto
go mod tidy
go run main.go
```

El servidor queda disponible en `http://localhost:8080`.

### 3. Frontend (Next.js)

```bash
cd frontend-next
npm install
npm run dev
```

La interfaz queda disponible en `http://localhost:3000`.

---

## 🔌 API Endpoints Principales

| Método | Ruta                  | Descripción                                   |
|--------|-----------------------|-----------------------------------------------|
| POST   | `/rooms/create`       | Crear sala (requiere token Admin)             |
| GET    | `/rooms/list`         | Listar todas las salas                        |
| POST   | `/rooms/join`         | Unirse a sala (valida PIN + sesión única IP)  |
| PUT    | `/rooms/update`       | Actualizar sala (requiere token Admin)        |
| DELETE | `/rooms/delete`       | Eliminar sala (requiere token Admin)          |
| POST   | `/rooms/leave`        | Abandonar sala manualmente                    |
| GET    | `/ws?room=ID&device_id=X` | Conexión WebSocket a sala             |
| POST   | `/upload/file`        | Subir archivo multimedia                      |
| GET    | `/upload/file/:name`  | Descargar/ver archivo desde MinIO             |

---

## 🛡️ Lógica de Sesión Única

1. El cliente genera un `deviceId` basado en hardware (GPU, CPU, RAM, pantalla) en el frontend.
2. Al hacer `/rooms/join`, el **servidor** verifica si la **IP local** del cliente ya tiene un usuario con `activo: true` en la base de datos.
3. Si ya existe una sesión activa, responde con **HTTP 409 Conflict**.
4. Al desconectarse el WebSocket, el Hub automáticamente actualiza `activo: false` en MongoDB, liberando el dispositivo para reconectarse.

---

## 🧑‍💻 Credenciales de Prueba

| Usuario | Contraseña | Rol   |
|---------|------------|-------|
| Admin   | admin123   | Admin |

> Los usuarios regulares solo necesitan elegir un nickname al unirse a una sala.

