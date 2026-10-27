// Punto de entrada del Servidor de Suscripciones (REST).
// Arma las capas, se registra en Consul y solo atiende peticiones del API Gateway.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"servidor.local/suscripciones/capaAccesoDatos"
	"servidor.local/suscripciones/capaControladores"
	"servidor.local/suscripciones/capaFachada"
	"servidor.local/suscripciones/configuracion"
	"servidor.local/suscripciones/descubrimiento"
	"servidor.local/suscripciones/middleware"
)

const nombreServicio = "servidor-suscripciones"

func main() {
	cfg := configuracion.Cargar()

	servidor := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Puerto),
		Handler: configurarServidor(cfg),
	}
	mostrarBienvenida(cfg)

	go func() {
		if err := servidor.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Error en el servidor: %v", err)
		}
	}()

	// Al lanzarse, el microservicio se registra en el directorio de descubrimiento.
	registro := descubrimiento.Registrar(cfg.DireccionConsul, nombreServicio, cfg.HostServicio, cfg.Puerto, []string{"rest", "suscripciones"})

	esperarApagado()
	registro.Desregistrar()
	contexto, cancelar := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelar()
	_ = servidor.Shutdown(contexto)
}

// configurarServidor crea las capas: acceso a datos -> fachada -> controladores.
func configurarServidor(cfg configuracion.Configuracion) http.Handler {
	repositorioPlanes := capaAccesoDatos.NuevoRepositorioPlanes()
	repositorioTokens, err := capaAccesoDatos.NuevoRepositorioTokens(filepath.Join(cfg.DirectorioDatos, "tokens.json"))
	if err != nil {
		log.Fatalf("No se pudo cargar el repositorio de tokens: %v", err)
	}
	repositorioTransacciones, err := capaAccesoDatos.NuevoRepositorioTransacciones(filepath.Join(cfg.DirectorioDatos, "transacciones.json"))
	if err != nil {
		log.Fatalf("No se pudo cargar el repositorio de transacciones: %v", err)
	}

	fachadaPlanes := capaFachada.NuevaFachadaPlanes(repositorioPlanes)
	fachadaPagos := capaFachada.NuevaFachadaPagos(repositorioTokens, repositorioTransacciones, repositorioPlanes)

	simulador := middleware.NuevoSimuladorFallos(cfg)
	enrutador := http.NewServeMux()
	capaControladores.NuevoControladorPlanes(fachadaPlanes, cfg).RegistrarRutas(enrutador)
	capaControladores.NuevoControladorPagos(fachadaPagos, simulador.Envolver).RegistrarRutas(enrutador)

	return middleware.FiltroGateway(cfg.SecretoGateway, enrutador)
}

// esperarApagado bloquea hasta recibir Ctrl+C.
func esperarApagado() {
	senal := make(chan os.Signal, 1)
	signal.Notify(senal, os.Interrupt, syscall.SIGTERM)
	<-senal
	fmt.Println("\nApagando el servidor de suscripciones...")
}

func mostrarBienvenida(cfg configuracion.Configuracion) {
	fmt.Println("==================================================")
	fmt.Println("   Servidor de Suscripciones (REST)")
	fmt.Printf("   Puerto:            %d\n", cfg.Puerto)
	fmt.Printf("   Consul:            %s\n", cfg.DireccionConsul)
	fmt.Printf("   MODO_FALLO:        %s\n", cfg.ModoFallo)
	fmt.Printf("   CANTIDAD_FALLOS:   %d\n", cfg.CantidadFallos)
	fmt.Printf("   RETARDO_FALLO_MS:  %d\n", cfg.RetardoFallo.Milliseconds())
	fmt.Printf("   DIRECTORIO_DATOS:  %s\n", cfg.DirectorioDatos)
	fmt.Println("==================================================")
	fmt.Println(" GET  /suscripciones/planes")
	fmt.Println(" GET  /suscripciones/estado")
	fmt.Println(" POST /suscripciones/tokens")
	fmt.Println(" POST /suscripciones/pagos")
	fmt.Println(" GET  /suscripciones/transacciones/{cedula}")
	fmt.Println("==================================================")
}
