@echo off
REM API Gateway en http://localhost:8000 (Consul en 127.0.0.1:8500)
cd /d "%~dp0"
go run ./main
