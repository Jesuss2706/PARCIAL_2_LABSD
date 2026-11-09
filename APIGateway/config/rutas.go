// Package config contiene la tabla de enrutamiento del API Gateway.
package config

// Ruta asocia un prefijo público del gateway con el nombre de un servicio en Consul.
type Ruta struct {
	Prefijo  string // ruta pública en el gateway, p. ej. /api/audios
	Servicio string // nombre con el que el microservicio se registró en Consul
	Destino  string // ruta interna en el microservicio, p. ej. /audios
}

// Rutas hacia los microservicios registrados en el Directorio Consul.
var Rutas = []Ruta{
	// Servidor de audios y metadatos (tipos, metadatos por tipo, audios, registro de audios)
	{Prefijo: "/api/tipos", Servicio: "servidor-audios", Destino: "/tipos"},
	{Prefijo: "/api/audios", Servicio: "servidor-audios", Destino: "/audios"},
	// Servidor de suscripciones (planes, tokens idempotentes, pagos)
	{Prefijo: "/api/suscripciones", Servicio: "servidor-suscripciones", Destino: "/suscripciones"},
	// Servidor de estadísticas (audios reproducidos por cliente)
	{Prefijo: "/api/estadisticas", Servicio: "servidor-estadisticas", Destino: "/estadisticas"},
	// Servidor de reacciones (WebSocket + STOMP). /ws-stomp es WebSocket nativo, /ws es SockJS.
	{Prefijo: "/ws-stomp", Servicio: "servidor-reacciones", Destino: "/ws-stomp"},
	{Prefijo: "/ws", Servicio: "servidor-reacciones", Destino: "/ws"},
}

// PrefijoGrpcWeb es el prefijo de las llamadas gRPC-Web (paquete.Servicio del audio.proto).
// Estas peticiones no van a Consul: se reenvían al Proxy Envoy, que las traduce a gRPC.
const PrefijoGrpcWeb = "/servicios.AudioService"
