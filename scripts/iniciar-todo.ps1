# Levanta todo el proyecto: infraestructura -> servidores -> clientes.
& "$PSScriptRoot\1-iniciar-infraestructura.ps1"
& "$PSScriptRoot\2-iniciar-servidores.ps1"
Write-Host "Esperando 25 s a que los servidores se registren en Consul..."
Start-Sleep -Seconds 25
& "$PSScriptRoot\3-iniciar-clientes.ps1"
