package capaFachada

import (
	"time"

	"servidor.local/suscripciones/dtos"
	"servidor.local/suscripciones/modelos"
)

// Funciones de mapeo entre modelos (capa de datos) y DTOs (capa de controladores).

const formatoFecha = time.RFC3339

func aPlanDTO(plan modelos.PlanSuscripcion) dtos.PlanDTO {
	return dtos.PlanDTO{
		Id: plan.Id, Nombre: plan.Nombre, Descripcion: plan.Descripcion,
		Precio: plan.Precio, Moneda: plan.Moneda, Cuentas: plan.Cuentas,
	}
}

func aRespuestaTokenDTO(token modelos.TokenIdempotencia) dtos.RespuestaTokenDTO {
	return dtos.RespuestaTokenDTO{
		TokenIdempotencia: token.Token,
		Cedula:            token.Cedula,
		IdPlan:            token.IdPlan,
		Monto:             token.Monto,
		FechaCreacion:     token.FechaCreacion.Format(formatoFecha),
	}
}

func aTokenDTO(token modelos.TokenIdempotencia) dtos.TokenDTO {
	return dtos.TokenDTO{
		TokenIdempotencia: token.Token,
		IdPlan:            token.IdPlan,
		Monto:             token.Monto,
		FechaCreacion:     token.FechaCreacion.Format(formatoFecha),
		Utilizado:         token.Utilizado,
	}
}

func aTransaccionDTO(transaccion modelos.Transaccion) dtos.TransaccionDTO {
	return dtos.TransaccionDTO{
		IdTransaccion:     transaccion.Id,
		TokenIdempotencia: transaccion.Token,
		IdPlan:            transaccion.IdPlan,
		Monto:             transaccion.Monto,
		Fecha:             transaccion.Fecha.Format(formatoFecha),
	}
}
