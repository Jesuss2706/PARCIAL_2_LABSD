// Package configuracion define las rutas públicas del API Gateway.
package configuracion

// Ruta asocia un prefijo público del gateway con un servicio.
//   - Si Servicio no está vacío, la dirección se resuelve en el directorio Consul.
//   - Si URLFija no está vacía, se usa esa dirección (ej. el Proxy Envoy, que no se registra en Consul).
type Ruta struct {
	Prefijo  string // ruta pública en el gateway, ej. /api/suscripciones
	Servicio string // nombre con el que el microservicio se registró en Consul
	URLFija  string // dirección fija cuando el destino no está en Consul
	Destino  string // prefijo interno en el servicio, ej. /suscripciones
}

// Rutas del sistema (ver diagrama de contenedores del segundo corte).
var Rutas = []Ruta{
	// Microservicios implementados en este laboratorio
	{Prefijo: "/api/suscripciones", Servicio: "servidor-suscripciones", Destino: "/suscripciones"},
	{Prefijo: "/api/reacciones", Servicio: "servidor-reacciones", Destino: "/reacciones"},

	// Cliente web -> API Gateway -> Proxy Envoy (gRPC-Web) -> Servidor de Streaming
	{Prefijo: "/api/streaming", URLFija: "http://127.0.0.1:8080", Destino: ""},

	// Servicios restantes del diagrama (se resuelven cuando se registren en Consul)
	{Prefijo: "/api/audios", Servicio: "servidor-audios-metadatos", Destino: "/audios"},
	{Prefijo: "/api/estadisticas", Servicio: "servidor-estadisticas", Destino: "/estadisticas"},
}
