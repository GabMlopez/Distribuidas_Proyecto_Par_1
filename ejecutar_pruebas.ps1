# Script para ejecutar pruebas de carga y seguridad (k6 + DDoS Protection)

Write-Host "--- Iniciando Suite de Pruebas de Estrés y Seguridad ---" -ForegroundColor Cyan

# 1. Prueba de Carga con k6 (WSL)
Write-Host "[1/2] Ejecutando prueba de carga con k6 en WSL (> 50 usuarios)..." -ForegroundColor Yellow

# Intentar obtener la IP del host de Windows para WSL
$all_ips = Get-NetIPAddress -AddressFamily IPv4 | Where-Object { $_.InterfaceAlias -match "WSL" }
$host_ip = $all_ips.IPAddress | Select-Object -First 1

Write-Host "DEBUG: IPs encontradas para WSL: $($all_ips.IPAddress -join ', ')" -ForegroundColor Gray

if ([string]::IsNullOrWhiteSpace($host_ip)) { 
    # Fallback: intentar obtenerla via wsl route
    $host_ip = wsl bash -c "ip route show | grep default | awk '{print `$3}'"
    if ($host_ip) { $host_ip = $host_ip.Trim() }
}

if ([string]::IsNullOrWhiteSpace($host_ip)) {
    $host_ip = "127.0.0.1"
    Write-Host "No se detectó IP de host, usando 127.0.0.1" -ForegroundColor Yellow
} else {
    Write-Host "Host IP detectado: $host_ip" -ForegroundColor Cyan
}

# Convertimos la ruta de Windows a formato WSL
$wsl_path = wsl wslpath -u (Get-Item .\load_test.js).FullName
wsl k6 run --env BASE_URL="$host_ip:8085" $wsl_path

# 2. Prueba de Protección DDoS / Brute Force (Rate Limiting)
Write-Host "`n[2/2] Probando protección contra DDoS y Fuerza Bruta (Rate Limiting)..." -ForegroundColor Yellow
Write-Host "Enviando ráfaga de 15 peticiones rápidas (Límite es 5 req/s, Burst 10)..."

$successCount = 0
$limitedCount = 0

for ($i=1; $i -le 15; $i++) {
    $response = curl.exe -s -o /dev/null -w "%{http_code}" http://127.0.0.1:8085/health
    if ($response -eq "429") {
        $limitedCount++
    } elseif ($response -eq "200") {
        $successCount++
    }
}

Write-Host "Resultados del Rate Limiter:"
Write-Host " - Peticiones exitosas: $successCount" -ForegroundColor Green
Write-Host " - Peticiones bloqueadas (429 Too Many Requests): $limitedCount" -ForegroundColor Red

if ($limitedCount -gt 0) {
    Write-Host "SISTEMA PROTEGIDO: El middleware de Rate Limiting bloqueó el exceso de peticiones." -ForegroundColor Green
} else {
    Write-Host "ALERTA: El Rate Limiter no pareció bloquear las peticiones. Revisa la configuración." -ForegroundColor Yellow
}

Write-Host "`n--- Pruebas finalizadas ---" -ForegroundColor Cyan
