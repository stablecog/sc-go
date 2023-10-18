package queue

import (
	"context"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/utils"
)

type MQClient interface {
	Close()
	Publish(routingKey string, id string, msg any, priority uint8) error
}

type RabbitMQClient struct {
	conn     *amqp.Connection
	Ctx      context.Context
	Channel  *amqp.Channel
	Exchange string
}

// Create new RabbitMQ amqp client, using amqp:// connection string
func NewRabbitMQClient(context context.Context, connUrl string) (*RabbitMQClient, error) {
	conn, err := amqp.Dial(connUrl)
	if err != nil {
		log.Errorf("Error connecting to RabbitMQ %v", err)
		return nil, err
	}
	channel, err := conn.Channel()
	if err != nil {
		log.Errorf("Error creating RabbitMQ channel %v", err)
		return nil, err
	}

	client := &RabbitMQClient{
		conn:     conn,
		Ctx:      context,
		Channel:  channel,
		Exchange: utils.GetEnv("RABBITMQ_EXCHANGE_NAME", "stablecog-dev-exchange"),
	}
	err = client.createExchange()
	if err != nil {
		log.Errorf("Error creating RabbitMQ exchange %v", err)
		return nil, err
	}
	return client, nil
}

func (c *RabbitMQClient) createExchange() error {
	return c.Channel.ExchangeDeclare(
		c.Exchange, // name
		"direct",   // type
		true,       // durable
		false,      // auto-deleted
		false,      // internal
		false,      // no-wait
		nil,        // arguments
	)
}

func (c *RabbitMQClient) Close() {
	c.conn.Close()
	c.Channel.Close()
}

func (c *RabbitMQClient) Publish(routingKey string, id string, msg any, priority uint8) error {
	marshalled, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return c.Channel.PublishWithContext(
		c.Ctx,
		c.Exchange, // exchange
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			MessageId:    id,
			DeliveryMode: amqp.Persistent,
			Priority:     priority,
			ContentType:  "text/plain",
			Body:         marshalled,
		})
}
