// Package proxy reenvía las peticiones del gateway a los microservicios y a Envoy.
package proxy

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"

	"api-gateway/config"
	"api-gateway/discovery"
)

// HeaderGateway es la cabecera que los microservicios exigen en su filtro.
const HeaderGateway = "X-Gateway-Token"

// quitarCORS elimina las cabeceras CORS de la respuesta del servicio interno:
// las pone el gateway (si se duplican, el navegador rechaza la respuesta).
func quitarCORS(resp *http.Response) error {
	for clave := range resp.Header {
		if strings.HasPrefix(strings.ToLower(clave), "access-control-") {
			resp.Header.Del(clave)
		}
	}
	return nil
}

func errorProxy(servicio string) func(http.ResponseWriter, *http.Request, error) {
	return func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("[gateway] error hacia %s: %v", servicio, err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(`{"error":"el servicio ` + servicio + ` no respondió"}`))
	}
}

// Handler crea un manejador Gin que busca el microservicio en Consul y le reenvía
// la petición (incluidas las conexiones WebSocket, que ReverseProxy soporta).
func Handler(ruta config.Ruta, resolver *discovery.Resolver, secreto string) gin.HandlerFunc {
	return func(c *gin.Context) {
		destino, err := resolver.Resolver(ruta.Servicio)
		if err != nil {
			log.Printf("[gateway] %v", err)
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
			return
		}

		rp := &httputil.ReverseProxy{
			Rewrite: func(pr *httputil.ProxyRequest) {
				pr.SetURL(destino)
				// /api/audios/7 -> /audios/7
				pr.Out.URL.Path = ruta.Destino + strings.TrimPrefix(pr.In.URL.Path, ruta.Prefijo)
				pr.Out.URL.RawPath = ""
				pr.SetXForwarded()
				// Marca la petición como proveniente del gateway (sobrescribe cualquier valor del cliente).
				pr.Out.Header.Set(HeaderGateway, secreto)
				// El CORS ya lo resolvió el gateway: sin Origin, el microservicio trata la petición
				// como del mismo origen (evita que Spring rechace el WebSocket abierto desde file://).
				pr.Out.Header.Del("Origin")
			},
			ModifyResponse: quitarCORS,
			ErrorHandler:   errorProxy(ruta.Servicio),
			FlushInterval:  -1,
		}

		tipo := "HTTP"
		if strings.EqualFold(c.GetHeader("Upgrade"), "websocket") {
			tipo = "WebSocket"
		}
		log.Printf("[gateway] %s %s %s -> %s (%s)", tipo, c.Request.Method, c.Request.URL.Path, destino, ruta.Servicio)
		rp.ServeHTTP(c.Writer, c.Request)
	}
}

// HandlerEnvoy reenvía las llamadas gRPC-Web al Proxy Envoy. FlushInterval -1 hace
// que cada fragmento de audio se entregue al navegador apenas llega (streaming).
func HandlerEnvoy(urlEnvoy string) gin.HandlerFunc {
	destino, err := url.Parse(urlEnvoy)
	if err != nil {
		log.Fatalf("URL de Envoy inválida %q: %v", urlEnvoy, err)
	}
	rp := &httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			pr.SetURL(destino)
			pr.SetXForwarded()
		},
		ModifyResponse: quitarCORS,
		ErrorHandler:   errorProxy("proxy-envoy"),
		FlushInterval:  -1,
	}
	return func(c *gin.Context) {
		log.Printf("[gateway] gRPC-Web %s %s -> %s (proxy-envoy)", c.Request.Method, c.Request.URL.Path, destino)
		rp.ServeHTTP(c.Writer, c.Request)
	}
}
