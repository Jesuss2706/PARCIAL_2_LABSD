// Package repository es la capa de acceso a datos del servidor de audios y metadatos.
package repository

import (
	"sort"
	"sync"
	"time"

	"servidor-audios/capaAccesoADatos/entity"
)

// Identificadores de los tipos de audio.
const (
	TipoMusica      = 1
	TipoPodcast     = 2
	TipoAudiolibro  = 3
	TipoRuidoBlanco = 4
)

// Ids de metadatos agrupados por tipo: 1xx Música, 2xx Podcasts, 3xx Audiolibros, 4xx Ruido blanco.
const (
	metaArtista  = 101
	metaAlbum    = 102
	metaGenero   = 103
	metaPodcast  = 201
	metaEpisodio = 202
	metaHost     = 203
	metaLibro    = 301
	metaAutor    = 302
	metaNarrador = 303
	metaSonido   = 401
	metaFuente   = 402
	metaUso      = 403
)

// AudioRepository guarda los tipos y metadatos (fijos, definidos en el requerimiento)
// y los audios registrados (persistidos en datos/audios.json).
type AudioRepository struct {
	mu        sync.RWMutex
	ruta      string
	tipos     []entity.TipoAudio
	metadatos []entity.Metadato
	audios    []entity.Audio
}

// NuevoAudioRepository carga los audios del archivo JSON; si no existe, crea los audios de ejemplo.
func NuevoAudioRepository(ruta string) (*AudioRepository, error) {
	r := &AudioRepository{ruta: ruta}
	r.cargarTipos()
	r.cargarMetadatos()
	existe, err := cargarJSON(ruta, &r.audios)
	if err != nil {
		return nil, err
	}
	if !existe {
		r.cargarAudiosDeEjemplo()
		if err := guardarJSON(ruta, r.audios); err != nil {
			return nil, err
		}
	}
	return r, nil
}

func (r *AudioRepository) ListarTipos() []entity.TipoAudio {
	return append([]entity.TipoAudio(nil), r.tipos...)
}

func (r *AudioRepository) BuscarTipoPorId(id int) (entity.TipoAudio, bool) {
	for _, t := range r.tipos {
		if t.Id == id {
			return t, true
		}
	}
	return entity.TipoAudio{}, false
}

func (r *AudioRepository) ListarMetadatosPorTipo(idTipo int) []entity.Metadato {
	resultado := []entity.Metadato{}
	for _, m := range r.metadatos {
		if m.IdTipoAudio == idTipo {
			resultado = append(resultado, m)
		}
	}
	return resultado
}

func (r *AudioRepository) BuscarMetadatoPorId(id int) (entity.Metadato, bool) {
	for _, m := range r.metadatos {
		if m.Id == id {
			return m, true
		}
	}
	return entity.Metadato{}, false
}

func (r *AudioRepository) ListarAudiosPorTipo(idTipo int) []entity.Audio {
	r.mu.RLock()
	defer r.mu.RUnlock()
	resultado := []entity.Audio{}
	for _, a := range r.audios {
		if a.IdTipoAudio == idTipo {
			resultado = append(resultado, a)
		}
	}
	return resultado
}

func (r *AudioRepository) ContarAudiosPorTipo(idTipo int) int {
	return len(r.ListarAudiosPorTipo(idTipo))
}

func (r *AudioRepository) BuscarAudioPorId(id int) (entity.Audio, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, a := range r.audios {
		if a.Id == id {
			return a, true
		}
	}
	return entity.Audio{}, false
}

// ExisteArchivo indica si ya hay un audio registrado con ese nombre de archivo.
func (r *AudioRepository) ExisteArchivo(nombreArchivo string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, a := range r.audios {
		if a.NombreArchivo == nombreArchivo {
			return true
		}
	}
	return false
}

// Guardar asigna el siguiente id al audio, lo agrega y persiste el archivo JSON.
func (r *AudioRepository) Guardar(audio entity.Audio) (entity.Audio, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	maximo := 0
	for _, a := range r.audios {
		if a.Id > maximo {
			maximo = a.Id
		}
	}
	audio.Id = maximo + 1
	r.audios = append(r.audios, audio)
	sort.Slice(r.audios, func(i, j int) bool { return r.audios[i].Id < r.audios[j].Id })
	if err := guardarJSON(r.ruta, r.audios); err != nil {
		r.audios = r.audios[:len(r.audios)-1]
		return entity.Audio{}, err
	}
	return audio, nil
}

// ---------------------------------------------------------------------------
// Datos fijos: tipos y metadatos propios de cada tipo (según el requerimiento)
// ---------------------------------------------------------------------------

func (r *AudioRepository) cargarTipos() {
	r.tipos = []entity.TipoAudio{
		{Id: TipoMusica, Nombre: "Música"},
		{Id: TipoPodcast, Nombre: "Podcasts"},
		{Id: TipoAudiolibro, Nombre: "Audiolibros"},
		{Id: TipoRuidoBlanco, Nombre: "Ruido Blanco"},
	}
}

