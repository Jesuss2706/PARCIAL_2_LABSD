# Laboratorio de Sistemas Distribuidos — Requerimiento segundo corte

Plataforma de streaming de audios (canciones, podcasts, audiolibros y ruido blanco) con 10 componentes:

| Componente | Tecnología | Puerto | Base (ejercicio de clase) |
|---|---|---|---|
| **ClienteWeb** | HTML + JS, gRPC-Web, STOMP (webpack) | archivo `index.html` | GUIA_10 `cliente-spotify` + GUIA_8 `vista` |
| **AdministradorWeb** | HTML + JS | archivo `index.html` | PARCIAL_1 `Administrador` (pasado a web) |
| **APIGateway** | Go + gin | 8000 | GUIA_9 `api-gateway` |
| **DirectorioCONSUL** | Consul en Docker | 8500 | GUIA_9 `directorio-consul` |
| **ServidorDeAudioyMetadatos** | Go + gin (REST) | 5001 | PARCIAL_1 `ServidorMetadataDeAudios` + `ServidorDeAudios` |
| **ServidorDeSuscripciones** | Go + gin (REST) | 5002 | GUIA_10 `servidor-pasarela-pagos` |
| **ServidorDeEstadisticas** | Go + gin (REST) + RabbitMQ | 5003 | PARCIAL_1 `ServidorDeEstadisticas` |
| **ServidorDeReacciones** | Java Spring Boot (WebSocket + STOMP) | 5000 | GUIA_8 `servidorchat` |
| **ProxyEnvoy** | Envoy en Docker (no se desarrolla, solo se configura) | 8080 | — |
| **ServidorDeStreaming** | Go + gRPC + RabbitMQ | 50051 | PARCIAL_1 `ServidorDeStreaming` |

Además se usa **RabbitMQ** (Docker, puerto 5672) para la comunicación asíncrona streaming → estadísticas.

```
Cliente web ─┐                    ┌─> Directorio CONSUL (resuelve dónde está cada microservicio)
             ├─> API Gateway :8000┼─> Servidor de suscripciones :5002
Admin web ───┘                    ├─> Servidor de reacciones :5000 (WebSocket)
                                  ├─> Servidor de audios y metadatos :5001 ──(guarda el mp3)──┐
                                  ├─> Servidor de estadísticas :5003                          │
                                  └─> Proxy Envoy :8080 ─gRPC─> Servidor de streaming :50051 <┘
                                                                   │
                                                  RabbitMQ (cola estadisticas_reproducciones)
                                                                   └─> Servidor de estadísticas
```

Los microservicios se registran solos en Consul al arrancar (latido TTL cada 5 s) y **solo aceptan peticiones que pasen por el API Gateway** (cabecera secreta `X-Gateway-Token`; si llama directo a `localhost:5001/tipos` responde 403).

---

## 1. Requisitos en Windows

| Programa | Para qué | Comprobar |
|---|---|---|
| Go 1.24 o superior | servidores en Go | `go version` |
| Java JDK 17 o superior (con `JAVA_HOME`) | servidor de reacciones | `java -version` |
| Node.js 18 o superior | empaquetar el cliente web | `node -v` |
| Docker Desktop | Consul, RabbitMQ y Envoy | `docker version` |

Maven **no** hace falta instalarlo: el proyecto trae `mvnw.cmd`.
La primera ejecución necesita internet (Go, Maven, npm y Docker descargan dependencias e imágenes).

---

## 2. Ejecución rápida

Doble clic en **`scripts\iniciar-todo.bat`** (o en PowerShell: `.\scripts\iniciar-todo.ps1`). El script:

1. `1-iniciar-infraestructura.ps1` → Consul + RabbitMQ (`docker compose up -d`) y Envoy en su propia ventana.
2. `2-iniciar-servidores.ps1` → abre **una ventana por servidor** para ver sus ecos.
3. `3-iniciar-clientes.ps1` → `npm install` + `npx webpack` (solo la primera vez) y abre el cliente y el administrador.

Si Windows pregunta por el firewall al iniciar los servidores, dé **Permitir**.
Puede comprobar que todo quedó registrado en **http://localhost:8500** (deben aparecer `servidor-audios`, `servidor-suscripciones`, `servidor-estadisticas` y `servidor-reacciones`).

Para apagar: `Ctrl+C` en cada ventana y `.\scripts\detener-infraestructura.ps1`.

## 3. Ejecución manual (servidor por servidor)

```powershell
# Infraestructura
cd DirectorioCONSUL;  docker compose up -d          # Consul + RabbitMQ
cd ..\ProxyEnvoy;     docker compose up             # Envoy (dejar la ventana abierta)

# Servidores (cada uno en su propia terminal)
cd ServidorDeStreaming;        go mod tidy; go run ./main
cd ServidorDeAudioyMetadatos;  go mod tidy; go run .
cd ServidorDeSuscripciones;    go mod tidy; go run .
cd ServidorDeEstadisticas;     go mod tidy; go run ./main
cd ServidorDeReacciones;       .\mvnw.cmd spring-boot:run
cd APIGateway;                 go mod tidy; go run .

# Clientes
cd ClienteWeb; npm install; npx webpack   # y abrir ClienteWeb\index.html
# abrir AdministradorWeb\index.html
```

