package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"servidor-suscripciones/services"
)

type SuscripcionController struct {
	servicio *services.SuscripcionService
}

func NuevoSuscripcionController(servicio *services.SuscripcionService) *SuscripcionController {
	return &SuscripcionController{servicio: servicio}
}

func (ctl *SuscripcionController) Registrar(rg *gin.RouterGroup) {
	rg.GET("/planes", ctl.listar)
}

// GET /suscripciones/planes
func (ctl *SuscripcionController) listar(c *gin.Context) {
	c.JSON(http.StatusOK, ctl.servicio.Listar())
}
