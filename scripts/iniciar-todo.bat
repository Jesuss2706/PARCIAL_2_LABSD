@echo off
REM Doble clic para levantar todo el proyecto (evita el bloqueo de scripts de PowerShell).
powershell -NoProfile -ExecutionPolicy Bypass -File "%~dp0iniciar-todo.ps1"
pause