> `go mod tidy` se ejecuta siempre antes de `go run` porque el proyecto se entrega sin `go.sum`: así cada equipo descarga y verifica sus propias dependencias.

---

## 4. Qué hace cada opción del requerimiento y dónde está

### a) Reacciones por WebSocket (ServidorDeReacciones)
* El cliente abre **una sola conexión** WebSocket/STOMP a `ws://localhost:8000/ws-stomp` (pasa por el gateway).
* Cada audio tiene su **canal**: `/brokerDeReacciones/audio/{idAudio}`.
* En `ClienteWeb/js/funciones.js` (función `configurarReproductor`) está el *listener* del reproductor:
  al darle **play** el cliente se **registra** en el canal (SUBSCRIBE) y al **pausar** se **des-registra** (UNSUBSCRIBE). Se muestra el eco "Se ha pausado el audio…".
* Las reacciones se envían a `/app/enviarReaccion/{idAudio}`; el servidor (`ReaccionesController`) las reenvía como callback, por la conexión ya abierta, **solo** a quienes escuchan ese mismo audio.
* `GestorCanales` lleva la cuenta de oyentes por canal y avisa cuando alguien entra o sale.

### b) Catálogo y streaming (ServidorDeAudioyMetadatos + Envoy + ServidorDeStreaming)
Tipos → audios del tipo → detalle con sus metadatos → **Reproducir**. El audio llega en fragmentos de 32 KB por gRPC-Web; el reproductor se vuelve visible con el **primer fragmento** y suena mientras siguen llegando (MediaSource).

### c) Suscripciones con token de idempotencia (ServidorDeSuscripciones)
1. `POST /api/suscripciones/tokens` → genera el token.
2. `POST /api/suscripciones/pagos` → paga con ese token. Si no hay respuesta en 3 s, un **temporizador** reintenta a los 2, 4, 8 y 16 s **con el mismo token**.
3. Ecos en el cliente (panel "Ecos del pago") y en el servidor (`[ECO PAGO]`, `[ECO ALMACENAMIENTO]`): token en uso, si el pago se realizó y si se almacenó en `datos\transacciones.json`.

Para demostrar los reintentos, inicie el servidor con fallos simulados:
```powershell
.\ServidorDeSuscripciones\scripts\iniciar-perdida-peticiones.ps1   # el pago NO se procesa en los 2 primeros intentos
.\ServidorDeSuscripciones\scripts\iniciar-perdida-respuestas.ps1   # el pago SÍ se procesa pero se pierden 2 respuestas
.\ServidorDeSuscripciones\scripts\limpiar-datos.ps1                # borra tokens y transacciones
```

### d) Historial del cliente (ServidorDeEstadisticas)
Cada vez que empieza un streaming, el servidor de streaming publica en RabbitMQ (asíncrono) quién escuchó qué. Estadísticas consume la cola, guarda en `datos\reproducciones.json` y expone `GET /api/estadisticas/clientes/{usuario}`. Pestaña **Mi historial** del cliente.

### e) Simular un token ya utilizado
Botón **"Simular token ya utilizado"** (pestaña Suscripciones): busca un token con el que ya se pagó y lo reenvía. El servidor responde `previamente_realizada: true` y **no cobra de nuevo**.

### f) Administrador web (AdministradorWeb)
Elige el tipo → el formulario pide **los metadatos propios de ese tipo** (Música: artista, álbum, género; Podcasts: nombre, episodio, anfitrión; Audiolibros: título, autor, narrador; Ruido blanco: tipo de sonido, fuente, uso) → título y mp3 (la duración se calcula sola) → `POST /api/audios` (multipart).
El servidor valida que estén todos los metadatos del tipo y que no vengan metadatos de otro tipo, guarda el mp3 en `ServidorDeStreaming\audios\` y el registro en `ServidorDeAudioyMetadatos\datos\audios.json`.

---

## 5. Cómo se construyeron el Proxy Envoy y el cliente gRPC-Web

No venían los archivos del profesor, así que se crearon así:

### Proxy Envoy (`ProxyEnvoy/`)
* **`envoyConfig2.yaml`**: un *listener* en `0.0.0.0:8080` con tres filtros HTTP en orden:
  1. `envoy.filters.http.grpc_web` → convierte gRPC-Web (HTTP/1.1, lo que entiende el navegador) en gRPC.
  2. `envoy.filters.http.cors` → permite llamadas desde otro origen y expone `grpc-status` / `grpc-message`.
  3. `envoy.filters.http.router` → envía al *cluster* `servidor_streaming`.

  El *cluster* apunta a `host.docker.internal:50051` (su PC visto desde Docker) con `http2_protocol_options`, porque gRPC exige HTTP/2. `timeout: 0s` evita que Envoy corte streams largos.
* **`docker-compose.yml`**: ejecuta la imagen oficial `envoyproxy/envoy:v1.31-latest` con ese archivo. Se usa Docker porque Envoy ya no publica binarios oficiales para Windows.
  El comando es el mismo del PDF: `envoy -c envoyConfig2.yaml --disable-hot-restart`.
* **`envoyConfig2.local.yaml`**: la misma configuración apuntando a `127.0.0.1:50051`, por si el profesor les da un `envoyServer`/`envoy.exe` nativo: `.\envoyServer -c envoyConfig2.local.yaml --disable-hot-restart`.

En el diagrama el cliente llega al Envoy **a través del API Gateway**: el gateway reenvía todo lo que empieza por `/servicios.AudioService/` a `http://127.0.0.1:8080` sin almacenar en búfer (`FlushInterval: -1`), así cada fragmento llega al navegador apenas sale del servidor. El gateway es el único que pone cabeceras CORS (borra las de Envoy y los microservicios para que no se dupliquen).

