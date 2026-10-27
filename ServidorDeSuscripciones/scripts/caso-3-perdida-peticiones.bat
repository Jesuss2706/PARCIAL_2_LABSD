@echo off
REM Caso 3: sin respuesta tras el cuarto reintento por perdida de mensajes de peticion
REM MODO_FALLO=PETICION  CANTIDAD_FALLOS=5
cd /d "%~dp0.."
set MODO_FALLO=PETICION
set CANTIDAD_FALLOS=5
set RETARDO_FALLO_MS=5000
go run ./main
