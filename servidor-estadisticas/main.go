package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	contador := NuevoContador()
	iniciarConsumidor(contador)

	r := gin.Default()
	r.Use(middlewareCORS())

	r.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"estado": "ok"})
	})

	// Ruta de degradación: recibe el evento por HTTP si no hay cola.
	r.POST("/eventos", func(ctx *gin.Context) {
		var ev Evento
		if err := ctx.ShouldBindJSON(&ev); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "evento inválido"})
			return
		}
		contador.Sumar(ev.Usuario, ev.AudioID)
		log.Printf("Reproducción contada (HTTP): usuario=%s audio=%s", ev.Usuario, ev.AudioID)
		ctx.JSON(http.StatusAccepted, gin.H{"estado": "registrado"})
	})

	// Opción d: qué audios reprodujo el usuario y cuántas veces.
	r.GET("/usuarios/:usuario", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, contador.Listar(ctx.Param("usuario")))
	})

	registrarEnConsul("servicio-estadisticas", 5004)

	log.Println("Servidor de estadísticas escuchando en :5004")
	if err := r.Run(":5004"); err != nil {
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

func registrarEnConsul(nombre string, puerto int) {
	registro := map[string]interface{}{
		"Name": nombre, "ID": nombre, "Address": "host.docker.internal", "Port": puerto,
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
