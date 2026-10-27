@echo off
REM Caso 2: transaccion realizada al tercer reintento (se pierden 3 peticiones)
REM MODO_FALLO=PETICION  CANTIDAD_FALLOS=3
cd /d "%~dp0.."
set MODO_FALLO=PETICION
set CANTIDAD_FALLOS=3
set RETARDO_FALLO_MS=5000
go run ./main
