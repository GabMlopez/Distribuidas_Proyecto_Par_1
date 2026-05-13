@echo off
echo =========================================
echo Iniciando aplicacion localmente...
echo =========================================

echo Levantando servicios base con Docker Compose (MongoDB, Redis, MinIO) en WSL...
wsl docker compose up -d

echo Esperando 3 segundos para asegurar que la base de datos MongoDB despierte...
timeout /t 3 /nobreak > NUL

echo.
echo Configurando red local automaticamente...
node setup_local.js

echo Iniciando el backend (Go)...
start "Backend (Go)" cmd /k "go run cmd/api/main.go"

echo Iniciando el frontend (Next.js)...
start "Frontend (Next.js)" cmd /k "cd frontend-next && npm run dev"

echo.
echo Los servicios se estan ejecutando en ventanas separadas.
echo.
pause
