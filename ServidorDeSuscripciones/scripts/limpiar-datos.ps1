# Borra los tokens y transacciones guardados (datos\*.json).
Set-Location (Split-Path -Parent $PSScriptRoot)
Remove-Item -ErrorAction SilentlyContinue datos\tokens.json, datos\transacciones.json
Write-Host "Datos de suscripciones eliminados."
