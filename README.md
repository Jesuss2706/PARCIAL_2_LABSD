# Parcial 2 · Laboratorio de Sistemas Distribuidos (Windows)

Implementación de los laboratorios de **reacciones con WebSockets** (GUIA_8) y de
**planes de suscripción con tokens idempotentes** (GUIA_10), integrados en la
arquitectura del segundo corte: todo el tráfico del cliente web pasa por el
**API Gateway**, que resuelve los microservicios en el **directorio Consul**.

```
Cliente web ──► API Gateway :8000 ──► Consul :8500 (descubrimiento)
                   │
                   ├─ /api/suscripciones ──► Servidor de Suscripciones :8084   (REST)
                   ├─ /api/reacciones    ──► Servidor de Reacciones   :8085   (WebSocket STOMP/SockJS)
                   └─ /api/streaming     ──► Proxy Envoy :8080 ──► Servidor de Streaming :50051 (gRPC-Web → gRPC)
```

## Contenido

| Carpeta | Qué es | Tecnología |
|---|---|---|
| `ApiGateway/` | Único punto de entrada; resuelve servicios en Consul, agrega `X-Gateway-Token`, maneja CORS, WebSocket y gRPC-Web | Go (solo librería estándar) |
| `ServidorDeSuscripciones/` | Planes, tokens idempotentes, pagos con simulación de fallos, historial | Go (solo librería estándar) |
| `ServidorDeReacciones/` | Canales de reacciones por audio, registro/des-registro y reenvío por callback | Spring Boot 3.3 + WebSocket (STOMP) |
| `ClienteWeb/` | `index.html` + `funciones.js` (reproductor y reacciones) y `planes.html` (suscripciones) | HTML/JS, gRPC-Web, SockJS/STOMP |
| `ServidorDeStreaming/` | Servidor de streaming entregado por el profesor (sin cambios) | Go + gRPC |
| `ProxyEnvoy/` | Configuración de Envoy (original y para Docker en Windows) | Envoy 1.36 |
| `DirectorioConsul/` | `docker-compose.yml` de Consul (GUIA_9) | Consul 1.20 |

Los dos microservicios nuevos usan capas: `capaControladores` → `capaFachada` → `capaAccesoDatos`,
con `dtos` (JSON) y `modelos`.

## Requisitos en Windows

- **Go 1.24+**: `go version`
- **JDK 17+** con la variable `JAVA_HOME` configurada: `java -version` (Maven no hace falta: se usa `mvnw.cmd`)
- **Docker Desktop**: para Consul y Envoy
- **Node.js**: solo si se modifica `cliente_receptor_fragmentos.js` y hay que regenerar `bundle.js`

> El `envoyServer` que entregó el profesor es un binario de **Linux**. En Windows se usa la imagen
> oficial `envoyproxy/envoy:v1.36.2` (la misma versión) con `ProxyEnvoy/envoyConfigDocker.yaml`, que
> apunta a `host.docker.internal:50051`. Alternativa: ejecutar `./envoyServer -c envoyConfig2.yaml --disable-hot-restart` dentro de WSL.

## Ejecución

Opción rápida: doble clic en **`iniciar-todo.bat`** (abre cada servicio en su propia ventana).

Opción manual, en este orden (una consola por servicio):

| # | Componente | Comando |
|---|---|---|
| 1 | Consul | `DirectorioConsul\iniciar-consul.bat` → UI en http://localhost:8500 |
| 2 | Streaming | `cd ServidorDeStreaming` y `go run main/servidor.go` |
| 3 | Envoy | `ProxyEnvoy\iniciar-envoy.bat` |
| 4 | Suscripciones | `ServidorDeSuscripciones\iniciar-suscripciones.bat` (o un `scripts\caso-*.bat`) |
| 5 | Reacciones | `ServidorDeReacciones\iniciar-reacciones.bat` (la primera vez descarga dependencias) |
| 6 | API Gateway | `ApiGateway\iniciar-gateway.bat` |
| 7 | Cliente | Abrir `ClienteWeb\index.html` en el navegador |

En http://localhost:8500 deben aparecer `servidor-suscripciones` y `servidor-reacciones` en verde.
Los servicios se registran solos al iniciar y se des-registran con Ctrl+C.

Si un microservicio se llama directamente (sin pasar por el gateway) responde **403**.

## Cómo demostrar cada requerimiento

### a) Reacciones por WebSocket (`index.html`)

1. Abra `ClienteWeb\index.html` en **dos navegadores** (o una ventana normal y una de incógnito) y
   escriba un nombre distinto en cada uno.
2. En ambos, pida `cancion1` (formato mp3). Al sonar (**play**) cada cliente se **registra** en el
   canal `cancion1`; la bitácora muestra `REGISTRADO en el canal de reacciones "cancion1"`.
3. Envíe una reacción desde uno: el servidor la reenvía a todos los oyentes del mismo audio.
4. **Pause** en uno de ellos: se **des-registra** del canal y deja de recibir reacciones.
   Al cambiar de audio o cuando el audio termina también se des-registra.

Archivos modificados según el enunciado: `ClienteWeb/index.html` y `ClienteWeb/funciones.js`
(los listeners de `play`/`pause` llaman a `registrarCanal()` y `desregistrarCanal()`).

Ecos del servidor (consola de Spring):

