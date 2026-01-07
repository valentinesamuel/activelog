package container

import (
	"fmt"
	"sync"
)

// Container is a simple dependency injection container
// Provides thread-safe singleton management with factory-based registration
type Container struct {
	services  map[string]interface{} // Instantiated singletons
	factories map[string]Factory     // Factory functions for lazy instantiation
	mu        sync.RWMutex           // Thread-safe access
}

// Factory is a function that creates a service instance
// The factory receives the container to resolve dependencies
type Factory func(c *Container) (interface{}, error)

// New creates a new empty container
func New() *Container {
	return &Container{
		services:  make(map[string]interface{}),
		factories: make(map[string]Factory),
	}
}

// Register registers a factory function for a service
// The service will be created lazily when first resolved
// All subsequent resolutions return the same instance (singleton pattern)
func (c *Container) Register(name string, factory Factory) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.factories[name] = factory
}

// RegisterSingleton registers an already-instantiated singleton
// Useful for registering primitive types or pre-configured instances
func (c *Container) RegisterSingleton(name string, instance interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.services[name] = instance
}

// Resolve resolves a service by name
// If the service has already been instantiated, returns the cached instance
// If not, calls the factory to create it, caches it, and returns it
// Returns an error if the service is not registered or if the factory fails
func (c *Container) Resolve(name string) (interface{}, error) {
	// Check if already instantiated
	c.mu.RLock()
	if service, exists := c.services[name]; exists {
		c.mu.RUnlock()
		return service, nil
	}
	c.mu.RUnlock()

	// Get factory
	c.mu.RLock()
	factory, exists := c.factories[name]
	c.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("service not registered: %s", name)
	}

	// Create instance (outside of lock to prevent deadlock if factory resolves other services)
	instance, err := factory(c)
	if err != nil {
		return nil, fmt.Errorf("failed to create service %s: %w", name, err)
	}

	// Store as singleton
	c.mu.Lock()
	// Double-check in case another goroutine created it while we were creating
	if existingService, exists := c.services[name]; exists {
		c.mu.Unlock()
		return existingService, nil
	}
	c.services[name] = instance
	c.mu.Unlock()

	return instance, nil
}

// MustResolve resolves a service or panics if resolution fails
// Useful for application initialization where missing dependencies should be fatal
// Use sparingly - prefer Resolve() for error handling in most cases
func (c *Container) MustResolve(name string) interface{} {
	service, err := c.Resolve(name)
	if err != nil {
		panic(fmt.Sprintf("container: failed to resolve %s: %v", name, err))
	}
	return service
}

// Has checks if a service is registered (either as factory or singleton)
func (c *Container) Has(name string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	_, inServices := c.services[name]
	_, inFactories := c.factories[name]

	return inServices || inFactories
}

// Clear removes all registered services and factories
// Useful for testing or resetting the container
func (c *Container) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.services = make(map[string]interface{})
	c.factories = make(map[string]Factory)
}

// List returns the names of all registered services (both factories and singletons)
// Useful for debugging and introspection
func (c *Container) List() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	names := make(map[string]bool)

	for name := range c.services {
		names[name] = true
	}

	for name := range c.factories {
		names[name] = true
	}

	result := make([]string, 0, len(names))
	for name := range names {
		result = append(result, name)
	}

	return result
}
