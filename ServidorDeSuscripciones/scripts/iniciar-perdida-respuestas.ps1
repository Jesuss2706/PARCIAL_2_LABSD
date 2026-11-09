# Se pierden las 2 primeras RESPUESTAS de cada token (el pago SI se procesa, pero el cliente no se entera).
# Al reintentar con el mismo token, el servidor responde que ya existia la transaccion (no cobra dos veces).
param([int]$fallos = 2)
Set-Location (Split-Path -Parent $PSScriptRoot)
$env:MODO_FALLO = "RESPUESTA"; $env:CANTIDAD_FALLOS = "$fallos"; $env:RETARDO_FALLO_MS = "5000"
go run .
