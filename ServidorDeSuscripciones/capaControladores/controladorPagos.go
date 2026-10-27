package capaControladores

import (
	"fmt"
	"net/http"

	"servidor.local/suscripciones/capaFachada"
	"servidor.local/suscripciones/dtos"
)

// ControladorPagos expone los servicios de tokens, pagos y transacciones.
type ControladorPagos struct {
	fachada         *capaFachada.FachadaPagos
	simuladorFallos func(http.HandlerFunc) http.HandlerFunc
}

// NuevoControladorPagos recibe el simulador de fallos, que solo se aplica al servicio de pago.
func NuevoControladorPagos(fachada *capaFachada.FachadaPagos, simuladorFallos func(http.HandlerFunc) http.HandlerFunc) *ControladorPagos {
	return &ControladorPagos{fachada: fachada, simuladorFallos: simuladorFallos}
}

// RegistrarRutas asocia los endpoints con sus manejadores.
func (c *ControladorPagos) RegistrarRutas(enrutador *http.ServeMux) {
	enrutador.HandleFunc("POST /suscripciones/tokens", c.generarToken)
	enrutador.HandleFunc("POST /suscripciones/pagos", c.simuladorFallos(c.realizarPago))
	enrutador.HandleFunc("GET /suscripciones/transacciones/{cedula}", c.consultarTransacciones)
}

// POST /suscripciones/tokens: servicio 1, generar token de idempotencia.
func (c *ControladorPagos) generarToken(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("\n[ECO REST] Invocado: POST /suscripciones/tokens -> generar token de idempotencia\n")
	var solicitud dtos.SolicitudTokenDTO
	if !leerJSON(w, r, &solicitud) {
		return
	}
	respuesta, err := c.fachada.GenerarToken(solicitud)
	if err != nil {
		responderError(w, err)
		return
	}
	fmt.Printf("[ECO REST] Respuesta: token %s enviado al cliente\n", respuesta.TokenIdempotencia)
	responderJSON(w, http.StatusCreated, respuesta)
}

// POST /suscripciones/pagos: servicio 2, realizar pago con token de idempotencia.
func (c *ControladorPagos) realizarPago(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("\n[ECO REST] Invocado: POST /suscripciones/pagos -> realizar pago\n")
	var solicitud dtos.SolicitudPagoDTO
	if !leerJSON(w, r, &solicitud) {
		return
	}
	respuesta, err := c.fachada.RealizarPago(solicitud)
	if err != nil {
		responderError(w, err)
		return
	}
	fmt.Printf("[ECO REST] Respuesta: transacción_realizada=%t, pago_almacenado=%t, previamente_realizada=%t\n",
		respuesta.TransaccionRealizada, respuesta.PagoAlmacenado, respuesta.PreviamenteRealizada)
	responderJSON(w, http.StatusOK, respuesta)
}

// GET /suscripciones/transacciones/{cedula}: servicio 3, consultar historial.
func (c *ControladorPagos) consultarTransacciones(w http.ResponseWriter, r *http.Request) {
	cedula := r.PathValue("cedula")
	fmt.Printf("\n[ECO REST] Invocado: GET /suscripciones/transacciones/%s -> consultar historial\n", cedula)
	respuesta, err := c.fachada.ConsultarTransacciones(cedula)
	if err != nil {
		responderError(w, err)
		return
	}
	fmt.Printf("[ECO REST] Respuesta: %d transacciones de la cédula %s\n", respuesta.TotalTransacciones, cedula)
	responderJSON(w, http.StatusOK, respuesta)
}
