// Package capaFachadaServices contiene la lógica de negocio del servidor de audios y metadatos.
package capaFachadaServices

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode"

	"golang.org/x/text/unicode/norm"

	"servidor-audios/capaAccesoADatos/entity"
	"servidor-audios/capaAccesoADatos/repository"
	"servidor-audios/capaFachadaServices/dto"
)

// prefijos del nombre del archivo según el tipo (ms_, pc_, al_, rb_), igual que en el parcial 1.
var prefijoPorTipo = map[int]string{
	repository.TipoMusica: "ms_", repository.TipoPodcast: "pc_",
	repository.TipoAudiolibro: "al_", repository.TipoRuidoBlanco: "rb_",
}

var noAlfanumerico = regexp.MustCompile(`[^a-z0-9]+`)

// AudioService es la fachada: valida, guarda el mp3 y mapea entidades <-> DTO.
type AudioService struct {
	repo         *repository.AudioRepository
	carpetaAudio string // carpeta donde el servidor de streaming lee los mp3
}

func NuevoAudioService(repo *repository.AudioRepository, carpetaAudio string) *AudioService {
	return &AudioService{repo: repo, carpetaAudio: carpetaAudio}
}

func (s *AudioService) ListarTipos() []dto.TipoAudioDTO {
	resultado := []dto.TipoAudioDTO{}
	for _, t := range s.repo.ListarTipos() {
		resultado = append(resultado, dto.TipoAudioDTO{Id: t.Id, Nombre: t.Nombre, CantidadAudios: s.repo.ContarAudiosPorTipo(t.Id)})
	}
	return resultado
}

func (s *AudioService) ListarMetadatosDeTipo(idTipo int) ([]dto.MetadatoDTO, error) {
	if _, ok := s.repo.BuscarTipoPorId(idTipo); !ok {
		return nil, errorNoEncontrado(fmt.Sprintf("el tipo de audio %d no existe", idTipo))
	}
	resultado := []dto.MetadatoDTO{}
	for _, m := range s.repo.ListarMetadatosPorTipo(idTipo) {
		resultado = append(resultado, dto.MetadatoDTO{Id: m.Id, Nombre: m.Nombre, Descripcion: m.Descripcion})
	}
	return resultado, nil
}

func (s *AudioService) ListarAudiosDeTipo(idTipo int) ([]dto.AudioResumenDTO, error) {
	tipo, ok := s.repo.BuscarTipoPorId(idTipo)
	if !ok {
		return nil, errorNoEncontrado(fmt.Sprintf("el tipo de audio %d no existe", idTipo))
	}
	resultado := []dto.AudioResumenDTO{}
	for _, a := range s.repo.ListarAudiosPorTipo(idTipo) {
		resultado = append(resultado, dto.AudioResumenDTO{Id: a.Id, Titulo: a.Titulo, Duracion: a.Duracion, IdTipo: tipo.Id, Tipo: tipo.Nombre})
	}
	return resultado, nil
}

func (s *AudioService) ObtenerDetalle(idAudio int) (dto.AudioDetalleDTO, error) {
	audio, ok := s.repo.BuscarAudioPorId(idAudio)
	if !ok {
		return dto.AudioDetalleDTO{}, errorNoEncontrado(fmt.Sprintf("el audio %d no existe", idAudio))
	}
	return s.aDetalleDTO(audio), nil
}

