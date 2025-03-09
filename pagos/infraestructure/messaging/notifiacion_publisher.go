package messaging

import (
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Notificacion representa la estructura de una notificación
type Notificacion struct {
	PedidoID uint   `json:"pedido_id"` // ID del pedido
	Mensaje  string `json:"mensaje"`   // Mensaje de la notificación
}

// NotificacionPublisher es el publicador de notificaciones a RabbitMQ
type NotificacionPublisher struct {
	conn    *amqp.Connection // Conexión a RabbitMQ
	channel *amqp.Channel    // Canal de RabbitMQ
	queue   amqp.Queue       // Cola de notificaciones
}

// NewNotificacionPublisher inicializa el publicador de notificaciones
func NewNotificacionPublisher() (*NotificacionPublisher, error) {
	// Conectar a RabbitMQ
	conn, err := amqp.Dial("amqp://dvelazquez:laconia@75.101.219.208:5672/")
	if err != nil {
		return nil, err
	}

	// Crear un canal
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	// Declarar la cola de notificaciones
	q, err := ch.QueueDeclare(
		"notificaciones", // Nombre de la cola
		true,             // Durable
		false,            // Auto-delete
		false,            // Exclusive
		false,            // No-wait
		nil,              // Argumentos
	)
	if err != nil {
		return nil, err
	}

	// Retornar el publicador inicializado
	return &NotificacionPublisher{
		conn:    conn,
		channel: ch,
		queue:   q,
	}, nil
}

// Publish envía una notificación a la cola de RabbitMQ
func (p *NotificacionPublisher) Publish(notificacion Notificacion) error {
	// Serializar la notificación como JSON
	body, err := json.Marshal(notificacion)
	if err != nil {
		return err
	}

	// Publicar el mensaje en la cola
	err = p.channel.Publish(
		"",            // Exchange
		p.queue.Name,  // Routing key
		false,         // Mandatory
		false,         // Immediate
		amqp.Publishing{
			ContentType: "application/json", // Tipo de contenido: JSON
			Body:        body,               // Cuerpo del mensaje (JSON)
		},
	)
	if err != nil {
		return err
	}

	// Log del mensaje enviado
	log.Printf("Notificación enviada: %+v", notificacion)
	return nil
}

// Close cierra la conexión y el canal de RabbitMQ
func (p *NotificacionPublisher) Close() {
	p.channel.Close()
	p.conn.Close()
}