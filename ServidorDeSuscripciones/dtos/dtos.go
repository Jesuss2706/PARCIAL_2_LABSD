// Package dtos contiene los objetos que viajan en formato JSON entre cliente y servidor.
package dtos

// PlanDTO describe un plan de suscripción.
type PlanDTO struct {
	Id          string  `json:"id"`
	Nombre      string  `json:"nombre"`
	Descripcion string  `json:"descripcion"`
	Precio      float64 `json:"precio"`
	Moneda      string  `json:"moneda"`
	Cuentas     int     `json:"cuentas"`
}

// SolicitudTokenDTO: datos del pago con los que se pide un token de idempotencia.
type SolicitudTokenDTO struct {
	Cedula        string  `json:"cedula"`
	IdPlan        string  `json:"id_plan"`
	Monto         float64 `json:"monto"`
	NombreTitular string  `json:"nombre_titular"`
	NumeroTarjeta string  `json:"numero_tarjeta"`
}

// RespuestaTokenDTO: token de idempotencia generado.
type RespuestaTokenDTO struct {
	TokenIdempotencia string  `json:"token_idempotencia"`
	Cedula            string  `json:"cedula"`
	IdPlan            string  `json:"id_plan"`
	Monto             float64 `json:"monto"`
	FechaCreacion     string  `json:"fecha_creacion"`
}

// SolicitudPagoDTO: pago que se realiza con un token de idempotencia.
type SolicitudPagoDTO struct {
	TokenIdempotencia string  `json:"token_idempotencia"`
	Monto             float64 `json:"monto"`
}

// RespuestaPagoDTO: resultado del pago.
type RespuestaPagoDTO struct {
	TokenIdempotencia      string   `json:"token_idempotencia"`
	MontoTransaccion       float64  `json:"monto_transaccion"`
	TransaccionRealizada   bool     `json:"transaccion_realizada"`
	PagoAlmacenado         bool     `json:"pago_almacenado"`
	PreviamenteRealizada   bool     `json:"previamente_realizada"`
	MontoTransaccionPrevia *float64 `json:"monto_transaccion_previa,omitempty"`
	IdTransaccion          string   `json:"id_transaccion,omitempty"`
	FechaTransaccion       string   `json:"fecha_transaccion,omitempty"`
	Mensaje                string   `json:"mensaje"`
}

// TransaccionDTO: pago del historial.
type TransaccionDTO struct {
	IdTransaccion     string  `json:"id_transaccion"`
	TokenIdempotencia string  `json:"token_idempotencia"`
	IdPlan            string  `json:"id_plan"`
	Monto             float64 `json:"monto"`
	Fecha             string  `json:"fecha"`
}

// TokenDTO: token del historial.
type TokenDTO struct {
	TokenIdempotencia string  `json:"token_idempotencia"`
	IdPlan            string  `json:"id_plan"`
	Monto             float64 `json:"monto"`
	FechaCreacion     string  `json:"fecha_creacion"`
	Utilizado         bool    `json:"utilizado"`
}

// RespuestaConsultaDTO: historial de pagos y tokens de una cédula.
type RespuestaConsultaDTO struct {
	Cedula             string           `json:"cedula"`
	PlanActual         *PlanDTO         `json:"plan_actual"`
	TotalTransacciones int              `json:"total_transacciones"`
	Transacciones      []TransaccionDTO `json:"transacciones"`
	TokensUtilizados   []TokenDTO       `json:"tokens_utilizados"`
	TokensSinUtilizar  []TokenDTO       `json:"tokens_sin_utilizar"`
}

// EstadoDTO: configuración de simulación activa en el servidor.
type EstadoDTO struct {
	Estado         string `json:"estado"`
	ModoFallo      string `json:"modo_fallo"`
	CantidadFallos int    `json:"cantidad_fallos"`
	RetardoFalloMs int64  `json:"retardo_fallo_ms"`
}

// ErrorDTO: respuesta cuando ocurre un error.
type ErrorDTO struct {
	Error string `json:"error"`
}
