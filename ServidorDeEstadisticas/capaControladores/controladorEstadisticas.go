// Package capacontroladores expone por REST las estadísticas de reproducción.
package capacontroladores

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	fachada "estadisticas/capaFachada"
)

type ControladorEstadisticas struct {
	fachada *fachada.FachadaEstadisticas
}

func NuevoControladorEstadisticas(f *fachada.FachadaEstadisticas) *ControladorEstadisticas {
	return &ControladorEstadisticas{fachada: f}
}

func (c *ControladorEstadisticas) RegistrarRutas(r *gin.Engine) {
	r.GET("/estadisticas/clientes/:cliente", c.historialCliente)
	r.GET("/estadisticas/resumen", c.resumen)
}

// GET /estadisticas/clientes/:cliente -> audios reproducidos por el cliente y cuántas veces.
func (c *ControladorEstadisticas) historialCliente(ctx *gin.Context) {
	cliente := ctx.Param("cliente")
	fmt.Printf("\n[ECO REST] GET /estadisticas/clientes/%s -> historial de reproducciones\n", cliente)
	historial := c.fachada.HistorialDeCliente(cliente)
	fmt.Printf("[ECO REST] Respuesta: %d audios distintos, %d reproducciones\n", len(historial.Audios), historial.TotalReproducciones)
	ctx.JSON(http.StatusOK, historial)
}

// GET /estadisticas/resumen -> acumulados generales.
func (c *ControladorEstadisticas) resumen(ctx *gin.Context) {
	fmt.Printf("\n[ECO REST] GET /estadisticas/resumen\n")
	ctx.JSON(http.StatusOK, c.fachada.Resumen())
}
