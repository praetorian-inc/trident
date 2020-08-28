package nozzle

import (
	"fmt"
	"sync"

	"git.praetorianlabs.com/mars/trident/functions/events"
)

var (
	driversMu sync.RWMutex
	drivers   = make(map[string]Driver)
)

type Driver interface {
	New(opts map[string]string) (Nozzle, error)
}

type Nozzle interface {
	Login(username, password string) (*events.AuthResponse, error)
}

func Open(name string, opts map[string]string) (Nozzle, error) {
	driversMu.RLock()
	n, ok := drivers[name]
	driversMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("nozzle: unknown driver %q (forgotten import?)", name)
	}

	return n.New(opts)
}

func Register(name string, driver Driver) {
	driversMu.Lock()
	defer driversMu.Unlock()
	if driver == nil {
		panic("nozzle: Register driver is nil")
	}
	if _, dup := drivers[name]; dup {
		panic("nozzle: Register called twice for driver " + name)
	}
	drivers[name] = driver
}
