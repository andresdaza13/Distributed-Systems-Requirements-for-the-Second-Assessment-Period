package capaPublicador

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

// EventoReproduccion es el mensaje que viaja hacia el servicio de estadísticas.
type EventoReproduccion struct {
	Usuario string `json:"usuario"`
	AudioID string `json:"audioId"`
}

// Publicador envía eventos de reproducción de forma asíncrona.
// Preferencia: cola RabbitMQ (desacoplamiento temporal). Si la cola no está
// disponible, degrada a un POST HTTP directo al servicio de estadísticas.
type Publicador struct {
	mu    sync.Mutex
	canal *amqp.Channel
}

var (
	objInstancia *Publicador
	once         sync.Once
)

// CrearUnicaInstancia implementa el patrón Singleton (igual que el logger).
func CrearUnicaInstancia() *Publicador {
	once.Do(func() {
		p := &Publicador{}
		conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
		if err != nil {
			fmt.Println("AVISO: RabbitMQ no disponible, los eventos se enviarán por HTTP directo:", err)
		} else {
			canal, err2 := conn.Channel()
			if err2 != nil {
				fmt.Println("AVISO: no se pudo abrir el canal AMQP:", err2)
			} else {
				_, err3 := canal.QueueDeclare("reproducciones", true, false, false, false, nil)
				if err3 != nil {
					fmt.Println("AVISO: no se pudo declarar la cola:", err3)
				} else {
					p.canal = canal
					fmt.Println("Conectado a RabbitMQ: cola 'reproducciones' lista")
				}
			}
		}
		objInstancia = p
	})
	return objInstancia
}

// Publicar envía el evento y retorna de inmediato para quien lo llama con `go`.
func (p *Publicador) Publicar(ev EventoReproduccion) {
	cuerpo, _ := json.Marshal(ev)

	p.mu.Lock()
	canal := p.canal
	p.mu.Unlock()

	if canal != nil {
		err := canal.PublishWithContext(context.Background(), "", "reproducciones", false, false,
			amqp.Publishing{
				ContentType:  "application/json",
				DeliveryMode: amqp.Persistent,
				Body:         cuerpo,
			})
		if err == nil {
			fmt.Println("Evento publicado en la cola:", string(cuerpo))
			return
		}
		fmt.Println("Error publicando en la cola, degradando a HTTP:", err)
	}

	resp, err := http.Post("http://localhost:5004/eventos", "application/json", bytes.NewReader(cuerpo))
	if err != nil {
		fmt.Println("AVISO: no se pudo entregar el evento a estadísticas:", err)
		return
	}
	resp.Body.Close()
	fmt.Println("Evento entregado a estadísticas por HTTP:", string(cuerpo))
}
