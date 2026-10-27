package capaAccesoDatos

import (
	"sync"

	"servidor.local/suscripciones/modelos"
)

// RepositorioTransacciones es el historial de pagos realizados (archivo JSON).
type RepositorioTransacciones struct {
	mutex         sync.RWMutex
	ruta          string
	transacciones []modelos.Transaccion
}

// NuevoRepositorioTransacciones carga el historial desde la ruta indicada.
func NuevoRepositorioTransacciones(ruta string) (*RepositorioTransacciones, error) {
	repositorio := &RepositorioTransacciones{ruta: ruta, transacciones: []modelos.Transaccion{}}
	if err := cargarJSON(ruta, &repositorio.transacciones); err != nil {
		return nil, err
	}
	return repositorio, nil
}

// Ruta retorna el archivo donde se almacenan las transacciones (para los ecos).
func (r *RepositorioTransacciones) Ruta() string {
	return r.ruta
}

// Guardar agrega la transacción al historial y persiste el archivo.
func (r *RepositorioTransacciones) Guardar(transaccion modelos.Transaccion) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.transacciones = append(r.transacciones, transaccion)
	if err := guardarJSON(r.ruta, r.transacciones); err != nil {
		r.transacciones = r.transacciones[:len(r.transacciones)-1]
		return err
	}
	return nil
}

// BuscarPorToken indica si ya existe una transacción realizada con el token.
func (r *RepositorioTransacciones) BuscarPorToken(token string) (modelos.Transaccion, bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	for _, transaccion := range r.transacciones {
		if transaccion.Token == token {
			return transaccion, true
		}
	}
	return modelos.Transaccion{}, false
}

// ListarPorCedula retorna las transacciones de una cédula en orden de llegada.
func (r *RepositorioTransacciones) ListarPorCedula(cedula string) []modelos.Transaccion {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	resultado := []modelos.Transaccion{}
	for _, transaccion := range r.transacciones {
		if transaccion.Cedula == cedula {
			resultado = append(resultado, transaccion)
		}
	}
	return resultado
}
