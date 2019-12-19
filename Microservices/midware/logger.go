package midware

import (
	"github.com/madmaxeatfax/homeworks/Microservices/proto"
)

func (m *middleware) StartLogger() {
	for event := range m.logChan {
		m.RLock()
		for _, waiter := range m.logWaiters {
			waiter <- event
		}
		m.RUnlock()
	}

	for _, waiter := range m.logWaiters {
		close(waiter)
	}
}

func (m *middleware) StopLogger() {
	close(m.logChan)
}

func (m *middleware) LogWaitersList() *[]chan *proto.Event {
	return &m.logWaiters
}
