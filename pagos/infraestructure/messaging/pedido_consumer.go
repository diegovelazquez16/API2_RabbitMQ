package messaging

import (
	"encoding/json"
	"fmt"
	"log"

	"api2/pagos/aplication/usecase"
	"api2/pagos/domain/models"

	amqp "github.com/rabbitmq/amqp091-go"
)

type PedidoConsumer struct {
	PagoUseCase *usecase.CreatePagoUseCase
	conn        *amqp.Connection
	channel     *amqp.Channel
	queue       amqp.Queue
}

// Nueva instancia del consumidor de pedidos
func NewPedidoConsumer(pagoUseCase *usecase.CreatePagoUseCase) (*PedidoConsumer, error) {
	conn, err := amqp.Dial("amqp://dvelazquez:laconia@75.101.219.208:5672/")
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	q, err := ch.QueueDeclare(
		"pedidos", // Cola de pedidos
		true,      // Durable
		false,     // Auto-delete
		false,     // Exclusive
		false,     // No-wait
		nil,       // Arguments
	)
	if err != nil {
		return nil, err
	}

	return &PedidoConsumer{
		PagoUseCase: pagoUseCase,
		conn:        conn,
		channel:     ch,
		queue:       q,
	}, nil
}

// Iniciar el consumo de mensajes de la cola de pedidos
func (c *PedidoConsumer) StartConsuming() {
	msgs, err := c.channel.Consume(
		c.queue.Name, // queue
		"",           // consumer
		true,         // auto-ack
		false,        // exclusive
		false,        // no-local
		false,        // no-wait
		nil,          // args
	)
	if err != nil {
		log.Fatalf("Error al consumir mensajes: %v", err)
	}

	go func() {
		for d := range msgs {
			// Usar un mapa genérico en lugar de una estructura de Pedido
			var pedidoData map[string]interface{}
			err := json.Unmarshal(d.Body, &pedidoData)
			if err != nil {
				log.Printf("Error al deserializar el pedido: %v", err)
				continue
			}

			log.Printf("Pedido recibido: %+v", pedidoData)

			// Extraer los valores del pedido
			pedidoID, ok := pedidoData["id"].(float64) // JSON parsea números como float64
			if !ok {
				log.Println("Error: El campo 'id' no es válido")
				continue
			}

			total, ok := pedidoData["total"].(float64)
			if !ok {
				log.Println("Error: El campo 'total' no es válido")
				continue
			}

			// Crear pago con los datos extraídos
			pago := models.Pago{
				PedidoID: uint(pedidoID),
				Monto:    total,
				Metodo:   "Tarjeta", // Puedes cambiarlo según el contexto
				Estado:   "Procesado",
			}

			err = c.PagoUseCase.Execute(&pago)
			if err != nil {
				log.Printf("Error al procesar pago: %v", err)
				continue
			}

			log.Printf("Pago procesado con éxito: %+v", pago)

			// Enviar notificación de pago completado
			notificacionPublisher, err := NewNotificacionPublisher()
			if err != nil {
				log.Printf("Error al conectar con RabbitMQ para notificaciones: %v", err)
				continue
			}
			defer notificacionPublisher.Close()

			// Crear mensaje de notificación
			notificacionMsg := fmt.Sprintf("Pago completado para el pedido ID: %d", uint(pedidoID))
			err = notificacionPublisher.Publish(notificacionMsg)
			if err != nil {
				log.Printf("Error al enviar notificación: %v", err)
			}
		}
	}()

	log.Println("Esperando pedidos...")
	select {} // Mantiene el proceso corriendo
}

// Cerrar conexiones al salir
func (c *PedidoConsumer) Close() {
	c.channel.Close()
	c.conn.Close()
}
