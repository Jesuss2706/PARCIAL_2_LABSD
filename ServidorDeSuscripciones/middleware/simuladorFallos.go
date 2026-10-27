package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"servidor.local/suscripciones/configuracion"
)

// SimuladorFallos hace que las primeras CANTIDAD_FALLOS peticiones de pago de
// cada token de idempotencia "se pierdan":
//
//   - PETICION:  el mensaje de petición nunca llega a la lógica de negocio (el pago NO se procesa).
//   - RESPUESTA: el pago SÍ se procesa, pero el mensaje de respuesta se pierde.
//
// En ambos casos el servidor retiene la conexión RETARDO_FALLO_MS milisegundos
// (más que el temporizador del cliente) y luego la cierra sin responder.
type SimuladorFallos struct {
	modo     string
	cantidad int
	retardo  time.Duration

	mutex    sync.Mutex
	intentos map[string]int
}

// NuevoSimuladorFallos crea el simulador con la configuración del servidor.
func NuevoSimuladorFallos(cfg configuracion.Configuracion) *SimuladorFallos {
	return &SimuladorFallos{
		modo:     cfg.ModoFallo,
		cantidad: cfg.CantidadFallos,
		retardo:  cfg.RetardoFallo,
		intentos: map[string]int{},
	}
}

// Envolver aplica la simulación de fallos al manejador de pagos.
func (s *SimuladorFallos) Envolver(siguiente http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.modo == configuracion.ModoNinguno || s.cantidad == 0 {
			siguiente(w, r)
			return
		}

		token := leerToken(r)
		intento := s.registrarIntento(token)
		if intento > s.cantidad {
			fmt.Printf("[SIMULADOR] Token %s, petición #%d: sin fallo (ya se simularon %d)\n", token, intento, s.cantidad)
			siguiente(w, r)
			return
		}

		switch s.modo {
		case configuracion.ModoPeticion:
			fmt.Printf("[SIMULADOR] Token %s, petición #%d: MENSAJE DE PETICIÓN PERDIDO (no se procesa el pago)\n", token, intento)
			time.Sleep(s.retardo)
			cortarConexion(w)

		case configuracion.ModoRespuesta:
			siguiente(&escritorDescartable{cabeceras: http.Header{}}, r) // el pago se procesa...
			fmt.Printf("[SIMULADOR] Token %s, petición #%d: pago procesado, MENSAJE DE RESPUESTA PERDIDO\n", token, intento)
			time.Sleep(s.retardo)
			cortarConexion(w) // ...pero la respuesta nunca llega al cliente
		}
	}
}

func (s *SimuladorFallos) registrarIntento(token string) int {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.intentos[token]++
	return s.intentos[token]
}

// leerToken extrae el token del cuerpo y lo restaura para que el controlador pueda leerlo.
func leerToken(r *http.Request) string {
	cuerpo, _ := io.ReadAll(r.Body)
	r.Body = io.NopCloser(bytes.NewReader(cuerpo))
	var solicitud struct {
		Token string `json:"token_idempotencia"`
	}
	_ = json.Unmarshal(cuerpo, &solicitud)
	return solicitud.Token
}

// cortarConexion cierra la conexión TCP sin enviar ninguna respuesta HTTP.
func cortarConexion(w http.ResponseWriter) {
	secuestrador, ok := w.(http.Hijacker)
	if !ok {
		fmt.Println("[SIMULADOR] No se pudo cerrar la conexión: el escritor no soporta Hijack")
		return
	}
	conexion, _, err := secuestrador.Hijack()
	if err != nil {
		fmt.Printf("[SIMULADOR] No se pudo cerrar la conexión: %v\n", err)
		return
	}
	_ = conexion.Close()
}

// escritorDescartable descarta todo lo que el controlador intente escribir.
type escritorDescartable struct {
	cabeceras http.Header
}

func (e *escritorDescartable) Header() http.Header         { return e.cabeceras }
func (e *escritorDescartable) WriteHeader(int)             {}
func (e *escritorDescartable) Write(b []byte) (int, error) { return len(b), nil }
