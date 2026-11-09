// Package dtos contiene los objetos que llegan por la cola y los que se exponen por REST.
package dtos

// NotificacionReproduccionDTO debe coincidir con el DTO que publica el servidor de streaming.
type NotificacionReproduccionDTO struct {
	IdAudio   int32  `json:"idAudio"`
	Titulo    string `json:"titulo"`
	Tipo      string `json:"tipo"`
	Archivo   string `json:"archivo"`
	Cliente   string `json:"cliente"`
	FechaHora string `json:"fechaHora"`
	Mensaje   string `json:"mensaje"`
}

// ConteoDTO es un par nombre/cantidad.
type ConteoDTO struct {
	Nombre   string `json:"nombre"`
	Cantidad int    `json:"cantidad"`
}

// ResumenEstadisticasDTO son los acumulados generales.
type ResumenEstadisticasDTO struct {
	TotalReproducciones int         `json:"totalReproducciones"`
	PorTipo             []ConteoDTO `json:"porTipo"`
	PorAudio            []ConteoDTO `json:"porAudio"`
}

// AudioReproducidoDTO es un audio que reprodujo un cliente y cuántas veces lo hizo.
type AudioReproducidoDTO struct {
	IdAudio            int32  `json:"idAudio"`
	Titulo             string `json:"titulo"`
	Tipo               string `json:"tipo"`
	Cantidad           int    `json:"cantidad"`
	UltimaReproduccion string `json:"ultimaReproduccion"`
}

// HistorialClienteDTO es la respuesta de GET /estadisticas/clientes/:cliente.
type HistorialClienteDTO struct {
	Cliente             string                `json:"cliente"`
	TotalReproducciones int                   `json:"totalReproducciones"`
	Audios              []AudioReproducidoDTO `json:"audios"`
}
