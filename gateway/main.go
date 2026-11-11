package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

// Tabla de enrutamiento: prefijo -> nombre lógico en Consul + respaldo fijo.
var rutas = []struct {
	prefijo  string
	servicio string
	respaldo string
}{
	{"/audios/", "servicio-audios", "http://localhost:5001"},
	{"/reacciones/", "servicio-reacciones", "http://localhost:5002"},
	{"/suscripciones/", "servicio-suscripciones", "http://localhost:5003"},
	{"/estadisticas/", "servicio-estadisticas", "http://localhost:5004"},
}

// resolver pregunta a Consul por una instancia SANA del servicio.
// Si Consul no responde o no hay instancias sanas, usa la dirección fija.
func resolver(servicio, respaldo string) string {
	cliente := http.Client{Timeout: 400 * time.Millisecond}
	respuesta, err := cliente.Get("http://localhost:8500/v1/health/service/" + servicio + "?passing=true")
	if err != nil {
		return respaldo
	}
	defer respuesta.Body.Close()
	var instancias []struct {
		Service struct {
			Port int
		}
	}
	if json.NewDecoder(respuesta.Body).Decode(&instancias) != nil || len(instancias) == 0 {
		return respaldo
	}
	// Todo corre en esta máquina: usamos el puerto que Consul reporta.
	return fmt.Sprintf("http://localhost:%d", instancias[0].Service.Port)
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// El preflight CORS se responde aquí mismo, sin reenviar.
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Idempotency-Key")
			w.WriteHeader(http.StatusNoContent)
			return
		}

		for _, ruta := range rutas {
			if strings.HasPrefix(r.URL.Path, ruta.prefijo) {
				base := resolver(ruta.servicio, ruta.respaldo)
				destino, err := url.Parse(base)
				if err != nil {
					http.Error(w, "destino inválido", http.StatusInternalServerError)
					return
				}
				// Se quita el prefijo: /audios/tipos -> /tipos
				r.URL.Path = strings.TrimPrefix(r.URL.Path, strings.TrimSuffix(ruta.prefijo, "/"))
				log.Printf("%s %s -> %s%s", r.Method, ruta.prefijo, base, r.URL.Path)

				// Proxy inverso: reenvía la petición tal cual (añade X-Forwarded-For).
				// También reenvía WebSockets (Upgrade) hacia el servidor de reacciones.
				proxy := httputil.NewSingleHostReverseProxy(destino)
				proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
					log.Printf("ERROR reenviando a %s: %v", base, err)
					w.Header().Set("Access-Control-Allow-Origin", "*")
					http.Error(w, "servicio no disponible", http.StatusBadGateway)
				}
				proxy.ServeHTTP(w, r)
				return
			}
		}
		http.NotFound(w, r)
	})

	log.Println("API Gateway escuchando en :8000")
	log.Fatal(http.ListenAndServe(":8000", nil))
}
