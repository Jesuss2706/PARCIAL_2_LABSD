package capaAccesoDatos

import (
	"sort"
	"sync"

	"servidor.local/suscripciones/modelos"
)

// RepositorioTokens almacena los tokens de idempotencia en un archivo JSON.
type RepositorioTokens struct {
	mutex  sync.RWMutex
	ruta   string
	tokens map[string]modelos.TokenIdempotencia
}

// NuevoRepositorioTokens carga los tokens existentes desde la ruta indicada.
func NuevoRepositorioTokens(ruta string) (*RepositorioTokens, error) {
	repositorio := &RepositorioTokens{ruta: ruta, tokens: map[string]modelos.TokenIdempotencia{}}
	if err := cargarJSON(ruta, &repositorio.tokens); err != nil {
		return nil, err
	}
	return repositorio, nil
}

// Guardar inserta o actualiza un token y persiste el archivo.
func (r *RepositorioTokens) Guardar(token modelos.TokenIdempotencia) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.tokens[token.Token] = token
	return guardarJSON(r.ruta, r.tokens)
}

// BuscarPorToken retorna el token y true si existe.
func (r *RepositorioTokens) BuscarPorToken(token string) (modelos.TokenIdempotencia, bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	encontrado, existe := r.tokens[token]
	return encontrado, existe
}

// ListarPorCedula retorna los tokens de una cédula ordenados por fecha.
func (r *RepositorioTokens) ListarPorCedula(cedula string) []modelos.TokenIdempotencia {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	resultado := []modelos.TokenIdempotencia{}
	for _, token := range r.tokens {
		if token.Cedula == cedula {
			resultado = append(resultado, token)
		}
	}
	sort.Slice(resultado, func(i, j int) bool {
		return resultado[i].FechaCreacion.Before(resultado[j].FechaCreacion)
	})
	return resultado
}
