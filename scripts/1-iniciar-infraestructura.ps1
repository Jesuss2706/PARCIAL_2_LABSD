# Levanta la infraestructura en Docker Desktop:
#   - Directorio Consul (http://localhost:8500) y RabbitMQ (http://localhost:15672, admin / 1234)
#   - Proxy Envoy (gRPC-Web -> gRPC) en el puerto 8080, en una ventana aparte para ver sus ecos
$raiz = Split-Path -Parent $PSScriptRoot

Write-Host "== Directorio Consul + RabbitMQ ==" -ForegroundColor Green
Push-Location "$raiz\DirectorioCONSUL"
docker compose up -d
Pop-Location

Write-Host "== Proxy Envoy (ventana aparte) ==" -ForegroundColor Green
Start-Process powershell -ArgumentList '-NoExit', '-Command', "`$host.UI.RawUI.WindowTitle='ProxyEnvoy'; cd '$raiz\ProxyEnvoy'; docker compose up"

Write-Host "Esperando 10 s a que Consul y RabbitMQ terminen de arrancar..."
Start-Sleep -Seconds 10
Write-Host "Listo. Consul: http://localhost:8500  RabbitMQ: http://localhost:15672" -ForegroundColor Green
