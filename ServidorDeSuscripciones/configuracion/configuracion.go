// Package configuracion lee la configuración del servidor desde variables de entorno.
package configuracion

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

// Modos de simulación de fallos (laboratorio de tokens idempotentes).
const (
	ModoNinguno   = "NINGUNO"   // el servidor funciona normalmente
	ModoPeticion  = "PETICION"  // se pierde el mensaje de petición (el pago NO se procesa)
	ModoRespuesta = "RESPUESTA" // se pierde el mensaje de respuesta (el pago SÍ se procesa)
)

// Configuracion agrupa todos los parámetros del servidor.
type Configuracion struct {
	Puerto          int
	HostServicio    string
	DireccionConsul string
	SecretoGateway  string
	ModoFallo       string
	CantidadFallos  int
	RetardoFallo    time.Duration
	DirectorioDatos string
}

// Cargar lee las variables de entorno y aplica valores por defecto.
//
//	PUERTO            puerto HTTP (8084)
//	SERVICE_HOST      dirección con la que se registra en Consul (127.0.0.1)
//	CONSUL_ADDR       dirección del agente Consul (127.0.0.1:8500)
//	GATEWAY_SECRET    secreto que agrega el API Gateway (secreto-gateway-2026)
//	MODO_FALLO        NINGUNO | PETICION | RESPUESTA (NINGUNO)
//	CANTIDAD_FALLOS   peticiones de pago por token que fallarán (0)
//	RETARDO_FALLO_MS  tiempo que se retiene la conexión antes de cortarla (5000)
//	DIRECTORIO_DATOS  carpeta donde se guardan tokens y transacciones (./datos)
func Cargar() Configuracion {
	cfg := Configuracion{
		Puerto:          entero("PUERTO", 8084),
		HostServicio:    texto("SERVICE_HOST", "127.0.0.1"),
		DireccionConsul: texto("CONSUL_ADDR", "127.0.0.1:8500"),
		SecretoGateway:  texto("GATEWAY_SECRET", "secreto-gateway-2026"),
		ModoFallo:       strings.ToUpper(texto("MODO_FALLO", ModoNinguno)),
		CantidadFallos:  entero("CANTIDAD_FALLOS", 0),
		RetardoFallo:    time.Duration(entero("RETARDO_FALLO_MS", 5000)) * time.Millisecond,
		DirectorioDatos: texto("DIRECTORIO_DATOS", "./datos"),
	}

	switch cfg.ModoFallo {
	case ModoNinguno, ModoPeticion, ModoRespuesta:
	default:
		log.Printf("MODO_FALLO '%s' no válido, se usará %s", cfg.ModoFallo, ModoNinguno)
		cfg.ModoFallo = ModoNinguno
	}
	if cfg.ModoFallo == ModoNinguno {
		cfg.CantidadFallos = 0
	}
	return cfg
}

func texto(clave, porDefecto string) string {
	if valor := strings.TrimSpace(os.Getenv(clave)); valor != "" {
		return valor
	}
	return porDefecto
}

func entero(clave string, porDefecto int) int {
	valor := strings.TrimSpace(os.Getenv(clave))
	if valor == "" {
		return porDefecto
	}
	numero, err := strconv.Atoi(valor)
	if err != nil || numero < 0 {
		log.Printf("Valor inválido para %s='%s', se usará %d", clave, valor, porDefecto)
		return porDefecto
	}
	return numero
}
