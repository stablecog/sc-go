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
	Publish(id string, msg any, priority uint8) error
}

type RabbitMQClient struct {
	connUrl string
	conn    *amqp.Connection
	Ctx     context.Context
	Channel *amqp.Channel
	Queue   string
}

// Create new RabbitMQ amqp client, using amqp:// connection string
func NewRabbitMQClient(ctx context.Context, connUrl string) (*RabbitMQClient, error) {
	client := &RabbitMQClient{
		connUrl: connUrl,
		Ctx:     ctx,
		Queue:   utils.GetEnv().RabbitMQQueueName,
	}
	err := client.connect()
	if err != nil {
		return nil, err
	}
	return client, nil
}

func (c *RabbitMQClient) connect() error {
	var err error
	c.conn, err = amqp.Dial(c.connUrl)
	if err != nil {
		return err
	}

	c.Channel, err = c.conn.Channel()
	if err != nil {
		return err
	}

	return c.createQueue()
}

func (c *RabbitMQClient) createQueue() error {
	_, err := c.Channel.QueueDeclare(
		c.Queue, // name
		true,    //durable
		false,   // auto_delete,
		false,   // exclusie,
		false,   // noWait
		amqp.Table{
			"x-max-priority": 10,
			"x-message-ttl":  1800000,
		},
	)
	return err
}

func (c *RabbitMQClient) Close() {
	if c.Channel != nil {
		c.Channel.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
}

// Reconnect reconnects the connection
func (c *RabbitMQClient) Reconnect() error {
	if err := c.connect(); err != nil {
		return err
	}
	return nil
}

func (c *RabbitMQClient) Publish(id string, msg any, priority uint8) error {
	marshalled, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	err = c.Channel.PublishWithContext(
		c.Ctx,
		"",      // exchange
		c.Queue, // routing key
		false,   // mandatory
		false,   // immediate
		amqp.Publishing{
			MessageId:    id,
			DeliveryMode: amqp.Persistent,
			Priority:     priority,
			ContentType:  "text/plain",
			Body:         marshalled,
		})
	// Handle reconnect if conn closed
	if err != nil {
		if c.conn.IsClosed() {
			log.Info("Connection to RabbitMQ lost. Attempting to reconnect...")
			reconnectErr := c.Reconnect()
			if reconnectErr != nil {
				log.Errorf("Failed to reconnect to RabbitMQ: %v", reconnectErr)
				return reconnectErr
			}
			// Re-publish
			return c.Publish(id, msg, priority)
		}
		return err
	}
	return nil
}
