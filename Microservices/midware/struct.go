package midware

import (
	"encoding/json"
	"sync"

	"github.com/madmaxeatfax/homeworks/Microservices/proto"
)

type middleware struct {
	sync.RWMutex
	acl map[string][]string // map{consumer: [method1, method2...]}

	logChan    chan *proto.Event
	logWaiters []chan *proto.Event
}

func New(ACLdata string) (*middleware, error) {
	m := middleware{}

	m.acl = make(map[string][]string, 0)
	err := json.Unmarshal([]byte(ACLdata), &m.acl)
	if err != nil {
		return nil, err
	}

	m.logChan = make(chan *proto.Event) // !!!buffer!!!
	m.logWaiters = make([]chan *proto.Event, 0)

	return &m, nil
}
