# Sistema de Chat Distribuido - Rama `desarrollo-ux`

## Descripción
Esta rama está enfocada en el desarrollo y mejora de la interfaz de usuario (UX/UI) y la experiencia del cliente para el Sistema de Chat Distribuido. 

## Características Principales en esta Rama
* **Migración a Next.js:** Se ha iniciado y desarrollado la migración del frontend hacia Next.js (`frontend-next`) para aprovechar Server-Side Rendering (SSR) y mejores prácticas de enrutamiento y rendimiento.
* **Frontend Svelte (Legacy):** Aún se conserva el frontend inicial desarrollado en Svelte (`frontend`) con diseño Glassmorphism, modo oscuro y animaciones.
* **Mejoras de UI:** 
  - Gestión avanzada de salas (edición y eliminación).
  - Mejoras en la interfaz de subida de archivos y vista previa.
  - Diseño responsivo para ocupar el 100% de la pantalla en dispositivos móviles y escritorio.
* **Conexión con Backend:** Integración con el backend en Go a través de WebSockets para mensajería en tiempo real.

## Estructura de Directorios Clave
* `/frontend-next`: Nuevo cliente web desarrollado con Next.js y React.
* `/frontend`: Cliente web original desarrollado en Svelte.
* `/controladores`, `/modelos`, `/db`: Lógica del backend en Go (compatible con las nuevas interfaces).

## Ejecución (Desarrollo)

### Backend (Go)
```bash
go run main.go
```

### Frontend (Next.js)
```bash
cd frontend-next
npm install
npm run dev
```

### Frontend (Svelte - Opcional)
```bash
cd frontend
npm install
npm run dev
```
