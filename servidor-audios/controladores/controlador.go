package controladores

import (
	"encoding/json"
	"errors"
	"net/http"
	"path/filepath"
	"strings"

	"audios.local/servidor-audios/fachada"
	"audios.local/servidor-audios/modelo"
	"github.com/gin-gonic/gin"
)

type ControladorAudios struct {
	fachada       *fachada.FachadaAudios
	rutaCanciones string
}

func NuevoControlador(f *fachada.FachadaAudios, rutaCanciones string) *ControladorAudios {
	return &ControladorAudios{fachada: f, rutaCanciones: rutaCanciones}
}

// GET /tipos
func (c *ControladorAudios) ListarTipos(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, c.fachada.ListarTipos())
}

// GET /tipos/:tipo/campos  (el formulario del admin lo usa para pintarse)
func (c *ControladorAudios) ListarCampos(ctx *gin.Context) {
	campos, err := c.fachada.CamposDeTipo(ctx.Param("tipo"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, campos)
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

// POST /audios  (opción f: multipart/form-data con metadatos + archivo binario)
func (c *ControladorAudios) RegistrarAudio(ctx *gin.Context) {
	titulo := ctx.PostForm("titulo")
	tipo := ctx.PostForm("tipo")

	var metadatos map[string]string
	if err := json.Unmarshal([]byte(ctx.PostForm("metadatos")), &metadatos); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "el campo 'metadatos' no es JSON válido"})
		return
	}

	archivo, err := ctx.FormFile("archivo")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "falta el archivo de audio"})
		return
	}
	nombre := filepath.Base(archivo.Filename)
	extension := strings.TrimPrefix(filepath.Ext(nombre), ".")
	base := strings.TrimSuffix(nombre, filepath.Ext(nombre))

	audio := modelo.Audio{
		Titulo: titulo, Tipo: tipo,
		Archivo: base, Formato: extension,
		Metadatos: metadatos,
	}

	creado, err := c.fachada.Registrar(audio)
	if errors.Is(err, fachada.ErrTipoInvalido) || errors.Is(err, fachada.ErrMetadatosIncompletos) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Se guarda donde el servidor de streaming lee (almacenamiento compartido).
	if err := ctx.SaveUploadedFile(archivo, filepath.Join(c.rutaCanciones, nombre)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo guardar el archivo: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, creado)
}
