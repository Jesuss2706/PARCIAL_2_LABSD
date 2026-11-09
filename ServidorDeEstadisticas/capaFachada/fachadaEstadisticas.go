// Package fachada procesa los mensajes de la cola y atiende las consultas REST.
package fachada

import (
	"encoding/json"
	"fmt"
	"strings"

	capaaccesodatos "estadisticas/capaAccesoDatos"
	"estadisticas/dtos"
	"estadisticas/vistas"
)

// FachadaEstadisticas registra las reproducciones y responde las consultas.
type FachadaEstadisticas struct {
	repo *capaaccesodatos.RepositorioEstadisticas
}

// NuevaFachadaEstadisticas crea la fachada con el repositorio singleton.
func NuevaFachadaEstadisticas() *FachadaEstadisticas {
	fmt.Println("Inicializando fachada de estadísticas ...")
	return &FachadaEstadisticas{repo: capaaccesodatos.GetRepositorioEstadisticas()}
}

// ProcesarMensaje convierte el JSON en DTO, lo almacena e imprime el resumen.
func (f *FachadaEstadisticas) ProcesarMensaje(cuerpo []byte) error {
	var reproduccion dtos.NotificacionReproduccionDTO
	if err := json.Unmarshal(cuerpo, &reproduccion); err != nil {
		return fmt.Errorf("mensaje con formato inválido: %v", err)
	}
	if strings.TrimSpace(reproduccion.Cliente) == "" {
		reproduccion.Cliente = "anonimo"
	}
	if err := f.repo.RegistrarReproduccion(reproduccion); err != nil {
		return fmt.Errorf("no se pudo almacenar la reproducción: %v", err)
	}
	vistas.MostrarReproduccion(reproduccion)
	vistas.MostrarResumen(f.repo.ObtenerResumen())
	return nil
}

// HistorialDeCliente retorna los audios que reprodujo el cliente y cuántas veces.
func (f *FachadaEstadisticas) HistorialDeCliente(cliente string) dtos.HistorialClienteDTO {
	return f.repo.HistorialDeCliente(strings.TrimSpace(cliente))
}

// Resumen retorna los acumulados generales.
func (f *FachadaEstadisticas) Resumen() dtos.ResumenEstadisticasDTO {
	return f.repo.ObtenerResumen()
}
