// Package capaControladores atiende los servicios REST, imprime los ecos y
// delega en la capa de fachada.
package capaControladores

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"servidor.local/suscripciones/capaFachada"
	"servidor.local/suscripciones/dtos"
)

// responderJSON serializa el objeto como JSON con el código indicado.
func responderJSON(w http.ResponseWriter, codigo int, datos any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(codigo)
	if err := json.NewEncoder(w).Encode(datos); err != nil {
		fmt.Printf("[ECO REST] Error serializando la respuesta: %v\n", err)
	}
}

// responderError traduce los errores de la fachada a respuestas HTTP.
func responderError(w http.ResponseWriter, err error) {
	fmt.Printf("[ECO REST] Error: %v\n", err)
	var errorServicio *capaFachada.ErrorServicio
	if errors.As(err, &errorServicio) {
		responderJSON(w, errorServicio.Codigo, dtos.ErrorDTO{Error: errorServicio.Mensaje})
		return
	}
	responderJSON(w, http.StatusInternalServerError, dtos.ErrorDTO{Error: err.Error()})
}

// leerJSON decodifica el cuerpo de la petición; si es inválido responde 400.
func leerJSON(w http.ResponseWriter, r *http.Request, destino any) bool {
	if err := json.NewDecoder(r.Body).Decode(destino); err != nil {
		fmt.Printf("[ECO REST] Cuerpo inválido: %v\n", err)
		responderJSON(w, http.StatusBadRequest, dtos.ErrorDTO{Error: "el cuerpo de la petición no es un JSON válido"})
		return false
	}
	return true
}
