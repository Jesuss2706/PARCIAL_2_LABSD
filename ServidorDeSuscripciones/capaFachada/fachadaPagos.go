package capaFachada

import (
	"crypto/rand"
	"fmt"
	"math"
	"regexp"
	"strings"
	"sync"
	"time"

	"servidor.local/suscripciones/capaAccesoDatos"
	"servidor.local/suscripciones/dtos"
	"servidor.local/suscripciones/modelos"
)

var (
	regexCedula  = regexp.MustCompile(`^[0-9]{6,10}$`)
	regexTarjeta = regexp.MustCompile(`^[0-9]{13,19}$`)
)

// FachadaPagos valida los pagos, aplica la regla de idempotencia y mapea a DTOs.
type FachadaPagos struct {
	// mutex hace atómico "¿ya se usó el token? -> registrar transacción",
	// evitando que dos peticiones con el mismo token paguen dos veces.
	mutex                    sync.Mutex
	repositorioTokens        *capaAccesoDatos.RepositorioTokens
	repositorioTransacciones *capaAccesoDatos.RepositorioTransacciones
	repositorioPlanes        *capaAccesoDatos.RepositorioPlanes
}

// NuevaFachadaPagos crea la fachada de pagos.
func NuevaFachadaPagos(
	repositorioTokens *capaAccesoDatos.RepositorioTokens,
	repositorioTransacciones *capaAccesoDatos.RepositorioTransacciones,
	repositorioPlanes *capaAccesoDatos.RepositorioPlanes,
) *FachadaPagos {
	return &FachadaPagos{
		repositorioTokens:        repositorioTokens,
		repositorioTransacciones: repositorioTransacciones,
		repositorioPlanes:        repositorioPlanes,
	}
}

// GenerarToken valida los datos del pago y emite un token de idempotencia.
func (f *FachadaPagos) GenerarToken(solicitud dtos.SolicitudTokenDTO) (dtos.RespuestaTokenDTO, error) {
	solicitud.Cedula = strings.TrimSpace(solicitud.Cedula)
	solicitud.NombreTitular = strings.TrimSpace(solicitud.NombreTitular)
	tarjeta := strings.ReplaceAll(strings.TrimSpace(solicitud.NumeroTarjeta), " ", "")

	if !regexCedula.MatchString(solicitud.Cedula) {
		return dtos.RespuestaTokenDTO{}, errorValidacion("la cédula debe tener entre 6 y 10 dígitos")
	}
	if solicitud.NombreTitular == "" {
		return dtos.RespuestaTokenDTO{}, errorValidacion("el nombre del titular es obligatorio")
	}
	if !regexTarjeta.MatchString(tarjeta) {
		return dtos.RespuestaTokenDTO{}, errorValidacion("el número de tarjeta debe tener entre 13 y 19 dígitos")
	}
	plan, existe := f.repositorioPlanes.BuscarPorId(solicitud.IdPlan)
	if !existe {
		return dtos.RespuestaTokenDTO{}, errorNoEncontrado(fmt.Sprintf("el plan '%s' no existe", solicitud.IdPlan))
	}
	if !mismoMonto(solicitud.Monto, plan.Precio) {
		return dtos.RespuestaTokenDTO{}, errorNegocio(fmt.Sprintf("el monto %.0f no corresponde al precio del plan (%.0f)", solicitud.Monto, plan.Precio))
	}

	token := modelos.TokenIdempotencia{
		Token:                 nuevoUUID(),
		Cedula:                solicitud.Cedula,
		IdPlan:                plan.Id,
		Monto:                 plan.Precio,
		NombreTitular:         solicitud.NombreTitular,
		UltimosDigitosTarjeta: tarjeta[len(tarjeta)-4:],
		FechaCreacion:         time.Now(),
	}
	if err := f.repositorioTokens.Guardar(token); err != nil {
		return dtos.RespuestaTokenDTO{}, errorInterno("no fue posible almacenar el token: " + err.Error())
	}
	fmt.Printf("[ECO TOKEN] Token de idempotencia generado: %s (cédula %s, plan %s, $%.0f)\n",
		token.Token, token.Cedula, token.IdPlan, token.Monto)
	return aRespuestaTokenDTO(token), nil
}

