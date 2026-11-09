// Servidor de Estadísticas:
//   - consume asíncronamente la cola de reproducciones que publica el servidor de streaming;
//   - expone por REST (vía API Gateway) los audios reproducidos por cada cliente.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	fachada "estadisticas/capaFachada"
	capacontroladores "estadisticas/capaControladores"
	componenteconexioncola "estadisticas/componenteConexionCola"
	"estadisticas/discovery"
	"estadisticas/middleware"
)

const nombreServicio = "servidor-estadisticas"

func main() {
	puerto, _ := strconv.Atoi(env("PORT", "5003"))
	host := env("SERVICE_HOST", "127.0.0.1")
	consulAddr := env("CONSUL_ADDR", "127.0.0.1:8500")
	secreto := env("GATEWAY_SECRET", "secreto-gateway-2026")

	fachadaEstadisticas := fachada.NuevaFachadaEstadisticas()

	// 1) Consumidor de la cola (comunicación asíncrona). Si RabbitMQ no está arriba
	//    se reintenta cada 5 s sin tumbar el servicio REST.
	go consumirCola(fachadaEstadisticas)

	// 2) Servicio REST registrado en Consul
	router := gin.Default()
	router.Use(middleware.FiltroGateway(secreto))
	capacontroladores.NuevoControladorEstadisticas(fachadaEstadisticas).RegistrarRutas(router)

	srv := &http.Server{Addr: fmt.Sprintf(":%d", puerto), Handler: router}
	go func() {
		fmt.Println("==================================================")
		fmt.Println("   Servidor de Estadísticas")
		fmt.Printf("   REST en :%d  |  cola %s\n", puerto, componenteconexioncola.NombreCola)
		fmt.Println("==================================================")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("error en el servidor REST: %v", err)
		}
	}()

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

func consumirCola(f *fachada.FachadaEstadisticas) {
	for {
		consumidor, err := componenteconexioncola.NewRabbitConsumer()
		if err != nil {
			fmt.Println("RabbitMQ no disponible, reintentando en 5 s:", err)
			time.Sleep(5 * time.Second)
			continue
		}
		fmt.Printf("Escuchando la cola %s ...\n", componenteconexioncola.NombreCola)
		err = consumidor.Consumir(f.ProcesarMensaje)
		consumidor.Cerrar()
		fmt.Println(err, "- reconectando en 5 s")
		time.Sleep(5 * time.Second)
	}
}

func env(clave, porDefecto string) string {
	if v := os.Getenv(clave); v != "" {
		return v
	}
	return porDefecto
}
