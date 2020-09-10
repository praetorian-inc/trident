package nozzle

import (
	"fmt"
	"sync"

	"github.com/praetorian-inc/trident/pkg/event"
)

var (
	driversMu sync.RWMutex
	drivers   = make(map[string]Driver)
)

// Driver is the interface the wraps creation of a Nozzle.
type Driver interface {
	New(opts map[string]string) (Nozzle, error)
}

// Nozzle is the interface that wraps a basic Login() method to be implemented for
// each authentication provider we support.
type Nozzle interface {
	Login(username, password string) (*event.AuthResponse, error)
}

// Open opens a nozzle specified by the nozzle driver name (e.g. okta) and
// configures that nozzle via the provided opts argument. Each Nozzle should
// document its configuration options in its New() method.
func Open(name string, opts map[string]string) (Nozzle, error) {
	driversMu.RLock()
	n, ok := drivers[name]
	driversMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("nozzle: unknown driver %q (forgotten import?)", name)
	}

	return n.New(opts)
}

// Register makes a nozzle driver available at the provided name. If register is
// called twice or if the driver is nil, if panics. Register() is typically
// called in the nozzle implementation's init() function to allow for easy
// importing of each nozzle.
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
