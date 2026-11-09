# Servidor de suscripciones sin simulacion de fallos.
Set-Location (Split-Path -Parent $PSScriptRoot)
$env:MODO_FALLO = "NINGUNO"; $env:CANTIDAD_FALLOS = "0"
go run .
