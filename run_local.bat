@echo off
echo =========================================
echo Iniciando aplicacion localmente...
echo =========================================

echo Levantando servicios base con Docker Compose (MongoDB, Redis, MinIO) en WSL...
wsl docker compose up -d

echo Iniciando el backend (Go)...
start "Backend (Go)" cmd /k "go run main.go"

echo Iniciando el frontend (Next.js)...
start "Frontend (Next.js)" cmd /k "cd frontend-next && npm run dev"

echo.
echo Los servicios se estan ejecutando en ventanas separadas.
echo Backend deberia estar en el puerto correspondiente (ej. 8080/8000).
echo Frontend deberia estar en http://localhost:3000
echo.
pause
