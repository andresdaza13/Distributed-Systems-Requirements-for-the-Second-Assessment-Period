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
	fachada := NuevaFachadaPagos()

	r := gin.Default()
	r.Use(middlewareCORS())

	r.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"estado": "ok"})
	})

	r.GET("/planes", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, Planes)
	})

	// POST /pagos con cabecera Idempotency-Key (opciones c y e)
	r.POST("/pagos", func(ctx *gin.Context) {
		token := ctx.GetHeader("Idempotency-Key")
		if token == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "falta la cabecera Idempotency-Key"})
			return
		}
		var pago Pago
		if err := ctx.ShouldBindJSON(&pago); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "cuerpo inválido: " + err.Error()})
			return
		}
		resultado, repetido := fachada.Pagar(token, pago)
		if repetido {
			ctx.JSON(http.StatusOK, resultado) // ya existía: se devuelve lo guardado
			return
		}
		ctx.JSON(http.StatusCreated, resultado)
	})

	r.GET("/pagos/:usuario", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, fachada.HistorialDe(ctx.Param("usuario")))
	})

	registrarEnConsul("servicio-suscripciones", 5003)

	log.Println("Servidor de suscripciones y pagos escuchando en :5003")
	if err := r.Run(":5003"); err != nil {
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
