package capaFachada

import (
	"servidor.local/suscripciones/capaAccesoDatos"
	"servidor.local/suscripciones/dtos"
)

// FachadaPlanes expone el catálogo de planes de suscripción.
type FachadaPlanes struct {
	repositorio *capaAccesoDatos.RepositorioPlanes
}

// NuevaFachadaPlanes crea la fachada.
func NuevaFachadaPlanes(repositorio *capaAccesoDatos.RepositorioPlanes) *FachadaPlanes {
	return &FachadaPlanes{repositorio: repositorio}
}

// ListarPlanes retorna todos los planes como DTOs.
func (f *FachadaPlanes) ListarPlanes() []dtos.PlanDTO {
	resultado := []dtos.PlanDTO{}
	for _, plan := range f.repositorio.ListarTodos() {
		resultado = append(resultado, aPlanDTO(plan))
	}
	return resultado
}
