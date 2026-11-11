package main

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

// Mensaje es lo único que viaja por el WebSocket, en ambos sentidos.
type Mensaje struct {
	Accion  string `json:"accion"`  // suscribir | desuscribir | reaccion
	AudioID string `json:"audioId"` // el canal = el audio
	Usuario string `json:"usuario"`
	Emoji   string `json:"emoji"`
}

// Hub implementa pub/sub: un canal por audio, un conjunto de conexiones por canal.
type Hub struct {
	mu      sync.Mutex
	canales map[string]map[*websocket.Conn]bool
}

func NuevoHub() *Hub {
	return &Hub{canales: make(map[string]map[*websocket.Conn]bool)}
}

// Suscribir: el listener de PLAY del cliente llega aquí.
// Suscribir: el listener de PLAY del cliente llega aquí.
// Regla del dominio: una conexión escucha UN solo audio a la vez,
// así que entrar a un canal la saca automáticamente de cualquier otro.
// El servidor impone la regla; no confía en que el cliente se desuscriba bien.
func (h *Hub) Suscribir(audioID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// sacar la conexión de todos los demás canales
	for otroCanal, conexiones := range h.canales {
		if otroCanal != audioID && conexiones[conn] {
			delete(conexiones, conn)
			log.Printf("Conexión movida: sale del canal '%s'", otroCanal)
		}
	}

	if h.canales[audioID] == nil {
		h.canales[audioID] = make(map[*websocket.Conn]bool)
	}
	h.canales[audioID][conn] = true
	log.Printf("Suscripción al canal '%s' (%d oyentes)", audioID, len(h.canales[audioID]))
}

// Desuscribir: el listener de PAUSA del cliente llega aquí.
func (h *Hub) Desuscribir(audioID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if conexiones, existe := h.canales[audioID]; existe {
		delete(conexiones, conn)
		log.Printf("Baja del canal '%s' (%d oyentes)", audioID, len(conexiones))
	}
}

// DesconectarTodo limpia una conexión cerrada de todos los canales.
func (h *Hub) DesconectarTodo(conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for audioID, conexiones := range h.canales {
		if conexiones[conn] {
			delete(conexiones, conn)
			log.Printf("Conexión retirada del canal '%s'", audioID)
		}
	}
}

// Publicar reenvía la reacción a TODOS los suscriptores del canal: el "callback".
// Todas las escrituras pasan por el mutex: una sola goroutine escribe a la vez.
func (h *Hub) Publicar(m Mensaje) {
	h.mu.Lock()
	defer h.mu.Unlock()
	cuerpo, _ := json.Marshal(m)
	entregados := 0
	for conn := range h.canales[m.AudioID] {
		if err := conn.WriteMessage(websocket.TextMessage, cuerpo); err != nil {
			conn.Close()
			delete(h.canales[m.AudioID], conn)
			continue
		}
		entregados++
	}
	log.Printf("Reacción '%s' de %s reenviada a %d clientes del canal '%s'",
		m.Emoji, m.Usuario, entregados, m.AudioID)
}
