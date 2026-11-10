package controladores

import (
	"errors"
	"net/http"

	"audios.local/servidor-audios/fachada"
	"github.com/gin-gonic/gin"
)

type ControladorAudios struct {
	fachada *fachada.FachadaAudios
}

func NuevoControlador(f *fachada.FachadaAudios) *ControladorAudios {
	return &ControladorAudios{fachada: f}
}

// GET /tipos
func (c *ControladorAudios) ListarTipos(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, c.fachada.ListarTipos())
}

// GET /audios?tipo=musica
func (c *ControladorAudios) ListarAudios(ctx *gin.Context) {
	tipo := ctx.Query("tipo")
	audios, err := c.fachada.ListarPorTipo(tipo)
	if errors.Is(err, fachada.ErrTipoInvalido) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, audios)
}

// GET /audios/:id
func (c *ControladorAudios) ObtenerAudio(ctx *gin.Context) {
	id := ctx.Param("id")
	audio, err := c.fachada.ObtenerDetalle(id)
	if errors.Is(err, fachada.ErrNoEncontrado) {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, audio)
}
