// Configuración del cliente web. Todas las peticiones pasan por el API Gateway.
window.CONFIG = {
  URL_GATEWAY: "http://localhost:8000",

  // gRPC-Web: Gateway -> Proxy Envoy (8080) -> Servidor de Streaming (50051)
  URL_STREAMING: "http://localhost:8000/api/streaming",

  // WebSocket (STOMP sobre SockJS): Gateway -> Servidor de Reacciones
  URL_REACCIONES: "http://localhost:8000/api/reacciones/ws",

  // REST: Gateway -> Servidor de Suscripciones
  URL_SUSCRIPCIONES: "http://localhost:8000/api/suscripciones",

  // Pagos: si el servidor no responde en este tiempo, el intento se considera fallido.
  TIEMPO_ESPERA_MS: 3000,

  // Temporizadores de reenvío (backoff exponencial): reintento 1 a los 2 s, 2 a los 4 s, 3 a los 8 s, 4 a los 16 s.
  ESPERAS_REINTENTO_MS: [2000, 4000, 8000, 16000],

  // Si se pierde la conexión con el servidor de reacciones, se reintenta cada 5 s.
  REINTENTO_REACCIONES_MS: 5000,
};
