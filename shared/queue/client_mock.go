package queue

type MockRabbitMQClient struct {
	PublishFunc func(id string, msg any, priority uint8) error
}

func (m *MockRabbitMQClient) Close() {}

func (m *MockRabbitMQClient) Publish(id string, msg any, priority uint8) error {
	if m.PublishFunc != nil {
		return m.PublishFunc(id, msg, priority)
	}
	return nil
}
