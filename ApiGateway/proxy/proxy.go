// Package proxy reenvía las peticiones al servicio correspondiente usando
// httputil.ReverseProxy (soporta REST, WebSocket y el streaming de gRPC-Web).
package proxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sort"
	"strings"

	"servidor.local/api-gateway/configuracion"
	"servidor.local/api-gateway/descubrimiento"
)

// CabeceraGateway es la cabecera que los microservicios exigen en su filtro.
const CabeceraGateway = "X-Gateway-Token"

// Gateway enruta cada petición según su prefijo.
type Gateway struct {
	rutas      []configuracion.Ruta
	resolvedor *descubrimiento.Resolvedor
	secreto    string
}

// NuevoGateway crea el gateway. Las rutas se ordenan de la más larga a la más corta.
func NuevoGateway(rutas []configuracion.Ruta, resolvedor *descubrimiento.Resolvedor, secreto string) *Gateway {
	ordenadas := append([]configuracion.Ruta(nil), rutas...)
	sort.Slice(ordenadas, func(i, j int) bool { return len(ordenadas[i].Prefijo) > len(ordenadas[j].Prefijo) })
	return &Gateway{rutas: ordenadas, resolvedor: resolvedor, secreto: secreto}
}

// ServeHTTP busca la ruta, resuelve el destino y reenvía la petición.
func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ruta, encontrada := g.buscarRuta(r.URL.Path)
	if !encontrada {
		responderError(w, http.StatusNotFound, "ruta no registrada en el API Gateway")
		return
	}

	destino, nombre, err := g.resolverDestino(ruta)
	if err != nil {
		fmt.Printf("[ECO GATEWAY] %s %s -> ERROR: %v\n", r.Method, r.URL.Path, err)
		responderError(w, http.StatusServiceUnavailable, err.Error())
		return
	}

	rutaInterna := ruta.Destino + strings.TrimPrefix(r.URL.Path, ruta.Prefijo)
	if rutaInterna == "" {
		rutaInterna = "/"
	}
	tipo := "HTTP"
	if strings.EqualFold(r.Header.Get("Upgrade"), "websocket") {
		tipo = "WebSocket"
	}
	fmt.Printf("[ECO GATEWAY] %s %s %s -> %s %s%s\n", tipo, r.Method, r.URL.Path, nombre, destino, rutaInterna)

	proxyInverso := &httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			pr.SetURL(destino)
			pr.Out.URL.Path = rutaInterna
			pr.Out.URL.RawPath = ""
			pr.SetXForwarded()
			// Marca la petición como proveniente del gateway (sobrescribe cualquier valor del cliente).
			pr.Out.Header.Set(CabeceraGateway, g.secreto)
		},
		ModifyResponse: func(respuesta *http.Response) error {
			eliminarCabecerasCORS(respuesta.Header)
			return nil
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			fmt.Printf("[ECO GATEWAY] Error hacia %s: %v\n", nombre, err)
			responderError(w, http.StatusBadGateway, "el servicio "+nombre+" no respondió")
		},
	}
	proxyInverso.ServeHTTP(w, r)
}

func (g *Gateway) buscarRuta(ruta string) (configuracion.Ruta, bool) {
	for _, candidata := range g.rutas {
		if ruta == candidata.Prefijo || strings.HasPrefix(ruta, candidata.Prefijo+"/") {
			return candidata, true
		}
	}
	return configuracion.Ruta{}, false
}

func (g *Gateway) resolverDestino(ruta configuracion.Ruta) (*url.URL, string, error) {
	if ruta.URLFija != "" {
		destino, err := url.Parse(ruta.URLFija)
		return destino, "proxy-envoy", err
	}
	destino, err := g.resolvedor.Resolver(ruta.Servicio)
	return destino, ruta.Servicio, err
}

// eliminarCabecerasCORS quita las cabeceras CORS del servicio; las pone el gateway.
func eliminarCabecerasCORS(cabeceras http.Header) {
	for clave := range cabeceras {
		if strings.HasPrefix(strings.ToLower(clave), "access-control-") {
			cabeceras.Del(clave)
		}
	}
}

func responderError(w http.ResponseWriter, codigo int, mensaje string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(codigo)
	fmt.Fprintf(w, "{\"error\":%q}\n", mensaje)
}
