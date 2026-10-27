@echo off
REM Servidor de suscripciones en el puerto 8084, sin simulación de fallos.
cd /d "%~dp0"
set MODO_FALLO=NINGUNO
set CANTIDAD_FALLOS=0
go run ./main
