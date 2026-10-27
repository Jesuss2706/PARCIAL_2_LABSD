// Package modelos contiene las entidades que se almacenan en la capa de acceso a datos.
package modelos

import "time"

// PlanSuscripcion representa un plan de la plataforma (Individual, Duo, Familiar...).
type PlanSuscripcion struct {
	Id          string  `json:"id"`
	Nombre      string  `json:"nombre"`
	Descripcion string  `json:"descripcion"`
	Precio      float64 `json:"precio"`
	Moneda      string  `json:"moneda"`
	Cuentas     int     `json:"cuentas"`
}

// TokenIdempotencia se genera antes de pagar y asocia al cliente con los datos del pago.
type TokenIdempotencia struct {
	Token                 string    `json:"token"`
	Cedula                string    `json:"cedula"`
	IdPlan                string    `json:"id_plan"`
	Monto                 float64   `json:"monto"`
	NombreTitular         string    `json:"nombre_titular"`
	UltimosDigitosTarjeta string    `json:"ultimos_digitos_tarjeta"`
	FechaCreacion         time.Time `json:"fecha_creacion"`
	Utilizado             bool      `json:"utilizado"`
}

// Transaccion es un pago realizado y almacenado exitosamente.
type Transaccion struct {
	Id     string    `json:"id"`
	Token  string    `json:"token"`
	Cedula string    `json:"cedula"`
	IdPlan string    `json:"id_plan"`
	Monto  float64   `json:"monto"`
	Fecha  time.Time `json:"fecha"`
}
