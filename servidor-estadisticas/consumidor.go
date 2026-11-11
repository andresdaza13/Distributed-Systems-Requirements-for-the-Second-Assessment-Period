package main

import (
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

// iniciarConsumidor se conecta a RabbitMQ y procesa la cola 'reproducciones'.
// Si la cola no está disponible, el servicio sigue vivo: los eventos pueden
// llegar igualmente por HTTP (POST /eventos), la ruta de degradación.
func iniciarConsumidor(contador *Contador) {
	conexion, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Println("AVISO: RabbitMQ no disponible; solo se recibirán eventos por HTTP:", err)
		return
	}
	canal, err := conexion.Channel()
	if err != nil {
		log.Println("AVISO: no se pudo abrir el canal AMQP:", err)
		return
	}
	if _, err := canal.QueueDeclare("reproducciones", true, false, false, false, nil); err != nil {
		log.Println("AVISO: no se pudo declarar la cola:", err)
		return
	}
	mensajes, err := canal.Consume("reproducciones", "", false, false, false, false, nil)
	if err != nil {
		log.Println("AVISO: no se pudo consumir la cola:", err)
		return
	}
	log.Println("Consumiendo la cola 'reproducciones' de RabbitMQ...")

	go func() {
		for msg := range mensajes {
			var ev Evento
			if err := json.Unmarshal(msg.Body, &ev); err != nil {
				log.Println("Mensaje inválido en la cola:", string(msg.Body))
				msg.Ack(false)
				continue
			}
			contador.Sumar(ev.Usuario, ev.AudioID)
			log.Printf("Reproducción contada: usuario=%s audio=%s", ev.Usuario, ev.AudioID)
			msg.Ack(false) // confirmación: la cola ya puede borrar el mensaje
		}
	}()
}