### Cliente gRPC-Web (`ClienteWeb/`)
1. Se copió `audio.proto` del servidor de streaming a `ClienteWeb/proto/`.
2. Se generaron los *stubs* JavaScript con `protoc` y dos plugins:
   ```
   protoc -I=proto audio.proto ^
     --js_out=import_style=commonjs:src/serviciosAudio ^
     --grpc-web_out=import_style=commonjs,mode=grpcwebtext:src/serviciosAudio
   ```
   Esto produce `audio_pb.js` (mensajes `AudioRequest`/`AudioChunk`) y `audio_grpc_web_pb.js` (`AudioServiceClient`). Se usa `mode=grpcwebtext` porque es el modo de gRPC-Web que soporta *server streaming*. Los archivos generados ya vienen incluidos; `generar-proto.ps1` explica cómo regenerarlos si cambia el `.proto`.
3. `src/cliente_receptor_fragmentos.js` crea el `AudioServiceClient`, llama `audioStream(...)` y escucha los eventos `data` (cada fragmento), `end` y `error`. Cada fragmento se agrega a un `MediaSource` conectado al `<audio>`, por eso el audio empieza a sonar antes de terminar la descarga.
4. `npx webpack` empaqueta ese archivo junto con `grpc-web`, `google-protobuf` y `@stomp/stompjs` en `dist/main.js`, que es lo que carga `index.html`. **Si modifica `cliente_receptor_fragmentos.js`, ejecute `npx webpack` otra vez** (igual que dice el PDF). `funciones.js`, `api.js` y `config.js` son scripts normales y no requieren webpack.

---

## 6. Rutas del API Gateway

| Ruta pública (gateway :8000) | Servicio en Consul | Ruta interna |
|---|---|---|
| `GET /api/tipos` · `GET /api/tipos/{id}/metadatos` · `GET /api/tipos/{id}/audios` | servidor-audios | `/tipos/...` |
| `GET /api/audios/{id}` · `POST /api/audios` (multipart) | servidor-audios | `/audios/...` |
| `GET /api/suscripciones/planes` · `POST .../tokens` · `POST .../pagos` · `GET .../transacciones/{cedula}` · `GET .../estado` | servidor-suscripciones | `/suscripciones/...` |
| `GET /api/estadisticas/clientes/{usuario}` · `GET /api/estadisticas/resumen` | servidor-estadisticas | `/estadisticas/...` |
| `ws://.../ws-stomp` (WebSocket nativo) · `/ws` (SockJS) | servidor-reacciones | `/ws-stomp` · `/ws` |
| `POST /servicios.AudioService/AudioStream` (gRPC-Web) | Proxy Envoy :8080 (fijo) | → streaming :50051 |

## 7. Datos guardados (archivos JSON, igual que la guía 10)

| Archivo | Contenido |
|---|---|
| `ServidorDeAudioyMetadatos\datos\audios.json` | audios registrados (se crea con 5 audios de ejemplo la primera vez) |
| `ServidorDeStreaming\audios\*.mp3` | los mp3 (5 de ejemplo de 30 s, generados, sin derechos de autor) |
| `ServidorDeSuscripciones\datos\tokens.json`, `transacciones.json` | tokens de idempotencia y pagos |
| `ServidorDeEstadisticas\datos\reproducciones.json` | historial de reproducciones |

Borre estos `.json` para volver al estado inicial.

## 8. Problemas comunes

| Síntoma | Causa / solución |
|---|---|
| `no hay instancias disponibles de "servidor-…"` | Ese servidor no está corriendo o no alcanzó a registrarse en Consul. Revise http://localhost:8500. |
| El audio no suena y el eco dice `Http response at 400 or 500 level` | Envoy no está arriba o no alcanza el streaming. Revise la ventana de Envoy y que el streaming esté en :50051. |
| `Reacciones: desconectado` | El servidor de reacciones (Spring) aún está arrancando; el cliente reintenta cada 5 s. |
| `Falta dist/main.js` | Ejecute `npx webpack` dentro de `ClienteWeb`. |
| `mvnw.cmd` falla con `JAVA_HOME` | Configure la variable de entorno `JAVA_HOME` con la carpeta del JDK 17+. |
| Puerto ocupado | Cambie el puerto con la variable `PORT` (Go) o `server.port` (Spring) y actualice `APIGateway`/`config.js` si aplica. |
