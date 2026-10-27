@echo off
REM Abre cada componente en su propia ventana, en el orden correcto.
REM Requisitos: Docker Desktop (Consul y Envoy), Go 1.24+, JDK 17+.
cd /d "%~dp0"
start "Directorio Consul"        cmd /k "cd DirectorioConsul && docker compose up"
timeout /t 5 /nobreak >nul
start "Servidor de Streaming"    cmd /k "cd ServidorDeStreaming && go run main/servidor.go"
start "Proxy Envoy"              cmd /k "ProxyEnvoy\iniciar-envoy.bat"
start "Servidor de Suscripciones" cmd /k "cd ServidorDeSuscripciones && go run ./main"
start "Servidor de Reacciones"   cmd /k "cd ServidorDeReacciones && mvnw.cmd spring-boot:run"
start "API Gateway"              cmd /k "cd ApiGateway && go run ./main"
echo.
echo Espere a que todos los servicios aparezcan en http://localhost:8500 y luego abra ClienteWeb\index.html
timeout /t 25 /nobreak >nul
start "" "ClienteWeb\index.html"
