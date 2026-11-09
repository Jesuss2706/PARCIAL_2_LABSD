// Package dto contiene los objetos que viajan en JSON entre el cliente y el servidor.
package dto

type TipoAudioDTO struct {
	Id             int    `json:"id"`
	Nombre         string `json:"nombre"`
	CantidadAudios int    `json:"cantidadAudios"`
}

type MetadatoDTO struct {
	Id          int    `json:"id"`
	Nombre      string `json:"nombre"`
	Descripcion string `json:"descripcion"`
}

type AudioResumenDTO struct {
	Id       int    `json:"id"`
	Titulo   string `json:"titulo"`
	Duracion string `json:"duracion"`
	IdTipo   int    `json:"idTipo"`
	Tipo     string `json:"tipo"`
}

type ValorMetadatoDTO struct {
	IdMetadato int    `json:"idMetadato"`
	Nombre     string `json:"nombre"`
	Valor      string `json:"valor"`
}

type AudioDetalleDTO struct {
	Id            int                `json:"id"`
	Titulo        string             `json:"titulo"`
	IdTipo        int                `json:"idTipo"`
	Tipo          string             `json:"tipo"`
	NombreArchivo string             `json:"nombreArchivo"`
	Duracion      string             `json:"duracion"`
	TamanoBytes   int64              `json:"tamanoBytes"`
	FechaRegistro string             `json:"fechaRegistro"`
	Metadatos     []ValorMetadatoDTO `json:"metadatos"`
}

// SolicitudRegistroDTO son los campos de texto del formulario multipart del administrador.
type SolicitudRegistroDTO struct {
	IdTipo    int
	Titulo    string
	Duracion  string
	Metadatos map[int]string // idMetadato -> valor
}

type ErrorDTO struct {
	Error string `json:"error"`
}
