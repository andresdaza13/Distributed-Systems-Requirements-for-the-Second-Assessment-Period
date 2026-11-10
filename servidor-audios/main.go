package main

import (
	"log"
	"net/http"

	"audios.local/servidor-audios/controladores"
	"audios.local/servidor-audios/fachada"
	"audios.local/servidor-audios/repositorio"
	"github.com/gin-gonic/gin"
)

func main() {
	repo := repositorio.NuevoRepositorio()
	fach := fachada.NuevaFachada(repo)
	ctrl := controladores.NuevoControlador(fach)

	r := gin.Default()
	r.Use(middlewareCORS())

	r.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"estado": "ok"})
	})
	r.GET("/tipos", ctrl.ListarTipos)
	r.GET("/audios", ctrl.ListarAudios)
	r.GET("/audios/:id", ctrl.ObtenerAudio)

	log.Println("Servidor de audios y metadatos escuchando en :5001")
	if err := r.Run(":5001"); err != nil {
		log.Fatalf("No se pudo iniciar el servidor: %v", err)
	}
}

func middlewareCORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Idempotency-Key")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
