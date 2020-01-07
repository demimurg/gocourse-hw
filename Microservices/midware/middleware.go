package midware

type middleware struct {
	acl aclmod
	log logmod
}

func New(ACLdata string) (*middleware, error) {
	m := middleware{}

	if err := m.acl.scan(ACLdata); err != nil {
		return nil, err
	}

	return &m, nil
}

func (m *middleware) CreateLogger() Logger {
	m.log.tunnels = make(tunnels, 0)

	return &m.log
}

func (m *middleware) CloseLogger() {
	for waiter := range m.log.tunnels {
		close(waiter)
	}
	m.log.tunnels = make(tunnels, 0)
}
