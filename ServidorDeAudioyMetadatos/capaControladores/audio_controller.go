// Package capaControladores recibe las peticiones REST, imprime el eco y delega en la fachada.
package capaControladores

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"servidor-audios/capaFachadaServices"
	"servidor-audios/capaFachadaServices/dto"
)

// tamanoMaximo del mp3 que acepta el servidor (200 MB).
const tamanoMaximo = 200 << 20

type AudioController struct {
	servicio *capaFachadaServices.AudioService
}

func NuevoAudioController(servicio *capaFachadaServices.AudioService) *AudioController {
	return &AudioController{servicio: servicio}
}

func (ctl *AudioController) RegistrarRutas(r *gin.Engine) {
	r.GET("/tipos", ctl.listarTipos)
	r.GET("/tipos/:idTipo/metadatos", ctl.listarMetadatos)
	r.GET("/tipos/:idTipo/audios", ctl.listarAudios)
	r.GET("/audios/:idAudio", ctl.detalle)
	r.POST("/audios", ctl.registrar)
}

// GET /tipos
func (ctl *AudioController) listarTipos(c *gin.Context) {
	log.Printf("[ECO REST] GET /tipos -> listar tipos de audio")
	c.JSON(http.StatusOK, ctl.servicio.ListarTipos())
}

// GET /tipos/:idTipo/metadatos
func (ctl *AudioController) listarMetadatos(c *gin.Context) {
	log.Printf("[ECO REST] GET /tipos/%s/metadatos -> metadatos propios del tipo", c.Param("idTipo"))
	id, ok := leerEntero(c, "idTipo")
	if !ok {
		return
	}
	resp, err := ctl.servicio.ListarMetadatosDeTipo(id)
	responder(c, resp, err)
}

// GET /tipos/:idTipo/audios
func (ctl *AudioController) listarAudios(c *gin.Context) {
	log.Printf("[ECO REST] GET /tipos/%s/audios -> audios del tipo", c.Param("idTipo"))
	id, ok := leerEntero(c, "idTipo")
	if !ok {
		return
	}
	resp, err := ctl.servicio.ListarAudiosDeTipo(id)
	responder(c, resp, err)
}

// GET /audios/:idAudio
func (ctl *AudioController) detalle(c *gin.Context) {
	log.Printf("[ECO REST] GET /audios/%s -> detalle del audio", c.Param("idAudio"))
	id, ok := leerEntero(c, "idAudio")
	if !ok {
		return
	}
	resp, err := ctl.servicio.ObtenerDetalle(id)
	responder(c, resp, err)
}

// POST /audios (multipart/form-data)
//
//	idTipo     id del tipo de audio
//	titulo     título del audio
//	duracion   (opcional) duración, p. ej. 3:45
//	metadatos  JSON {"idMetadato": "valor", ...} con los metadatos propios del tipo
//	archivo    el mp3
func (ctl *AudioController) registrar(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, tamanoMaximo+(1<<20))
	log.Printf("[ECO REST] POST /audios -> registrar audio (titulo='%s', idTipo=%s)", c.PostForm("titulo"), c.PostForm("idTipo"))

	idTipo, err := strconv.Atoi(c.PostForm("idTipo"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorDTO{Error: "el campo idTipo debe ser un número"})
		return
	}
	metadatosTexto := map[string]string{}
	if crudo := c.PostForm("metadatos"); crudo != "" {
		if err := json.Unmarshal([]byte(crudo), &metadatosTexto); err != nil {
			c.JSON(http.StatusBadRequest, dto.ErrorDTO{Error: "el campo metadatos debe ser un JSON {\"idMetadato\": \"valor\"}"})
			return
		}
	}
	metadatos := map[int]string{}
	for clave, valor := range metadatosTexto {
		id, err := strconv.Atoi(clave)
		if err != nil {
			c.JSON(http.StatusBadRequest, dto.ErrorDTO{Error: "id de metadato inválido: " + clave})
			return
		}
		metadatos[id] = valor
	}

	cabecera, err := c.FormFile("archivo")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorDTO{Error: "debe adjuntar el archivo mp3 en el campo 'archivo'"})
		return
	}
	if cabecera.Size > tamanoMaximo {
		c.JSON(http.StatusRequestEntityTooLarge, dto.ErrorDTO{Error: "el mp3 supera el tamaño máximo de 200 MB"})
		return
	}
	archivo, err := cabecera.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorDTO{Error: "no fue posible leer el archivo adjunto"})
		return
	}
	defer archivo.Close()

	resp, err := ctl.servicio.RegistrarAudio(dto.SolicitudRegistroDTO{
		IdTipo: idTipo, Titulo: c.PostForm("titulo"), Duracion: c.PostForm("duracion"), Metadatos: metadatos,
	}, cabecera.Filename, cabecera.Size, archivo)
	if err != nil {
		responder(c, nil, err)
		return
	}
	log.Printf("[ECO REST] Respuesta: audio #%d registrado (%s)", resp.Id, resp.NombreArchivo)
	c.JSON(http.StatusCreated, resp)
}

func leerEntero(c *gin.Context, param string) (int, bool) {
	id, err := strconv.Atoi(c.Param(param))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorDTO{Error: "el parámetro " + param + " debe ser un número"})
		return 0, false
	}
	return id, true
}

func responder(c *gin.Context, cuerpo any, err error) {
	if err != nil {
		var e *capaFachadaServices.ErrorServicio
		if errors.As(err, &e) {
			log.Printf("[ECO REST] Error %d: %s", e.Codigo, e.Mensaje)
			c.JSON(e.Codigo, dto.ErrorDTO{Error: e.Mensaje})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorDTO{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, cuerpo)
}
