# Script para correr toda la aplicación (Backend, Frontend y Servicios)
# Comprueba servicios en WSL2/Docker

Write-Host "--- Iniciando Aplicación Distribuida ---" -ForegroundColor Cyan

# 1. Comprobar Docker y Servicios
Write-Host "[1/4] Verificando Docker y contenedores (WSL)..." -ForegroundColor Yellow
wsl docker compose up -d
if ($LASTEXITCODE -ne 0) {
    Write-Host "Intentando con docker-compose legacy..."
    wsl docker-compose up -d
}
if ($LASTEXITCODE -ne 0) {
    Write-Host "Error al iniciar Docker Compose. Asegúrate de que Docker Desktop esté corriendo." -ForegroundColor Red
    exit $LASTEXITCODE
}

# 2. Esperar a que los servicios estén listos
Write-Host "Esperando 5 segundos para que los servicios se estabilicen..."
Start-Sleep -Seconds 5

# 3. Iniciar Backend (Go)
Write-Host "[2/4] Iniciando Backend (Go) en puerto 8080..." -ForegroundColor Yellow
Start-Process powershell -ArgumentList "-NoExit", "-Command", "go run main.go" -WindowStyle Normal

# 4. Iniciar Frontend (Next.js)
Write-Host "[3/4] Iniciando Frontend (Next.js) en puerto 3000..." -ForegroundColor Yellow
Set-Location .\frontend-next
npm install
Start-Process powershell -ArgumentList "-NoExit", "-Command", "npm run dev" -WindowStyle Normal
Set-Location ..

Write-Host "[4/4] Aplicación iniciada exitosamente." -ForegroundColor Green
Write-Host "Backend: http://localhost:8080"
Write-Host "Frontend: http://localhost:3000"
Write-Host "Redis: localhost:6380"
Write-Host "MongoDB: localhost:27018"
Write-Host "MinIO: http://localhost:9292 (Console)"
