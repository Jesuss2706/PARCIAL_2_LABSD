@echo off
REM Borra el historial de tokens y transacciones para empezar la practica desde cero.
cd /d "%~dp0.."
if exist datos\tokens.json del /q datos\tokens.json
if exist datos\transacciones.json del /q datos\transacciones.json
echo Historial de tokens y transacciones eliminado.