func (r *AudioRepository) cargarMetadatos() {
	r.metadatos = []entity.Metadato{
		{Id: metaArtista, IdTipoAudio: TipoMusica, Nombre: "Artista principal", Descripcion: "El intérprete o banda líder del proyecto"},
		{Id: metaAlbum, IdTipoAudio: TipoMusica, Nombre: "Álbum", Descripcion: "El nombre del disco, EP o sencillo que agrupa la obra"},
		{Id: metaGenero, IdTipoAudio: TipoMusica, Nombre: "Género musical", Descripcion: "La categoría rítmica o cultural (Pop, Rock, Jazz)"},

		{Id: metaPodcast, IdTipoAudio: TipoPodcast, Nombre: "Nombre del podcast", Descripcion: "El título general del programa o serie"},
		{Id: metaEpisodio, IdTipoAudio: TipoPodcast, Nombre: "Título del episodio", Descripcion: "El nombre del tema específico de ese audio"},
		{Id: metaHost, IdTipoAudio: TipoPodcast, Nombre: "Anfitrión (Host)", Descripcion: "La persona que dirige o presenta el contenido"},

		{Id: metaLibro, IdTipoAudio: TipoAudiolibro, Nombre: "Título del libro", Descripcion: "El nombre de la obra literaria completa"},
		{Id: metaAutor, IdTipoAudio: TipoAudiolibro, Nombre: "Autor", Descripcion: "El escritor original del texto"},
		{Id: metaNarrador, IdTipoAudio: TipoAudiolibro, Nombre: "Narrador", Descripcion: "El actor de voz que realiza la lectura"},

		{Id: metaSonido, IdTipoAudio: TipoRuidoBlanco, Nombre: "Tipo de sonido", Descripcion: "La variante de frecuencia (Ruido Blanco, Marrón, Rosa)"},
		{Id: metaFuente, IdTipoAudio: TipoRuidoBlanco, Nombre: "Fuente del audio", Descripcion: "El origen del sonido (Lluvia, Ventilador, Bosque)"},
		{Id: metaUso, IdTipoAudio: TipoRuidoBlanco, Nombre: "Uso sugerido", Descripcion: "La intención del audio (Dormir, Concentración, Meditación)"},
	}
}

// cargarAudiosDeEjemplo se usa solo la primera vez (cuando no existe datos/audios.json).
// Los mp3 de ejemplo vienen en la carpeta audios del servidor de streaming.
func (r *AudioRepository) cargarAudiosDeEjemplo() {
	ahora := time.Now()
	r.audios = []entity.Audio{
		{Id: 1, IdTipoAudio: TipoMusica, Titulo: "Melodía de prueba", NombreArchivo: "ms_melodia_de_prueba.mp3", Duracion: "0:30", FechaRegistro: ahora,
			Valores: []entity.ValorMetadato{{IdMetadato: metaArtista, Valor: "Laboratorio SD"}, {IdMetadato: metaAlbum, Valor: "Demos"}, {IdMetadato: metaGenero, Valor: "Electrónica"}}},
		{Id: 2, IdTipoAudio: TipoPodcast, Titulo: "Episodio piloto", NombreArchivo: "pc_episodio_piloto.mp3", Duracion: "0:30", FechaRegistro: ahora,
			Valores: []entity.ValorMetadato{{IdMetadato: metaPodcast, Valor: "Sistemas Distribuidos al Día"}, {IdMetadato: metaEpisodio, Valor: "Episodio piloto"}, {IdMetadato: metaHost, Valor: "Monitor del laboratorio"}}},
		{Id: 3, IdTipoAudio: TipoAudiolibro, Titulo: "Capítulo de muestra", NombreArchivo: "al_capitulo_de_muestra.mp3", Duracion: "0:30", FechaRegistro: ahora,
			Valores: []entity.ValorMetadato{{IdMetadato: metaLibro, Valor: "Relatos de prueba"}, {IdMetadato: metaAutor, Valor: "Autor anónimo"}, {IdMetadato: metaNarrador, Valor: "Voz sintética"}}},
		{Id: 4, IdTipoAudio: TipoRuidoBlanco, Titulo: "Ruido blanco para concentrarse", NombreArchivo: "rb_ruido_blanco.mp3", Duracion: "0:30", FechaRegistro: ahora,
			Valores: []entity.ValorMetadato{{IdMetadato: metaSonido, Valor: "Ruido Blanco"}, {IdMetadato: metaFuente, Valor: "Ruido generado"}, {IdMetadato: metaUso, Valor: "Concentración"}}},
		{Id: 5, IdTipoAudio: TipoRuidoBlanco, Titulo: "Ruido marrón para dormir", NombreArchivo: "rb_ruido_marron.mp3", Duracion: "0:30", FechaRegistro: ahora,
			Valores: []entity.ValorMetadato{{IdMetadato: metaSonido, Valor: "Ruido Marrón"}, {IdMetadato: metaFuente, Valor: "Ruido generado"}, {IdMetadato: metaUso, Valor: "Dormir"}}},
	}
}
