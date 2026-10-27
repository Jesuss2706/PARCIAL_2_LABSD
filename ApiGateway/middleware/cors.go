// Package middleware contiene los filtros del API Gateway.
package middleware

import "net/http"

// CORS permite que el cliente web (abierto como archivo o desde otro puerto)
// consuma los servicios. El gateway es el único que agrega cabeceras CORS:
// las que envíen los microservicios o Envoy se eliminan en el proxy para no duplicarlas.
//
// Se refleja el Origin y se permiten credenciales porque SockJS (reacciones)
// hace sus peticiones con withCredentials.
func CORS(siguiente http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origen := r.Header.Get("Origin")
		if origen != "" {
			w.Header().Set("Access-Control-Allow-Origin", origen)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Expose-Headers", "grpc-status, grpc-message, grpc-status-details-bin")
			w.Header().Add("Vary", "Origin")
		}

		// Petición de verificación previa (preflight): la responde el gateway.
		if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			if cabeceras := r.Header.Get("Access-Control-Request-Headers"); cabeceras != "" {
				w.Header().Set("Access-Control-Allow-Headers", cabeceras)
			}
			w.Header().Set("Access-Control-Max-Age", "1728000")
			w.WriteHeader(http.StatusNoContent)
			return
		}
		siguiente.ServeHTTP(w, r)
	})
}
