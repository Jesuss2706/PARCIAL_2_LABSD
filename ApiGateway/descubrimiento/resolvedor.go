// Package descubrimiento consulta el directorio Consul (API HTTP) para obtener
// la dirección de una instancia sana de un servicio.
package descubrimiento

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync/atomic"
	"time"
)

// Resolvedor obtiene direcciones de servicios desde Consul (balanceo round-robin).
type Resolvedor struct {
	direccionConsul string
	cliente         *http.Client
	turno           atomic.Uint64
}

// NuevoResolvedor crea el resolvedor apuntando al agente Consul.
func NuevoResolvedor(direccionConsul string) *Resolvedor {
	return &Resolvedor{direccionConsul: direccionConsul, cliente: &http.Client{Timeout: 3 * time.Second}}
}

type instanciaConsul struct {
	Node    struct{ Address string }
	Service struct {
		Address string
		Port    int
	}
}

// Resolver devuelve la URL base de una instancia sana del servicio.
func (r *Resolvedor) Resolver(servicio string) (*url.URL, error) {
	respuesta, err := r.cliente.Get(fmt.Sprintf("http://%s/v1/health/service/%s?passing=true",
		r.direccionConsul, url.PathEscape(servicio)))
	if err != nil {
		return nil, fmt.Errorf("error consultando Consul: %w", err)
	}
	defer respuesta.Body.Close()
	if respuesta.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Consul respondió HTTP %d", respuesta.StatusCode)
	}

	var instancias []instanciaConsul
	if err := json.NewDecoder(respuesta.Body).Decode(&instancias); err != nil {
		return nil, fmt.Errorf("respuesta inválida de Consul: %w", err)
	}
	if len(instancias) == 0 {
		return nil, fmt.Errorf("no hay instancias disponibles de %q en Consul", servicio)
	}

	instancia := instancias[int(r.turno.Add(1)%uint64(len(instancias)))]
	host := instancia.Service.Address
	if host == "" {
		host = instancia.Node.Address
	}
	return url.Parse(fmt.Sprintf("http://%s:%d", host, instancia.Service.Port))
}
