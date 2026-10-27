@echo off
REM Caso 6: se pierden las 2 primeras respuestas; el pago se hizo en el primer intento y el reintento 2 NO cobra de nuevo
REM MODO_FALLO=RESPUESTA  CANTIDAD_FALLOS=2
cd /d "%~dp0.."
set MODO_FALLO=RESPUESTA
set CANTIDAD_FALLOS=2
set RETARDO_FALLO_MS=5000
go run ./main