// RealizarPago ejecuta el pago una única vez por token de idempotencia.
func (f *FachadaPagos) RealizarPago(solicitud dtos.SolicitudPagoDTO) (dtos.RespuestaPagoDTO, error) {
	solicitud.TokenIdempotencia = strings.TrimSpace(solicitud.TokenIdempotencia)
	if solicitud.TokenIdempotencia == "" {
		return dtos.RespuestaPagoDTO{}, errorValidacion("el token de idempotencia es obligatorio")
	}
	if solicitud.Monto <= 0 {
		return dtos.RespuestaPagoDTO{}, errorValidacion("el monto debe ser mayor que cero")
	}

	f.mutex.Lock()
	defer f.mutex.Unlock()

	fmt.Printf("[ECO PAGO] Token utilizado en el pago: %s (monto $%.0f)\n", solicitud.TokenIdempotencia, solicitud.Monto)

	token, existe := f.repositorioTokens.BuscarPorToken(solicitud.TokenIdempotencia)
	if !existe {
		fmt.Printf("[ECO PAGO] Pago rechazado: el token %s no existe\n", solicitud.TokenIdempotencia)
		return dtos.RespuestaPagoDTO{}, errorNoEncontrado("el token de idempotencia no existe")
	}

	// Regla de idempotencia: antes de pagar se pregunta si el token ya fue utilizado.
	if previa, yaUtilizado := f.repositorioTransacciones.BuscarPorToken(token.Token); yaUtilizado {
		fmt.Printf("[ECO PAGO] Token %s YA UTILIZADO (transacción %s por $%.0f). NO se cobra de nuevo.\n",
			token.Token, previa.Id, previa.Monto)
		montoPrevio := previa.Monto
		return dtos.RespuestaPagoDTO{
			TokenIdempotencia:      token.Token,
			MontoTransaccion:       solicitud.Monto,
			TransaccionRealizada:   false,
			PagoAlmacenado:         false,
			PreviamenteRealizada:   true,
			MontoTransaccionPrevia: &montoPrevio,
			IdTransaccion:          previa.Id,
			FechaTransaccion:       previa.Fecha.Format(formatoFecha),
			Mensaje:                "El pago no se ejecutó de nuevo: ya existía una transacción con este token de idempotencia",
		}, nil
	}

	if !mismoMonto(solicitud.Monto, token.Monto) {
		return dtos.RespuestaPagoDTO{}, errorNegocio(fmt.Sprintf("el monto %.0f no coincide con el monto del token (%.0f)", solicitud.Monto, token.Monto))
	}

	transaccion := modelos.Transaccion{
		Id:     nuevoUUID(),
		Token:  token.Token,
		Cedula: token.Cedula,
		IdPlan: token.IdPlan,
		Monto:  solicitud.Monto,
		Fecha:  time.Now(),
	}
	fmt.Printf("[ECO PAGO] Pago realizado exitosamente: transacción %s, cédula %s, plan %s, $%.0f\n",
		transaccion.Id, transaccion.Cedula, transaccion.IdPlan, transaccion.Monto)

	if err := f.repositorioTransacciones.Guardar(transaccion); err != nil {
		fmt.Printf("[ECO ALMACENAMIENTO] ERROR: el pago no se pudo almacenar: %v\n", err)
		return dtos.RespuestaPagoDTO{}, errorInterno("no fue posible almacenar la transacción: " + err.Error())
	}
	fmt.Printf("[ECO ALMACENAMIENTO] Pago almacenado en %s\n", f.repositorioTransacciones.Ruta())

	token.Utilizado = true
	if err := f.repositorioTokens.Guardar(token); err != nil {
		return dtos.RespuestaPagoDTO{}, errorInterno("no fue posible actualizar el token: " + err.Error())
	}

	return dtos.RespuestaPagoDTO{
		TokenIdempotencia:    token.Token,
		MontoTransaccion:     transaccion.Monto,
		TransaccionRealizada: true,
		PagoAlmacenado:       true,
		PreviamenteRealizada: false,
		IdTransaccion:        transaccion.Id,
		FechaTransaccion:     transaccion.Fecha.Format(formatoFecha),
		Mensaje:              "Pago realizado y almacenado exitosamente",
	}, nil
}

// ConsultarTransacciones devuelve el plan actual, el historial de pagos y los tokens de una cédula.
func (f *FachadaPagos) ConsultarTransacciones(cedula string) (dtos.RespuestaConsultaDTO, error) {
	cedula = strings.TrimSpace(cedula)
	if !regexCedula.MatchString(cedula) {
		return dtos.RespuestaConsultaDTO{}, errorValidacion("la cédula debe tener entre 6 y 10 dígitos")
	}

	respuesta := dtos.RespuestaConsultaDTO{
		Cedula:            cedula,
		Transacciones:     []dtos.TransaccionDTO{},
		TokensUtilizados:  []dtos.TokenDTO{},
		TokensSinUtilizar: []dtos.TokenDTO{},
	}
	transacciones := f.repositorioTransacciones.ListarPorCedula(cedula)
	for _, transaccion := range transacciones {
		respuesta.Transacciones = append(respuesta.Transacciones, aTransaccionDTO(transaccion))
	}
	// El plan actual es el de la última transacción realizada.
	if len(transacciones) > 0 {
		if plan, existe := f.repositorioPlanes.BuscarPorId(transacciones[len(transacciones)-1].IdPlan); existe {
			planDTO := aPlanDTO(plan)
			respuesta.PlanActual = &planDTO
		}
	}
	for _, token := range f.repositorioTokens.ListarPorCedula(cedula) {
		if token.Utilizado {
			respuesta.TokensUtilizados = append(respuesta.TokensUtilizados, aTokenDTO(token))
		} else {
			respuesta.TokensSinUtilizar = append(respuesta.TokensSinUtilizar, aTokenDTO(token))
		}
	}
	respuesta.TotalTransacciones = len(respuesta.Transacciones)
	return respuesta, nil
}

func mismoMonto(a, b float64) bool {
	return math.Abs(a-b) < 0.005
}

// nuevoUUID genera un identificador UUID v4.
func nuevoUUID() string {
	bytes := make([]byte, 16)
	_, _ = rand.Read(bytes)
	bytes[6] = (bytes[6] & 0x0f) | 0x40
	bytes[8] = (bytes[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", bytes[0:4], bytes[4:6], bytes[6:8], bytes[8:10], bytes[10:16])
}
