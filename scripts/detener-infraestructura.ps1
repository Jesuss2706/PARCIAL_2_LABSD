# Apaga Consul, RabbitMQ y Envoy. Los servidores se cierran con Ctrl+C en su ventana.
$raiz = Split-Path -Parent $PSScriptRoot
Push-Location "$raiz\ProxyEnvoy";       docker compose down; Pop-Location
Push-Location "$raiz\DirectorioCONSUL"; docker compose down; Pop-Location
