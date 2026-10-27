@echo off
REM Directorio Consul (UI en http://localhost:8500). Requiere Docker Desktop.
REM Alternativa sin Docker: descargar consul.exe y ejecutar "consul agent -dev -client=0.0.0.0 -ui"
cd /d "%~dp0"
docker compose up
