// Package descubrimiento registra el microservicio en el directorio Consul
// usando su API HTTP (sin dependencias externas).
package descubrimiento

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Registro mantiene el registro del microservicio en Consul con un health check TTL
// que el propio servicio renueva cada 5 segundos.
type Registro struct {
	direccionConsul string
	id              string
	idCheck         string
	cuerpoRegistro  []byte
	cliente         *http.Client
	parar           chan struct{}
}

// Registrar inscribe el servicio en Consul. Si Consul no está disponible, sigue
// intentándolo en segundo plano (el servicio funciona igual).
func Registrar(direccionConsul, nombre, host string, puerto int, etiquetas []string) *Registro {
	id := fmt.Sprintf("%s-%s-%d", nombre, host, puerto)
	registro := &Registro{
		direccionConsul: direccionConsul,
		id:              id,
		idCheck:         "check-" + id,
		cliente:         &http.Client{Timeout: 3 * time.Second},
		parar:           make(chan struct{}),
	}
	registro.cuerpoRegistro, _ = json.Marshal(map[string]any{
		"ID":      id,
		"Name":    nombre,
		"Address": host,
		"Port":    puerto,
		"Tags":    etiquetas,
		"Check": map[string]any{
			"CheckID":                        registro.idCheck,
			"TTL":                            "15s",
			"DeregisterCriticalServiceAfter": "1m",
		},
	})
	go registro.mantenerRegistro(nombre)
	return registro
}

// mantenerRegistro registra el servicio y renueva el TTL; si Consul se reinicia, se vuelve a registrar.
func (r *Registro) mantenerRegistro(nombre string) {
	registrado := false
	temporizador := time.NewTicker(5 * time.Second)
	defer temporizador.Stop()
	for {
		if !registrado {
			if err := r.enviar(http.MethodPut, "/v1/agent/service/register", r.cuerpoRegistro); err != nil {
				fmt.Printf("[CONSUL] No fue posible registrarse en %s: %v (se reintenta en 5 s)\n", r.direccionConsul, err)
			} else {
				registrado = true
				fmt.Printf("[CONSUL] Servicio %q registrado en Consul (%s) como %s\n", nombre, r.direccionConsul, r.id)
			}
		}
		if registrado {
			if err := r.enviar(http.MethodPut, "/v1/agent/check/pass/"+r.idCheck, nil); err != nil {
				fmt.Printf("[CONSUL] Error renovando el TTL: %v\n", err)
				registrado = false
			}
		}
		select {
		case <-temporizador.C:
		case <-r.parar:
			return
		}
	}
}

// Desregistrar elimina el servicio de Consul (al apagarse).
func (r *Registro) Desregistrar() {
	close(r.parar)
	if err := r.enviar(http.MethodPut, "/v1/agent/service/deregister/"+r.id, nil); err != nil {
		fmt.Printf("[CONSUL] Error desregistrando: %v\n", err)
		return
	}
	fmt.Printf("[CONSUL] Servicio %s desregistrado de Consul\n", r.id)
}

func (r *Registro) enviar(metodo, ruta string, cuerpo []byte) error {
	peticion, err := http.NewRequest(metodo, "http://"+r.direccionConsul+ruta, bytes.NewReader(cuerpo))
	if err != nil {
		return err
	}
	respuesta, err := r.cliente.Do(peticion)
	if err != nil {
		return err
	}
	defer respuesta.Body.Close()
	if respuesta.StatusCode != http.StatusOK {
		return fmt.Errorf("Consul respondió HTTP %d", respuesta.StatusCode)
	}
	return nil
}
