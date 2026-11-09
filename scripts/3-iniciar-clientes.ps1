# Compila el cliente web con webpack (solo si falta dist\main.js) y abre ambos clientes.
$raiz = Split-Path -Parent $PSScriptRoot

Push-Location "$raiz\ClienteWeb"
if (-not (Test-Path "node_modules")) {
    Write-Host "Instalando dependencias del cliente web (npm install)..." -ForegroundColor Green
    npm install
}
if (-not (Test-Path "dist\main.js")) {
    Write-Host "Empaquetando el cliente web (npx webpack)..." -ForegroundColor Green
    npx webpack
}
Pop-Location

Start-Process "$raiz\ClienteWeb\index.html"
Start-Process "$raiz\AdministradorWeb\index.html"
Write-Host "Cliente web y administrador web abiertos en el navegador." -ForegroundColor Green
