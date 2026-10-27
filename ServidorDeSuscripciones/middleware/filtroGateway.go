// Package middleware contiene filtros que se ejecutan antes de los controladores.
package middleware

import (
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"net/http"
)

// CabeceraGateway es la cabecera que solo el API Gateway agrega a las peticiones.
const CabeceraGateway = "X-Gateway-Token"

// FiltroGateway rechaza toda petición que no haya pasado por el API Gateway.
func FiltroGateway(secreto string, siguiente http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get(CabeceraGateway)
		if token == "" || subtle.ConstantTimeCompare([]byte(token), []byte(secreto)) != 1 {
			fmt.Printf("[ECO REST] Petición %s %s RECHAZADA: no pasó por el API Gateway\n", r.Method, r.URL.Path)
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusForbidden)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error": "acceso denegado: la petición debe pasar por el API Gateway",
			})
			return
		}
		siguiente.ServeHTTP(w, r)
	})
}
