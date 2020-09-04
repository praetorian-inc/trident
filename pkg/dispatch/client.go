package dispatch

import (
	"encoding/json"
	"fmt"
	"sync"
	"trident/pkg/event"
)

var (
	driversMu sync.RWMutex
	drivers   = make(map[string]Driver)
)

type WorkerClient interface {
	Submit(event.AuthRequest) (*event.AuthResponse, error)
}

type Driver interface {
	New(opts map[string]string) (WorkerClient, error)
}

type WorkerOptions map[string]string

func (opts *WorkerOptions) UnmarshalText(text []byte) error {
	return json.Unmarshal(text, opts)
}

func (opts *WorkerOptions) UnmarshalJSON(b []byte) error {
	var s map[string]string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	*opts = WorkerOptions(s)
	return nil
}

func Open(name string, opts WorkerOptions) (WorkerClient, error) {
	driversMu.RLock()
	n, ok := drivers[name]
	driversMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("workerclient: unknown driver %q (forgotten import?)", name)
	}

	return n.New(opts)
}

func Register(name string, driver Driver) {
	driversMu.Lock()
	defer driversMu.Unlock()
	if driver == nil {
		panic("workerclient: Register driver is nil")
	}
	if _, dup := drivers[name]; dup {
		panic("workerclient: Register called twice for driver " + name)
	}
	drivers[name] = driver
}
