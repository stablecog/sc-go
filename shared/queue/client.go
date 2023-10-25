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
	conn    *amqp.Connection
	Ctx     context.Context
	Channel *amqp.Channel
	Queue   string
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
		conn:    conn,
		Ctx:     context,
		Channel: channel,
		Queue:   utils.GetEnv().RabbitMQQueueName,
	}
	err = client.createQueue()
	if err != nil {
		log.Errorf("Error creating RabbitMQ queue %v", err)
		return nil, err
	}
	return client, nil
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
	c.conn.Close()
	c.Channel.Close()
}

func (c *RabbitMQClient) Publish(id string, msg any, priority uint8) error {
	marshalled, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return c.Channel.PublishWithContext(
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
}
