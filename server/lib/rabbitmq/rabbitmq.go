package rabbitmq

import (
	"log"

	"github.com/streadway/amqp"
)

// Connect to RabbitMQ
func Connect() (*amqp.Connection, error) {
	return amqp.Dial("amqp://guest:guest@localhost:5672/")
}

// Initialize RabbitMQ publisher
func InitRabbitMq() {
	conn, err := Connect()
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}
	defer ch.Close()

	// Declare an exchange
	err = ch.ExchangeDeclare(
		"orders", // exchange name
		"direct", // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)

	if err != nil {
		log.Fatalf("Failed to declare an exchange: %v", err)
	}
}

// StartConsumer listens to the RabbitMQ queue and processes incoming messages
func StartConsumer(queueName string) {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}
	defer ch.Close()

	// Declare the queue from which messages are received
	q, err := ch.QueueDeclare(
		queueName,
		true,  // durable
		false, // auto delete
		false, // exclusive
		false, // no wait
		nil,   // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %v", err)
	}

	// Bind the queue to the exchange
	err = ch.QueueBind(
		q.Name,    // queue name
		"go-auth", // routing key (ignored for fanout)
		"orders",  // exchange
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to bind the queue to the exchange: %v", err)
	}

	msgs, err := ch.Consume(
		q.Name,
		"",    // consumer
		true,  // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		log.Fatalf("Failed to register a consumer: %v", err)
	}

	log.Printf(" [*] Waiting for messages in %s. To exit press CTRL+C", q.Name)
	for msg := range msgs {
		log.Printf("Received a message: [%s] %s", msg.RoutingKey, msg.Body)
		// Process the message here (e.g., log, store in DB, etc.)
	}
}
