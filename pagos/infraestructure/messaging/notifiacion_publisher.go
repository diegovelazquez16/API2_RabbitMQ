package messaging

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type NotificacionPublisher struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   amqp.Queue
}

func NewNotificacionPublisher() (*NotificacionPublisher, error) {
	conn, err := amqp.Dial("amqp://dvelazquez:laconia@75.101.219.208:5672/")
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	q, err := ch.QueueDeclare(
		"notificaciones", // Nombre de la cola
		true,             // Durable
		false,            // Auto-delete
		false,            // Exclusive
		false,            // No-wait
		nil,              // Arguments
	)
	if err != nil {
		return nil, err
	}

	return &NotificacionPublisher{
		conn:    conn,
		channel: ch,
		queue:   q,
	}, nil
}

func (p *NotificacionPublisher) Publish(message string) error {
	err := p.channel.Publish(
		"",            // Exchange
		p.queue.Name,  // Routing key
		false,         // Mandatory
		false,         // Immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		},
	)
	if err != nil {
		return err
	}

	log.Printf("Notificación enviada: %s", message)
	return nil
}

func (p *NotificacionPublisher) Close() {
	p.channel.Close()
	p.conn.Close()
}
