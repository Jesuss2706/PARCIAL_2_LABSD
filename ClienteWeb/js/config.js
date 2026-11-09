// Configuración del cliente web. Todo pasa por el API Gateway.
const URL_GATEWAY = "http://localhost:8000";

const CONFIG = {
  // gRPC-Web (el gateway lo reenvía al Proxy Envoy -> servidor de streaming)
  URL_GATEWAY,
  // REST: audios y metadatos, suscripciones, estadísticas
  URL_API: URL_GATEWAY + "/api",
  // WebSocket (STOMP) del servidor de reacciones
  URL_WS: URL_GATEWAY.replace(/^http/, "ws") + "/ws-stomp",

  // Pagos: si el servidor no responde en este tiempo el intento se considera fallido...
  TIEMPO_ESPERA_MS: 3000,
  // ...y se reintenta con el MISMO token tras estas esperas (backoff exponencial).
  ESPERAS_REINTENTO_MS: [2000, 4000, 8000, 16000],
};