// RegistrarAudio valida los metadatos propios del tipo, guarda el mp3 en la carpeta
// del servidor de streaming y registra el audio en el archivo JSON.
func (s *AudioService) RegistrarAudio(sol dto.SolicitudRegistroDTO, nombreOriginal string, tamano int64, contenido io.Reader) (dto.AudioDetalleDTO, error) {
	sol.Titulo = strings.TrimSpace(sol.Titulo)
	sol.Duracion = strings.TrimSpace(sol.Duracion)

	tipo, ok := s.repo.BuscarTipoPorId(sol.IdTipo)
	if !ok {
		return dto.AudioDetalleDTO{}, errorValidacion(fmt.Sprintf("el tipo de audio %d no existe", sol.IdTipo))
	}
	if sol.Titulo == "" {
		return dto.AudioDetalleDTO{}, errorValidacion("el título del audio es obligatorio")
	}
	if !strings.EqualFold(filepath.Ext(nombreOriginal), ".mp3") {
		return dto.AudioDetalleDTO{}, errorValidacion("el archivo debe ser un .mp3")
	}
	if tamano <= 0 {
		return dto.AudioDetalleDTO{}, errorValidacion("el archivo de audio está vacío")
	}

	// Cada tipo tiene sus propios metadatos: todos son obligatorios y no se aceptan los de otro tipo.
	valores := []entity.ValorMetadato{}
	for _, m := range s.repo.ListarMetadatosPorTipo(tipo.Id) {
		valor := strings.TrimSpace(sol.Metadatos[m.Id])
		if valor == "" {
			return dto.AudioDetalleDTO{}, errorValidacion(fmt.Sprintf("el metadato '%s' es obligatorio para el tipo %s", m.Nombre, tipo.Nombre))
		}
		valores = append(valores, entity.ValorMetadato{IdMetadato: m.Id, Valor: valor})
	}
	for idMeta := range sol.Metadatos {
		if m, ok := s.repo.BuscarMetadatoPorId(idMeta); !ok || m.IdTipoAudio != tipo.Id {
			return dto.AudioDetalleDTO{}, errorValidacion(fmt.Sprintf("el metadato %d no pertenece al tipo %s", idMeta, tipo.Nombre))
		}
	}

	nombreArchivo := prefijoPorTipo[tipo.Id] + slug(sol.Titulo) + ".mp3"
	if s.repo.ExisteArchivo(nombreArchivo) {
		return dto.AudioDetalleDTO{}, errorConflicto(fmt.Sprintf("ya existe un audio registrado con el archivo %s (cambie el título)", nombreArchivo))
	}
	if err := s.guardarArchivo(nombreArchivo, contenido); err != nil {
		return dto.AudioDetalleDTO{}, errorInterno("no fue posible guardar el mp3: " + err.Error())
	}
	if sol.Duracion == "" {
		sol.Duracion = "—"
	}

	audio, err := s.repo.Guardar(entity.Audio{
		IdTipoAudio:   tipo.Id,
		Titulo:        sol.Titulo,
		NombreArchivo: nombreArchivo,
		Duracion:      sol.Duracion,
		TamanoBytes:   tamano,
		FechaRegistro: time.Now(),
		Valores:       valores,
	})
	if err != nil {
		_ = os.Remove(filepath.Join(s.carpetaAudio, nombreArchivo))
		return dto.AudioDetalleDTO{}, errorInterno("no fue posible registrar el audio: " + err.Error())
	}
	log.Printf("[REGISTRO] Audio #%d '%s' (%s) guardado como %s (%d bytes)", audio.Id, audio.Titulo, tipo.Nombre, nombreArchivo, tamano)
	return s.aDetalleDTO(audio), nil
}

// guardarArchivo escribe primero un temporal y luego lo renombra, para que el
// servidor de streaming nunca lea un mp3 a medio copiar.
func (s *AudioService) guardarArchivo(nombre string, contenido io.Reader) error {
	if err := os.MkdirAll(s.carpetaAudio, 0o755); err != nil {
		return err
	}
	destino := filepath.Join(s.carpetaAudio, nombre)
	temporal := destino + ".subiendo"
	f, err := os.Create(temporal)
	if err != nil {
		return err
	}
	if _, err := io.Copy(f, contenido); err != nil {
		f.Close()
		os.Remove(temporal)
		return err
	}
	if err := f.Close(); err != nil {
		os.Remove(temporal)
		return err
	}
	return os.Rename(temporal, destino)
}

func (s *AudioService) aDetalleDTO(a entity.Audio) dto.AudioDetalleDTO {
	tipo, _ := s.repo.BuscarTipoPorId(a.IdTipoAudio)
	metadatos := []dto.ValorMetadatoDTO{}
	for _, v := range a.Valores {
		m, _ := s.repo.BuscarMetadatoPorId(v.IdMetadato)
		metadatos = append(metadatos, dto.ValorMetadatoDTO{IdMetadato: v.IdMetadato, Nombre: m.Nombre, Valor: v.Valor})
	}
	return dto.AudioDetalleDTO{
		Id: a.Id, Titulo: a.Titulo, IdTipo: tipo.Id, Tipo: tipo.Nombre,
		NombreArchivo: a.NombreArchivo, Duracion: a.Duracion, TamanoBytes: a.TamanoBytes,
		FechaRegistro: a.FechaRegistro.Format("2006-01-02 15:04"), Metadatos: metadatos,
	}
}

// slug convierte "Canción Número 1" en "cancion_numero_1" (sin tildes ni espacios).
func slug(texto string) string {
	sinTildes := strings.Map(func(r rune) rune {
		if unicode.Is(unicode.Mn, r) {
			return -1
		}
		return r
	}, norm.NFD.String(strings.ToLower(texto)))
	resultado := strings.Trim(noAlfanumerico.ReplaceAllString(sinTildes, "_"), "_")
	if resultado == "" {
		resultado = fmt.Sprintf("audio_%d", time.Now().Unix())
	}
	if len(resultado) > 60 {
		resultado = strings.Trim(resultado[:60], "_")
	}
	return resultado
}
