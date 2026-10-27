@echo off
REM Caso 4: sin respuesta tras el cuarto reintento por perdida de mensajes de respuesta
REM MODO_FALLO=RESPUESTA  CANTIDAD_FALLOS=5
cd /d "%~dp0.."
set MODO_FALLO=RESPUESTA
set CANTIDAD_FALLOS=5
set RETARDO_FALLO_MS=5000
go run ./main
