package broker

import (
	"context"
	"sync"
	"therealbroker/pkg/broker"
	"time"
)

type Module struct {
	topics map[string] *Topic
	Closed bool
	lock sync.Mutex
	db DB
}

func NewModule() broker.Broker {
	return &Module{
		Closed: false,
		topics: make(map[string] *Topic),
		db: NewPsql(),
	}
}
func (m *Module) getOrCreateTopic(name string) (*Topic, bool){
	m.lock.Lock()
	defer m.lock.Unlock()
	if _, ok := m.topics[name]; !ok{
		m.topics[name] = NewTopic(name, m.db)
		return m.topics[name], false
	}
	return m.topics[name], true
}
func (m *Module) Close() error {
	if m.Closed{
		return broker.ErrUnavailable
	}
	m.Closed = true
	return nil
}

func (m *Module) Publish(ctx context.Context, subject string, msg broker.Message) (int, error) {
	if m.Closed{
		return -1, broker.ErrUnavailable
	}
	topic, _ := m.getOrCreateTopic(subject)
	creatTime := time.Now()
	msgID := topic.Publish(msg, creatTime)
	return msgID, nil
}

func (m *Module) Subscribe(ctx context.Context, subject string) (<-chan broker.Message, error) {
	if m.Closed{
	return nil, broker.ErrUnavailable
	}
	topic, _ := m.getOrCreateTopic(subject)
	out := make(chan broker.Message)
	topic.AddSubscriber(ctx, out)
	return out, nil
}

func (m *Module) Fetch(ctx context.Context, subject string, id int) (broker.Message, error) {
	if m.Closed{
		return broker.Message{}, broker.ErrUnavailable
	}
	topic, exist := m.getOrCreateTopic(subject)
	if !exist {
		return broker.Message{}, broker.ErrInvalidID
	}
	return topic.FetchMessage(id)
}