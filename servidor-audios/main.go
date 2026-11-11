package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"audios.local/servidor-audios/controladores"
	"audios.local/servidor-audios/fachada"
	"audios.local/servidor-audios/repositorio"
	"github.com/gin-gonic/gin"
)

// Carpeta de canciones del servidor de streaming (almacenamiento compartido).
const rutaCanciones = "../servidor/canciones"

func main() {
	repo := repositorio.NuevoRepositorio()
	fach := fachada.NuevaFachada(repo)
	ctrl := controladores.NuevoControlador(fach, rutaCanciones)

	r := gin.Default()
	r.Use(middlewareCORS())

	r.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"estado": "ok"})
	})
	r.GET("/tipos", ctrl.ListarTipos)
	r.GET("/tipos/:tipo/campos", ctrl.ListarCampos)
	r.GET("/audios", ctrl.ListarAudios)
	r.GET("/audios/:id", ctrl.ObtenerAudio)
	r.POST("/audios", ctrl.RegistrarAudio)

	registrarEnConsul("servicio-audios", 5001)

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

// registrarEnConsul anuncia este servicio en el registro (service discovery).
// Es opcional: si Consul no está arriba, el gateway usa direcciones fijas.
func registrarEnConsul(nombre string, puerto int) {
	registro := map[string]interface{}{
		"Name":    nombre,
		"ID":      nombre,
		"Address": "host.docker.internal",
		"Port":    puerto,
		"Check": map[string]string{
			"HTTP":     fmt.Sprintf("http://host.docker.internal:%d/health", puerto),
			"Interval": "10s",
		},
	}
	cuerpo, _ := json.Marshal(registro)
	peticion, _ := http.NewRequest(http.MethodPut,
		"http://localhost:8500/v1/agent/service/register", bytes.NewReader(cuerpo))
	peticion.Header.Set("Content-Type", "application/json")
	cliente := &http.Client{Timeout: 2 * time.Second}
	respuesta, err := cliente.Do(peticion)
	if err != nil {
		log.Println("AVISO: Consul no disponible, el gateway usará direcciones fijas:", err)
		return
	}
	respuesta.Body.Close()
	log.Printf("Registrado en Consul como '%s'", nombre)
}
