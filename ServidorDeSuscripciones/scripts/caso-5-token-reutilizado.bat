@echo off
REM Caso 5: transaccion con un token previamente utilizado (boton 'Simular reutilizar este token')
REM MODO_FALLO=NINGUNO  CANTIDAD_FALLOS=0
cd /d "%~dp0.."
set MODO_FALLO=NINGUNO
set CANTIDAD_FALLOS=0
set RETARDO_FALLO_MS=5000
go run ./main
