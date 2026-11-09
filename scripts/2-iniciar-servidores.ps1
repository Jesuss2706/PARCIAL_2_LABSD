# Abre una ventana de PowerShell por servidor (asi se ven los ecos de cada uno).
# La primera vez "go mod tidy" y Maven descargan dependencias (necesita internet).
$raiz = Split-Path -Parent $PSScriptRoot

function Abrir($titulo, $carpeta, $comando) {
    Write-Host "Iniciando $titulo ..." -ForegroundColor Green
    Start-Process powershell -ArgumentList '-NoExit', '-Command', "`$host.UI.RawUI.WindowTitle='$titulo'; cd '$raiz\$carpeta'; $comando"
}

Abrir 'ServidorDeStreaming'       'ServidorDeStreaming'       'go mod tidy; go run ./main'
Abrir 'ServidorDeAudioyMetadatos' 'ServidorDeAudioyMetadatos' 'go mod tidy; go run .'
Abrir 'ServidorDeSuscripciones'   'ServidorDeSuscripciones'   'go mod tidy; go run .'
Abrir 'ServidorDeEstadisticas'    'ServidorDeEstadisticas'    'go mod tidy; go run ./main'
Abrir 'ServidorDeReacciones'      'ServidorDeReacciones'      '.\mvnw.cmd spring-boot:run'
Start-Sleep -Seconds 3
Abrir 'APIGateway'                'APIGateway'                'go mod tidy; go run .'

Write-Host ""
Write-Host "Cuando todas las ventanas muestren que estan escuchando, ejecute 3-iniciar-clientes.ps1" -ForegroundColor Yellow
Write-Host "Puede revisar los servicios registrados en http://localhost:8500" -ForegroundColor Yellow
