# Cómo aplicar la extensión al servidor de streaming del profesor

1. Copia `capaPublicador/publicador.go` dentro de `servidor/capaPublicador/`.
2. Reemplaza `servidor/capaControladores/controladorEnvioAudio.go` por el de esta carpeta.
3. En la carpeta `servidor/` ejecuta:

   go get github.com/rabbitmq/amqp091-go
   go mod tidy

4. Arranca normal: `go run main/servidor.go`

Qué agrega:
- Lee el metadato `x-usuario` de la llamada gRPC-Web (sin tocar el .proto).
- Al terminar un envío exitoso publica {usuario, audioId} en la cola
  `reproducciones` de RabbitMQ (o por HTTP directo a estadísticas si no hay cola).
