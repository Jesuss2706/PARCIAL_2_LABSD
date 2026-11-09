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
// NombreArchivo debe coincidir EXACTAMENTE con el mp3 que está en ServidorDeStreaming/audios.
func (r *AudioRepository) cargarAudiosDeEjemplo() {
	ahora := time.Now()
	audio := func(id, tipo int, titulo, archivo, duracion string, valores ...entity.ValorMetadato) entity.Audio {
		return entity.Audio{Id: id, IdTipoAudio: tipo, Titulo: titulo, NombreArchivo: archivo, Duracion: duracion, FechaRegistro: ahora, Valores: valores}
	}
	v := func(idMetadato int, valor string) entity.ValorMetadato {
		return entity.ValorMetadato{IdMetadato: idMetadato, Valor: valor}
	}

	r.audios = []entity.Audio{
		// ----------------------------- Música -----------------------------
		audio(1, TipoMusica, "Ayer y Hoy - Jean Carlos Centeno", "ms_ayer_y_hoy.mp3", "5:02",
			v(metaArtista, "Jean Carlos Centeno y La Combinación Vallenata"), v(metaAlbum, "No especificado"), v(metaGenero, "Vallenato")),
		audio(2, TipoMusica, "Creep - Radiohead", "ms_creep.mp3", "3:56",
			v(metaArtista, "Radiohead"), v(metaAlbum, "Pablo Honey"), v(metaGenero, "Rock alternativo")),
		audio(3, TipoMusica, "Las Noches - Junior H", "ms_las_noches.mp3", "3:46",
			v(metaArtista, "Junior H"), v(metaAlbum, "$ad Boyz 4 Life II"), v(metaGenero, "Regional mexicano")),
		audio(4, TipoMusica, "Mojabi Ghost - Tainy y Bad Bunny", "ms_mojabi_ghost.mp3", "3:54",
			v(metaArtista, "Tainy y Bad Bunny"), v(metaAlbum, "Data"), v(metaGenero, "Synth-pop / Electropop")),
		audio(5, TipoMusica, "Piénsalo - Junior H", "ms_piensalo.mp3", "4:01",
			v(metaArtista, "Junior H"), v(metaAlbum, "$ad Boyz 4 Life II"), v(metaGenero, "Regional mexicano")),
		audio(6, TipoMusica, "Se Acabó el Amor - Adolescent's Orquesta", "ms_se_acabo_el_amor.mp3", "4:05",
			v(metaArtista, "Adolescent's Orquesta"), v(metaAlbum, "No especificado"), v(metaGenero, "Salsa")),

		// ----------------------------- Podcasts -----------------------------
		audio(7, TipoPodcast, "The Wild Project - ¿Existió realmente Jesús?", "pc_existio_realmente_jesus.mp3", "17:56",
			v(metaPodcast, "The Wild Project"), v(metaEpisodio, "¿Existió realmente Jesús?"), v(metaHost, "Jordi Wild")),
		audio(8, TipoPodcast, "The Wild Project - ¿El tiempo existe o es una ilusión humana?", "pc_el_tiempo_existe.mp3", "14:09",
			v(metaPodcast, "The Wild Project"), v(metaEpisodio, "¿El tiempo existe o es una ilusión humana?"), v(metaHost, "Jordi Wild")),
		audio(9, TipoPodcast, "The Wild Project - ¿Fue Hitler un loco o no?", "pc_fue_hitler_un_loco.mp3", "6:48",
			v(metaPodcast, "The Wild Project"), v(metaEpisodio, "¿Fue Hitler un loco o no?"), v(metaHost, "Jordi Wild")),
		audio(10, TipoPodcast, "Learn English with Podcasts - Tom Holland", "pc_learn_english_tom_holland.mp3", "16:02",
			v(metaPodcast, "Learn English with Podcasts"), v(metaEpisodio, "Why Tom Holland DELETED His Instagram"), v(metaHost, "Learn English With TV Series")),

		// ----------------------------- Audiolibros -----------------------------
		audio(11, TipoAudiolibro, "Beatriz (Una palabra enorme)", "al_beatriz_una_palabra_enorme.mp3", "8:59",
			v(metaLibro, "Beatriz (Una palabra enorme)"), v(metaAutor, "Mario Benedetti"), v(metaNarrador, "No especificado")),
		audio(12, TipoAudiolibro, "El retrato oval", "al_el_retrato_oval.mp3", "17:09",
			v(metaLibro, "El retrato oval"), v(metaAutor, "Edgar Allan Poe"), v(metaNarrador, "No especificado")),
		audio(13, TipoAudiolibro, "El verdadero significado de la vida", "al_el_verdadero_significado_de_la_vida.mp3", "11:57",
			v(metaLibro, "El verdadero significado de la vida"), v(metaAutor, "León Tolstói"), v(metaNarrador, "No especificado")),
		audio(14, TipoAudiolibro, "Un día de estos", "al_un_dia_de_estos.mp3", "12:46",
			v(metaLibro, "Un día de estos"), v(metaAutor, "Gabriel García Márquez"), v(metaNarrador, "No especificado")),

		// ----------------------------- Ruido blanco -----------------------------
		audio(15, TipoRuidoBlanco, "15 minutos de ruido blanco para relajar", "rb_ruido_blanco_relajar.mp3", "14:58",
			v(metaSonido, "Ruido Blanco"), v(metaFuente, "Ruido generado"), v(metaUso, "Relajación")),
		audio(16, TipoRuidoBlanco, "Clair de Lune (slowed)", "rb_clair_de_lune_slowed.mp3", "7:34",
			v(metaSonido, "Música ambiental"), v(metaFuente, "Piano (Claude Debussy, versión ralentizada)"), v(metaUso, "Concentración")),
		audio(17, TipoRuidoBlanco, "Dormir rápido en 10 minutos", "rb_dormir_rapido.mp3", "10:00",
			v(metaSonido, "Sonido relajante"), v(metaFuente, "Ambiente para dormir"), v(metaUso, "Dormir")),
		audio(18, TipoRuidoBlanco, "Sonido blanco para meditación profunda (432 Hz)", "rb_meditacion_432hz.mp3", "12:30",
			v(metaSonido, "Ruido Blanco"), v(metaFuente, "Tono de 432 Hz"), v(metaUso, "Meditación")),
	}
}
