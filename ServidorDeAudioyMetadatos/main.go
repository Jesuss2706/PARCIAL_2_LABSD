// Servidor de audios y metadatos: catálogo de tipos, metadatos propios de cada tipo,
// audios, y registro de nuevos audios (metadatos + mp3) desde el administrador web.
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
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"servidor-audios/capaAccesoADatos/repository"
	"servidor-audios/capaControladores"
	"servidor-audios/capaFachadaServices"
	"servidor-audios/discovery"
	"servidor-audios/middleware"
)

const nombreServicio = "servidor-audios"

func main() {
	puerto, _ := strconv.Atoi(env("PORT", "5001"))
	host := env("SERVICE_HOST", "127.0.0.1")
	consulAddr := env("CONSUL_ADDR", "127.0.0.1:8500")
	secreto := env("GATEWAY_SECRET", "secreto-gateway-2026")
	directorioDatos := env("DIRECTORIO_DATOS", "./datos")
	// Carpeta compartida con el servidor de streaming: allí se guardan los mp3.
	carpetaAudios := env("RUTA_AUDIOS", filepath.Join("..", "ServidorDeStreaming", "audios"))

	// Capas: acceso a datos -> fachada (services) -> controladores
	repo, err := repository.NuevoAudioRepository(filepath.Join(directorioDatos, "audios.json"))
	if err != nil {
		log.Fatalf("no se pudo cargar el repositorio de audios: %v", err)
	}
	servicio := capaFachadaServices.NuevoAudioService(repo, carpetaAudios)
	controlador := capaControladores.NuevoAudioController(servicio)

	router := gin.Default()
	router.MaxMultipartMemory = 32 << 20
	router.Use(middleware.FiltroGateway(secreto))
	controlador.RegistrarRutas(router)

	srv := &http.Server{Addr: fmt.Sprintf(":%d", puerto), Handler: router}
	go func() {
		log.Println("==================================================")
		log.Printf("   Servidor de audios y metadatos en :%d", puerto)
		log.Printf("   Carpeta de mp3 (streaming): %s", carpetaAudios)
		log.Println("==================================================")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("error en el servidor: %v", err)
		}
	}()

	// Al lanzarse, el microservicio se registra en el Directorio Consul.
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
