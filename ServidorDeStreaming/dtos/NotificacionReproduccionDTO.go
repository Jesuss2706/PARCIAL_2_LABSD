// Package dtos contiene los objetos que viajan en JSON por la cola de mensajes.
package dtos

// NotificacionReproduccionDTO son los datos de una reproducción que se envían
// asíncronamente al servidor de estadísticas.
type NotificacionReproduccionDTO struct {
	IdAudio   int32  `json:"idAudio"`
	Titulo    string `json:"titulo"`
	Tipo      string `json:"tipo"`
	Archivo   string `json:"archivo"`
	Cliente   string `json:"cliente"`
	FechaHora string `json:"fechaHora"`
	Mensaje   string `json:"mensaje"`
}
