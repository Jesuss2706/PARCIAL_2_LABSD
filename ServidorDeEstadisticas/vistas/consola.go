// Package vistas imprime por pantalla las estadísticas de reproducción.
package vistas

import (
	"fmt"

	"estadisticas/dtos"
)

// MostrarReproduccion imprime los datos de la reproducción recibida.
func MostrarReproduccion(reproduccion dtos.NotificacionReproduccionDTO) {
	fmt.Println("--------------------------------------------------")
	fmt.Println(" Nueva reproducción")
	fmt.Printf("   Audio:      %s (id %d)\n", reproduccion.Titulo, reproduccion.IdAudio)
	fmt.Printf("   Tipo:       %s\n", reproduccion.Tipo)
	fmt.Printf("   Archivo:    %s\n", reproduccion.Archivo)
	fmt.Printf("   Cliente:    %s\n", reproduccion.Cliente)
	fmt.Printf("   Fecha/hora: %s\n", reproduccion.FechaHora)
}

// MostrarResumen imprime los acumulados de reproducciones.
func MostrarResumen(resumen dtos.ResumenEstadisticasDTO) {
	fmt.Printf(" Total de reproducciones: %d\n", resumen.TotalReproducciones)
	fmt.Println(" Por tipo de audio:")
	for _, conteo := range resumen.PorTipo {
		fmt.Printf("   - %-25s %d\n", conteo.Nombre, conteo.Cantidad)
	}
	fmt.Println(" Por audio:")
	for _, conteo := range resumen.PorAudio {
		fmt.Printf("   - %-45s %d\n", conteo.Nombre, conteo.Cantidad)
	}
	fmt.Println("--------------------------------------------------")
}
