package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	// La página se abre como file://, así que aceptamos cualquier origen (laboratorio).
	CheckOrigin: func(r *http.Request) bool { return true },
}

func main() {
	hub := NuevoHub()

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"estado":"ok"}`))
	})

	// Punto de entrada: HTTP con Upgrade -> 101 -> conexión abierta full-duplex.
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("Error en el upgrade a WebSocket:", err)
			return
		}
		log.Println("Cliente conectado:", conn.RemoteAddr())
		go atender(hub, conn)
	})

	registrarEnConsul("servicio-reacciones", 5002)

	log.Println("Servidor de reacciones (WebSocket) escuchando en :5002")
	log.Fatal(http.ListenAndServe(":5002", nil))
}

// atender corre en una goroutine por cliente: lee mensajes hasta que se desconecte.
func atender(hub *Hub, conn *websocket.Conn) {
	defer func() {
		hub.DesconectarTodo(conn)
		conn.Close()
		log.Println("Cliente desconectado:", conn.RemoteAddr())
	}()
	for {
		_, datos, err := conn.ReadMessage()
		if err != nil {
			return
		}
		var m Mensaje
		if err := json.Unmarshal(datos, &m); err != nil {
			log.Println("Mensaje inválido:", string(datos))
			continue
		}
		switch m.Accion {
		case "suscribir":
			hub.Suscribir(m.AudioID, conn)
		case "desuscribir":
			hub.Desuscribir(m.AudioID, conn)
		case "reaccion":
			hub.Publicar(m)
		}
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
