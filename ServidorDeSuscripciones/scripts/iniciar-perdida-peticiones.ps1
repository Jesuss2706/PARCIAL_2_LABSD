# Se pierden las 2 primeras PETICIONES de pago de cada token (el pago NO se procesa en esos intentos).
# El cliente reintenta con temporizadores y el MISMO token; el pago se hace una sola vez.
param([int]$fallos = 2)
Set-Location (Split-Path -Parent $PSScriptRoot)
$env:MODO_FALLO = "PETICION"; $env:CANTIDAD_FALLOS = "$fallos"; $env:RETARDO_FALLO_MS = "5000"
go run .
