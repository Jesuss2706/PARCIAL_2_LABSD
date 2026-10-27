package capaControladores

import (
	"fmt"
	"net/http"

	"servidor.local/suscripciones/capaFachada"
	"servidor.local/suscripciones/configuracion"
	"servidor.local/suscripciones/dtos"
)

// ControladorPlanes expone el catálogo de planes y el estado del servidor.
type ControladorPlanes struct {
	fachada *capaFachada.FachadaPlanes
	cfg     configuracion.Configuracion
}

// NuevoControladorPlanes crea el controlador.
func NuevoControladorPlanes(fachada *capaFachada.FachadaPlanes, cfg configuracion.Configuracion) *ControladorPlanes {
	return &ControladorPlanes{fachada: fachada, cfg: cfg}
}

// RegistrarRutas asocia los endpoints con sus manejadores.
func (c *ControladorPlanes) RegistrarRutas(enrutador *http.ServeMux) {
	enrutador.HandleFunc("GET /suscripciones/planes", c.listarPlanes)
	enrutador.HandleFunc("GET /suscripciones/estado", c.consultarEstado)
}

// GET /suscripciones/planes
func (c *ControladorPlanes) listarPlanes(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("\n[ECO REST] Invocado: GET /suscripciones/planes -> listar planes\n")
	planes := c.fachada.ListarPlanes()
	fmt.Printf("[ECO REST] Respuesta: %d planes enviados\n", len(planes))
	responderJSON(w, http.StatusOK, planes)
}

// GET /suscripciones/estado: configuración de simulación activa (se muestra en el cliente).
func (c *ControladorPlanes) consultarEstado(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("\n[ECO REST] Invocado: GET /suscripciones/estado\n")
	responderJSON(w, http.StatusOK, dtos.EstadoDTO{
		Estado:         "activo",
		ModoFallo:      c.cfg.ModoFallo,
		CantidadFallos: c.cfg.CantidadFallos,
		RetardoFalloMs: c.cfg.RetardoFallo.Milliseconds(),
	})
}
