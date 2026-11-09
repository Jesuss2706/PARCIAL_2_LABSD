// Package middleware contiene los filtros HTTP del API Gateway.
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORS permite que el cliente web y el administrador web (abiertos desde file:// o
// desde otro puerto) consuman el gateway. El gateway es el ÚNICO que pone cabeceras CORS:
// las que vengan de los microservicios o de Envoy se eliminan en el proxy para no duplicarlas.
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.Writer.Header()
		h.Set("Access-Control-Allow-Origin", "*")
		h.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		h.Set("Access-Control-Allow-Headers",
			"Content-Type, Authorization, X-User-Agent, X-Grpc-Web, Grpc-Timeout, Keep-Alive, Cache-Control, "+
				"Content-Transfer-Encoding, X-Accept-Content-Transfer-Encoding, X-Accept-Response-Streaming")
		h.Set("Access-Control-Expose-Headers", "Grpc-Status, Grpc-Message")
		h.Set("Access-Control-Max-Age", "1728000")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
