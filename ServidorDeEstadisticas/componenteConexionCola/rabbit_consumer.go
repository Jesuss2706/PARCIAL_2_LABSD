// Package componenteconexioncola consume de RabbitMQ los mensajes de reproducción.
package componenteconexioncola

import (
	"fmt"
	"os"

	"github.com/streadway/amqp"
)

// NombreCola debe ser la misma cola en la que publica el servidor de streaming.
const NombreCola = "estadisticas_reproducciones"

// urlPorDefecto apunta al bróker local; en otra máquina use la variable RABBITMQ_URL.
const urlPorDefecto = "amqp://admin:1234@localhost:5672/"

// RabbitConsumer guarda la conexión, el canal y la cola declarada.
type RabbitConsumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   amqp.Queue
}

// NewRabbitConsumer se conecta al bróker y declara la cola (con los mismos
// parámetros que el publicador, si no RabbitMQ rechaza la declaración).
func NewRabbitConsumer() (*RabbitConsumer, error) {
	url := os.Getenv("RABBITMQ_URL")
	if url == "" {
		url = urlPorDefecto
	}

	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("error conectando a RabbitMQ: %v", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("error abriendo el canal: %v", err)
	}
	q, err := ch.QueueDeclare(
		NombreCola, // nombre de la cola
		true,       // durable
		false,      // auto-delete
		false,      // exclusive
		false,      // no-wait
		nil,        // argumentos
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("error declarando la cola: %v", err)
	}
	return &RabbitConsumer{conn: conn, channel: ch, queue: q}, nil
}

// Consumir queda escuchando la cola y ejecuta procesar por cada mensaje.
// Solo confirma (ack) el mensaje si se procesó correctamente.
func (c *RabbitConsumer) Consumir(procesar func([]byte) error) error {
	mensajes, err := c.channel.Consume(
		c.queue.Name, // cola
		"",           // consumidor
		false,        // auto-ack
		false,        // exclusive
		false,        // no-local
		false,        // no-wait
		nil,          // argumentos
	)
	if err != nil {
		return fmt.Errorf("error registrando el consumidor: %v", err)
	}

	for mensaje := range mensajes {
		fmt.Printf("\n[ECO COLA] Mensaje consumido de la cola %s: %s\n", c.queue.Name, string(mensaje.Body))
		if err := procesar(mensaje.Body); err != nil {
			fmt.Println("Error procesando el mensaje:", err)
			mensaje.Nack(false, false) // se descarta para no reprocesarlo en bucle
			continue
		}
		mensaje.Ack(false)
	}
	return fmt.Errorf("la conexión con RabbitMQ se cerró")
}

// Cerrar libera el canal y la conexión.
func (c *RabbitConsumer) Cerrar() {
	c.channel.Close()
	c.conn.Close()
}
