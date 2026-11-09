// Package componenteconexioncola publica en RabbitMQ los datos de cada reproducción.
package componenteconexioncola

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/streadway/amqp"
	"servidor.local/grpc-servidor/dtos"
)

// NombreCola es la cola compartida con el servidor de estadísticas.
const NombreCola = "estadisticas_reproducciones"

// urlPorDefecto apunta al bróker local; en otra máquina use la variable RABBITMQ_URL,
// por ejemplo amqp://admin:1234@192.168.0.15:5672/
const urlPorDefecto = "amqp://admin:1234@localhost:5672/"

// RabbitPublisher guarda la conexión, el canal y la cola declarada.
type RabbitPublisher struct {
	mu      sync.Mutex // el canal de amqp no es seguro para varias goroutines
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   amqp.Queue
}

// NewRabbitPublisher crea una conexión con RabbitMQ y declara la cola.
func NewRabbitPublisher() (*RabbitPublisher, error) {
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

	return &RabbitPublisher{conn: conn, channel: ch, queue: q}, nil
}

// PublicarNotificacion serializa el DTO en JSON y lo envía a la cola.
func (p *RabbitPublisher) PublicarNotificacion(msg dtos.NotificacionReproduccionDTO) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("error convirtiendo mensaje a JSON: %v", err)
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	err = p.channel.Publish(
		"",           // exchange
		p.queue.Name, // routing key (nombre de la cola)
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
		},
	)
	if err != nil {
		return fmt.Errorf("error publicando el mensaje: %v", err)
	}

	fmt.Printf("[ECO COLA] Notificación enviada a la cola %s: %s\n", p.queue.Name, string(body))
	return nil
}

// Cerrar libera el canal y la conexión.
func (p *RabbitPublisher) Cerrar() {
	p.channel.Close()
	p.conn.Close()
}
