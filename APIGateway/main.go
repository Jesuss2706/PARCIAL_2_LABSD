// API Gateway: único punto de entrada del cliente web y del administrador web.
// Resuelve los microservicios en el Directorio Consul y reenvía las llamadas
// gRPC-Web al Proxy Envoy (que a su vez las envía al servidor de streaming).
package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"api-gateway/config"
	"api-gateway/discovery"
	"api-gateway/middleware"
	"api-gateway/proxy"
)

func main() {
	puerto := env("PORT", "8000")
	consulAddr := env("CONSUL_ADDR", "127.0.0.1:8500")
	secreto := env("GATEWAY_SECRET", "secreto-gateway-2026")
	urlEnvoy := env("ENVOY_URL", "http://127.0.0.1:8080")

	resolver, err := discovery.NewResolver(consulAddr)
	if err != nil {
		log.Fatalf("no fue posible crear el cliente de Consul: %v", err)
	}

	router := gin.Default()
	router.Use(middleware.CORS())
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"estado": "UP"})
	})

	for _, ruta := range config.Rutas {
		h := proxy.Handler(ruta, resolver, secreto)
		router.Any(ruta.Prefijo, h)
		router.Any(ruta.Prefijo+"/*resto", h)
	}
	router.Any(config.PrefijoGrpcWeb+"/*metodo", proxy.HandlerEnvoy(urlEnvoy))

	log.Println("==================================================")
	log.Printf("   API Gateway escuchando en :%s", puerto)
	log.Printf("   Directorio Consul: %s", consulAddr)
	log.Printf("   Proxy Envoy (gRPC-Web): %s", urlEnvoy)
	log.Println("==================================================")
	if err := router.Run(":" + puerto); err != nil {
		log.Fatal(err)
	}
}

func env(clave, porDefecto string) string {
	if v := os.Getenv(clave); v != "" {
		return v
	}
	return porDefecto
}
