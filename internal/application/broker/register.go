package broker

import (
	"database/sql"

	"github.com/valentinesamuel/activelog/internal/container"
)

// CoreRawDBKey is the key for the raw database connection singleton
const CoreRawDBKey = "rawDB"

// RegisterBroker registers the use case orchestrator with the container
// Dependencies: Requires "rawDB" to be registered first
func RegisterBroker(c *container.Container) {
	c.Register(BrokerKey, func(c *container.Container) (interface{}, error) {
		rawDB := c.MustResolve(CoreRawDBKey).(*sql.DB)
		return NewBroker(rawDB), nil
	})
}
