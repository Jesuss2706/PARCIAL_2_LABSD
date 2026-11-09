package capaFachadaServices

import "net/http"

// ErrorServicio es un error de negocio con el código HTTP que debe devolver el controlador.
type ErrorServicio struct {
	Codigo  int
	Mensaje string
}

func (e *ErrorServicio) Error() string { return e.Mensaje }

func errorValidacion(m string) *ErrorServicio  { return &ErrorServicio{http.StatusBadRequest, m} }
func errorNoEncontrado(m string) *ErrorServicio { return &ErrorServicio{http.StatusNotFound, m} }
func errorConflicto(m string) *ErrorServicio    { return &ErrorServicio{http.StatusConflict, m} }
func errorInterno(m string) *ErrorServicio      { return &ErrorServicio{http.StatusInternalServerError, m} }
