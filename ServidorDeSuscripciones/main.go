// Servidor de suscripciones: planes, tokens de idempotencia y pagos (basado en la guía 10).
// Se registra en el Directorio Consul y solo acepta peticiones que pasen por el API Gateway.
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"servidor-suscripciones/config"
	"servidor-suscripciones/controllers"
	"servidor-suscripciones/data/repositories"
	"servidor-suscripciones/discovery"
	"servidor-suscripciones/middleware"
	"servidor-suscripciones/services"
)

const nombreServicio = "servidor-suscripciones"

func main() {
	cfg := config.Cargar()
	host := env("SERVICE_HOST", "127.0.0.1")
	consulAddr := env("CONSUL_ADDR", "127.0.0.1:8500")
	secreto := env("GATEWAY_SECRET", "secreto-gateway-2026")

	// Capa de acceso a datos (archivos JSON en ./datos)
	suscripcionRepo := repositories.NuevoSuscripcionRepository()
	tokenRepo, err := repositories.NuevoTokenRepository(filepath.Join(cfg.DirectorioDatos, "tokens.json"))
	if err != nil {
		log.Fatalf("No se pudo cargar el repositorio de tokens: %v", err)
	}
	transaccionRepo, err := repositories.NuevoTransaccionRepository(filepath.Join(cfg.DirectorioDatos, "transacciones.json"))
	if err != nil {
		log.Fatalf("No se pudo cargar el repositorio de transacciones: %v", err)
	}

	// Capa de fachada
	suscripcionService := services.NuevoSuscripcionService(suscripcionRepo)
	pagoService := services.NuevoPagoService(tokenRepo, transaccionRepo, suscripcionRepo)

	// Capa de controladores
	simulador := middleware.NuevoSimuladorFallos(cfg)
	router := gin.Default()
	router.Use(middleware.FiltroGateway(secreto))
	api := router.Group("/suscripciones")
	controllers.NuevoSuscripcionController(suscripcionService).Registrar(api)
	controllers.NuevoPagoController(pagoService, simulador.Middleware()).Registrar(api)
	controllers.NuevoEstadoController(cfg).Registrar(api)

	srv := &http.Server{Addr: ":" + cfg.Puerto, Handler: router}
	go func() {
		log.Println("==============================================")
		log.Println(" Servidor de suscripciones (pasarela de pagos)")
		log.Printf(" Puerto:            %s", cfg.Puerto)
		log.Printf(" MODO_FALLO:        %s", cfg.ModoFallo)
		log.Printf(" CANTIDAD_FALLOS:   %d", cfg.CantidadFallos)
		log.Printf(" RETARDO_FALLO_MS:  %d", cfg.RetardoFallo.Milliseconds())
		log.Printf(" DIRECTORIO_DATOS:  %s", cfg.DirectorioDatos)
		log.Println("==============================================")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	puerto, _ := strconv.Atoi(cfg.Puerto)
	registro, err := discovery.Registrar(consulAddr, nombreServicio, host, puerto)
	if err != nil {
		log.Fatalf("no fue posible registrarse en Consul: %v", err)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	registro.Desregistrar()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}

func env(clave, porDefecto string) string {
	if v := os.Getenv(clave); v != "" {
		return v
	}
	return porDefecto
}
