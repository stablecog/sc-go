package queue

type MockRabbitMQClient struct {
	PublishFunc func(routingKey string, msg any, priority uint8) error
}

func (m *MockRabbitMQClient) Close() {}

func (m *MockRabbitMQClient) Publish(routingKey string, msg any, priority uint8) error {
	if m.PublishFunc != nil {
		return m.PublishFunc(routingKey, msg, priority)
	}
	return nil
}
