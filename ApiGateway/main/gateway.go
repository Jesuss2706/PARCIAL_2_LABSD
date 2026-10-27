// Punto de entrada del API Gateway: único punto de acceso del cliente web y
// del administrador. Resuelve los microservicios en el directorio Consul.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"servidor.local/api-gateway/configuracion"
	"servidor.local/api-gateway/descubrimiento"
	"servidor.local/api-gateway/middleware"
	"servidor.local/api-gateway/proxy"
)

func main() {
	puerto := variable("PORT", "8000")
	direccionConsul := variable("CONSUL_ADDR", "127.0.0.1:8500")
	secreto := variable("GATEWAY_SECRET", "secreto-gateway-2026")

	resolvedor := descubrimiento.NuevoResolvedor(direccionConsul)
	gateway := proxy.NuevoGateway(configuracion.Rutas, resolvedor, secreto)

	enrutador := http.NewServeMux()
	enrutador.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"estado": "UP"})
	})
	enrutador.Handle("/", gateway)

	mostrarBienvenida(puerto, direccionConsul)
	if err := http.ListenAndServe(":"+puerto, middleware.CORS(enrutador)); err != nil {
		log.Fatal(err)
	}
}

func variable(clave, porDefecto string) string {
	if valor := os.Getenv(clave); valor != "" {
		return valor
	}
	return porDefecto
}

func mostrarBienvenida(puerto, direccionConsul string) {
	fmt.Println("==================================================")
	fmt.Printf("   API Gateway escuchando en http://localhost:%s\n", puerto)
	fmt.Printf("   Directorio Consul: %s\n", direccionConsul)
	fmt.Println("==================================================")
	for _, ruta := range configuracion.Rutas {
		destino := ruta.Servicio + " (Consul)"
		if ruta.URLFija != "" {
			destino = ruta.URLFija
		}
		fmt.Printf(" %-20s -> %s\n", ruta.Prefijo, destino)
	}
	fmt.Println("==================================================")
}
