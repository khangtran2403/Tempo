package connector

import (
	"context"
	"fmt"
)

type Connector interface {
	//ten cua Connector
	Name() string
	//execute voi config cho truoc
	Execute(ctx context.Context, cfg map[string]interface{}, prevRes map[string]interface{}) (interface{}, error)
	//kiem tra config co hop le
	ValidateConfig(cfg map[string]interface{}) error
}

type Registry struct {
	connectors map[string]Connector
}

func NewRegistry() *Registry {
	return &Registry{
		connectors: make(map[string]Connector),
	}
}

func (r *Registry) Register(connector Connector) {
	r.connectors[connector.Name()] = connector
}

// GetConnector lấy connector theo tên
func (r *Registry) GetConnector(name string) (Connector, error) {
	connector, ok := r.connectors[name]
	if !ok {
		return nil, fmt.Errorf("connector '%s' khong tim thay", name)
	}
	return connector, nil
}
func (r *Registry) Count() int {
	return len(r.connectors)
}

func (r *Registry) ListConnectors() []string {
	names := make([]string, 0, len(r.connectors))
	for name := range r.connectors {
		names = append(names, name)
	}
	return names
}
