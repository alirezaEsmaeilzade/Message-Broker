package broker

import (
	"context"
	"sync"
	"therealbroker/pkg/broker"
)

type Subscriber struct {
	QueueMessage  []*broker.Message
	worked        bool
	C             chan broker.Message
	channel       chan bool
	lock          sync.Mutex
	ctx context.Context
}
func NewSubscriber(ctx context.Context, out chan broker.Message) *Subscriber{
	return &Subscriber{
		QueueMessage: make([]*broker.Message, 0),
		C: out,
		worked: false,
		channel: make(chan bool),
		ctx: ctx,
	}
}
func (sub *Subscriber) AddMessageToQueue(message broker.Message) {
	sub.lock.Lock()
	defer sub.lock.Unlock()
	sub.QueueMessage = append(sub.QueueMessage, &message)
	if !sub.worked{
		go func() {
			sub.channel <- true
		}()
		sub.worked = true
	}
}

func (sub *Subscriber) Listen() {
	for {
		select {
		case <-sub.ctx.Done():
			return

		case <-sub.channel:
			sub.lock.Lock()
			sub.worked = false
			queue := make([]*broker.Message, len(sub.QueueMessage))
			copy(queue, sub.QueueMessage)
			sub.QueueMessage = sub.QueueMessage[:0]
			sub.lock.Unlock()
			for _, msg := range queue{
				sub.C <- *msg
			}
		}
	}
}
