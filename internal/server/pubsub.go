package server

import (
	"sync"
)

type PubSub struct {
	topics map[string][]chan string
	Mutex  sync.Mutex
}

func NewPubSub() *PubSub {
	return &PubSub{
		topics: make(map[string][]chan string),
	}
}

func (ps *PubSub) Subscribe(topic string) chan string {
	ps.Mutex.Lock()
	defer ps.Mutex.Unlock()

	if _, ok := ps.topics[topic]; !ok {
		ps.topics[topic] = make([]chan string, 0)
	}

	ch := make(chan string)
	ps.topics[topic] = append(ps.topics[topic], ch)
	return ch
}

func (ps *PubSub) Publish(topic, message string) {
	ps.Mutex.Lock()
	defer ps.Mutex.Unlock()

	if subscribers, ok := ps.topics[topic]; ok {
		for _, ch := range subscribers {
			go func(ch chan string) {
				ch <- message
			}(ch)
		}
	}
}
