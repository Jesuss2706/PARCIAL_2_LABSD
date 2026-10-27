@echo off
REM Caso 1: transaccion realizada al primer intento
REM MODO_FALLO=NINGUNO  CANTIDAD_FALLOS=0
cd /d "%~dp0.."
set MODO_FALLO=NINGUNO
set CANTIDAD_FALLOS=0
set RETARDO_FALLO_MS=5000
go run ./main
