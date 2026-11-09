// Package capaaccesodatos guarda las reproducciones recibidas en datos/reproducciones.json.
package capaaccesodatos

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"estadisticas/dtos"
)

// RepositorioEstadisticas almacena el historial de reproducciones (persistido en JSON).
type RepositorioEstadisticas struct {
	mu             sync.Mutex
	ruta           string
	reproducciones []dtos.NotificacionReproduccionDTO
}

var (
	instancia *RepositorioEstadisticas
	once      sync.Once
)

// GetRepositorioEstadisticas aplica el patrón Singleton y carga el archivo JSON.
func GetRepositorioEstadisticas() *RepositorioEstadisticas {
	once.Do(func() {
		directorio := os.Getenv("DIRECTORIO_DATOS")
		if directorio == "" {
			directorio = "./datos"
		}
		instancia = &RepositorioEstadisticas{ruta: filepath.Join(directorio, "reproducciones.json")}
		if contenido, err := os.ReadFile(instancia.ruta); err == nil && len(contenido) > 0 {
			_ = json.Unmarshal(contenido, &instancia.reproducciones)
		}
	})
	return instancia
}

// RegistrarReproduccion guarda la reproducción y persiste el archivo.
func (r *RepositorioEstadisticas) RegistrarReproduccion(reproduccion dtos.NotificacionReproduccionDTO) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.reproducciones = append(r.reproducciones, reproduccion)
	if err := os.MkdirAll(filepath.Dir(r.ruta), 0o755); err != nil {
		return err
	}
	contenido, err := json.MarshalIndent(r.reproducciones, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(r.ruta, contenido, 0o644)
}

// ObtenerResumen retorna el total y los conteos generales ordenados de mayor a menor.
func (r *RepositorioEstadisticas) ObtenerResumen() dtos.ResumenEstadisticasDTO {
	r.mu.Lock()
	defer r.mu.Unlock()
	porTipo, porAudio := map[string]int{}, map[string]int{}
	for _, rep := range r.reproducciones {
		porTipo[rep.Tipo]++
		porAudio[rep.Titulo]++
	}
	return dtos.ResumenEstadisticasDTO{
		TotalReproducciones: len(r.reproducciones),
		PorTipo:             ordenarConteos(porTipo),
		PorAudio:            ordenarConteos(porAudio),
	}
}

// HistorialDeCliente agrupa por audio las reproducciones de un cliente.
func (r *RepositorioEstadisticas) HistorialDeCliente(cliente string) dtos.HistorialClienteDTO {
	r.mu.Lock()
	defer r.mu.Unlock()
	porAudio := map[int32]*dtos.AudioReproducidoDTO{}
	total := 0
	for _, rep := range r.reproducciones {
		if !strings.EqualFold(rep.Cliente, cliente) {
			continue
		}
		total++
		a, ok := porAudio[rep.IdAudio]
		if !ok {
			a = &dtos.AudioReproducidoDTO{IdAudio: rep.IdAudio, Titulo: rep.Titulo, Tipo: rep.Tipo}
			porAudio[rep.IdAudio] = a
		}
		a.Cantidad++
		if rep.FechaHora > a.UltimaReproduccion {
			a.UltimaReproduccion = rep.FechaHora
		}
	}
	audios := make([]dtos.AudioReproducidoDTO, 0, len(porAudio))
	for _, a := range porAudio {
		audios = append(audios, *a)
	}
	sort.Slice(audios, func(i, j int) bool {
		if audios[i].Cantidad != audios[j].Cantidad {
			return audios[i].Cantidad > audios[j].Cantidad
		}
		return audios[i].Titulo < audios[j].Titulo
	})
	return dtos.HistorialClienteDTO{Cliente: cliente, TotalReproducciones: total, Audios: audios}
}

func ordenarConteos(conteos map[string]int) []dtos.ConteoDTO {
	lista := make([]dtos.ConteoDTO, 0, len(conteos))
	for nombre, cantidad := range conteos {
		lista = append(lista, dtos.ConteoDTO{Nombre: nombre, Cantidad: cantidad})
	}
	sort.Slice(lista, func(i, j int) bool {
		if lista[i].Cantidad != lista[j].Cantidad {
			return lista[i].Cantidad > lista[j].Cantidad
		}
		return lista[i].Nombre < lista[j].Nombre
	})
	return lista
}