```
[ECO WS] Cliente CONECTADO por WebSocket (sesión 1a2b...)
[ECO WS] Invocado: SUBSCRIBE /canalesReacciones/audio/cancion1 (sesión 1a2b...)
[ECO WS] Val REGISTRADO en el canal "cancion1" (audio: cancion1) · oyentes: 2
[ECO WS] Invocado: SEND /app/enviarReaccion/cancion1 (sesión 1a2b...)
[ECO WS] Reacción "heart" de Val en "cancion1" -> reenviando a 2 cliente(s) del canal
[ECO WS] Invocado: UNSUBSCRIBE sub-1 (sesión 3c4d...)
[ECO WS] JuanPablo DES-REGISTRADO del canal "cancion1" · oyentes: 1
```

Contrato STOMP:

| Acción | Destino |
|---|---|
| Conexión (SockJS) | `http://localhost:8000/api/reacciones/ws` |
| Registrarse en un canal (play) | `SUBSCRIBE /canalesReacciones/audio/{canal}` con cabeceras `usuario` y `tituloAudio` |
| Des-registrarse (pause) | `UNSUBSCRIBE` de esa suscripción |
| Enviar reacción | `SEND /app/enviarReaccion/{canal}` con `{"usuario","tituloAudio","tipo"}`; `tipo` ∈ like, heart, sad, fun |
| Consultar canales (REST) | `GET http://localhost:8000/api/reacciones/canales` |

### b) Streaming

Pedir audio en `index.html`. La petición gRPC-Web pasa por el gateway (`/api/streaming`) hacia Envoy.
Los audios disponibles son los de `ServidorDeStreaming/canciones` (`cancion1`, `cancion2`).

### c) y e) Planes de suscripción con tokens idempotentes (`planes.html`)

1. Elija un plan y pague. El cliente pide primero un **token de idempotencia**
   (`POST /api/suscripciones/tokens`) y luego paga con él (`POST /api/suscripciones/pagos`).
2. Si el servidor no responde en 3 s, el cliente **reintenta con el mismo token**, esperando 2, 4, 8 y 16 s
   (temporizadores en `ClienteWeb/js/config.js`).
3. **e)** El botón *Simular reutilizar este token* reenvía un token ya usado: el servidor responde
   `previamente_realizada: true` y **no cobra de nuevo**.

Ecos del cliente (bitácora y consola del navegador): token utilizado, si el pago se realizó y si quedó almacenado.
Ecos del servidor:

```
[ECO TOKEN] Token de idempotencia generado: 8671cd3e-... (cédula 1061000000, plan duo, $21900)
[ECO PAGO] Token utilizado en el pago: 8671cd3e-... (monto $21900)
[ECO PAGO] Pago realizado exitosamente: transacción 4cdbc56a-..., cédula 1061000000, plan duo, $21900
[ECO ALMACENAMIENTO] Pago almacenado en datos\transacciones.json
[ECO PAGO] Token 8671cd3e-... YA UTILIZADO (transacción 4cdbc56a-... por $21900). NO se cobra de nuevo.
```

Casos de prueba del laboratorio (ejecutar el `.bat` en vez de `iniciar-suscripciones.bat`):

| Script | MODO_FALLO | CANTIDAD_FALLOS | Resultado esperado |
|---|---|---|---|
| `caso-1-primer-intento.bat` | NINGUNO | 0 | Pago al primer intento |
| `caso-2-tercer-reintento.bat` | PETICION | 3 | Pago en el reintento 3 |
| `caso-3-perdida-peticiones.bat` | PETICION | 5 | Sin respuesta tras 4 reintentos; no se cobra |
| `caso-4-perdida-respuestas.bat` | RESPUESTA | 5 | Se cobra 1 vez, el cliente no recibe respuesta |
| `caso-5-token-reutilizado.bat` | NINGUNO | 0 | Usar *Simular reutilizar este token* |
| `caso-6-respuesta-perdida-sin-doble-cobro.bat` | RESPUESTA | 2 | Se cobra en el 1.er intento; el reintento 2 confirma sin cobrar otra vez |

`scripts\limpiar-datos.bat` borra el historial.

## Puertos y variables de entorno

| Servicio | Puerto | Variables |
|---|---|---|
| API Gateway | 8000 | `PORT`, `CONSUL_ADDR`, `GATEWAY_SECRET` |
| Suscripciones | 8084 | `PUERTO`, `CONSUL_ADDR`, `SERVICE_HOST`, `GATEWAY_SECRET`, `MODO_FALLO`, `CANTIDAD_FALLOS`, `RETARDO_FALLO_MS`, `DIRECTORIO_DATOS` |
| Reacciones | 8085 | `PUERTO`, `CONSUL_ADDR`, `SERVICE_HOST`, `GATEWAY_SECRET`, `EXIGIR_GATEWAY` |
| Envoy | 8080 | — |
| Streaming | 50051 | — |
| Consul | 8500 | — |

El secreto por defecto (`secreto-gateway-2026`) es el mismo del laboratorio GUIA_9.

## Si modifica el receptor de fragmentos

```
cd ClienteWeb
npm install
npx webpack
```

Esto regenera `bundle.js`. La URL de cada servicio está en `ClienteWeb/js/config.js`; para ir directo a
Envoy sin el gateway, cambie `URL_STREAMING` a `http://localhost:8080`.
