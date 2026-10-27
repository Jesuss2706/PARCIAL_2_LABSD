package capaAccesoDatos

import "servidor.local/suscripciones/modelos"

// RepositorioPlanes mantiene en memoria el catálogo de planes (datos quemados).
type RepositorioPlanes struct {
	planes []modelos.PlanSuscripcion
}

// NuevoRepositorioPlanes crea el repositorio con los planes disponibles.
func NuevoRepositorioPlanes() *RepositorioPlanes {
	return &RepositorioPlanes{
		planes: []modelos.PlanSuscripcion{
			{Id: "individual", Nombre: "Premium Individual", Descripcion: "Audios sin anuncios, descargas y saltos ilimitados.", Precio: 16900, Moneda: "COP", Cuentas: 1},
			{Id: "duo", Nombre: "Premium Duo", Descripcion: "2 cuentas Premium para personas que viven juntas.", Precio: 21900, Moneda: "COP", Cuentas: 2},
			{Id: "familiar", Nombre: "Premium Familiar", Descripcion: "Hasta 6 cuentas Premium y control parental.", Precio: 26900, Moneda: "COP", Cuentas: 6},
			{Id: "estudiantes", Nombre: "Premium Estudiantes", Descripcion: "Precio especial para estudiantes universitarios.", Precio: 8900, Moneda: "COP", Cuentas: 1},
		},
	}
}

// ListarTodos retorna una copia del catálogo.
func (r *RepositorioPlanes) ListarTodos() []modelos.PlanSuscripcion {
	copia := make([]modelos.PlanSuscripcion, len(r.planes))
	copy(copia, r.planes)
	return copia
}

// BuscarPorId retorna el plan y true si existe.
func (r *RepositorioPlanes) BuscarPorId(id string) (modelos.PlanSuscripcion, bool) {
	for _, plan := range r.planes {
		if plan.Id == id {
			return plan, true
		}
	}
	return modelos.PlanSuscripcion{}, false
}
