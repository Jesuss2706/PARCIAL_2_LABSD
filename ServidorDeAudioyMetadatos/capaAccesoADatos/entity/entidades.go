// Package entity contiene las entidades que se guardan en la capa de acceso a datos.
package entity

import "time"

// TipoAudio es una categoría de audio (Música, Podcasts, Audiolibros, Ruido Blanco).
type TipoAudio struct {
	Id     int    `json:"id"`
	Nombre string `json:"nombre"`
}

// Metadato define un campo propio de UN tipo de audio (p. ej. "Artista principal" -> Música).
type Metadato struct {
	Id          int    `json:"id"`
	IdTipoAudio int    `json:"idTipoAudio"`
	Nombre      string `json:"nombre"`
	Descripcion string `json:"descripcion"`
}

// ValorMetadato asocia un valor concreto a un metadato dentro de un audio.
type ValorMetadato struct {
	IdMetadato int    `json:"idMetadato"`
	Valor      string `json:"valor"`
}

// Audio es un audio registrado. NombreArchivo es el mp3 que el cliente pide al servidor de streaming.
type Audio struct {
	Id            int             `json:"id"`
	IdTipoAudio   int             `json:"idTipoAudio"`
	Titulo        string          `json:"titulo"`
	NombreArchivo string          `json:"nombreArchivo"`
	Duracion      string          `json:"duracion"`
	TamanoBytes   int64           `json:"tamanoBytes"`
	FechaRegistro time.Time       `json:"fechaRegistro"`
	Valores       []ValorMetadato `json:"valores"`
}
