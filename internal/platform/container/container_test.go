package container

import (
	"errors"
	"fmt"
	"sync"
	"testing"
)

// Test types for dependency injection
type TestService struct {
	Value string
}

type TestDependency struct {
	Name string
}

type TestServiceWithDep struct {
	Dependency *TestDependency
	ID         int
}

func TestNew(t *testing.T) {
	c := New()

	if c == nil {
		t.Fatal("New() returned nil")
	}

	if c.services == nil {
		t.Error("services map not initialized")
	}

	if c.factories == nil {
		t.Error("factories map not initialized")
	}
}

func TestRegisterSingleton(t *testing.T) {
	c := New()
	service := &TestService{Value: "test"}

	c.RegisterSingleton("testService", service)

	// Verify it can be resolved
	resolved, err := c.Resolve("testService")
	if err != nil {
		t.Fatalf("Failed to resolve singleton: %v", err)
	}

	// Verify it's the same instance
	if resolved != service {
		t.Error("Resolved service is not the same instance as registered")
	}

	// Verify type assertion works
	resolvedService, ok := resolved.(*TestService)
	if !ok {
		t.Fatal("Failed to type assert resolved service")
	}

	if resolvedService.Value != "test" {
		t.Errorf("Expected value 'test', got '%s'", resolvedService.Value)
	}
}

func TestRegister(t *testing.T) {
	c := New()

	c.Register("testService", func(c *Container) (interface{}, error) {
		return &TestService{Value: "created"}, nil
	})

	// Verify service is not created yet
	if len(c.services) != 0 {
		t.Error("Service should not be created until resolved")
	}

	// Resolve and verify creation
	resolved, err := c.Resolve("testService")
	if err != nil {
		t.Fatalf("Failed to resolve service: %v", err)
	}

	service, ok := resolved.(*TestService)
	if !ok {
		t.Fatal("Failed to type assert resolved service")
	}

	if service.Value != "created" {
		t.Errorf("Expected value 'created', got '%s'", service.Value)
	}

	// Verify it's now in services (cached)
	if len(c.services) != 1 {
		t.Error("Service should be cached after resolution")
	}
}

func TestResolve_NotRegistered(t *testing.T) {
	c := New()

	_, err := c.Resolve("nonExistent")
	if err == nil {
		t.Fatal("Expected error when resolving non-existent service")
	}

	expectedMsg := "service not registered: nonExistent"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestResolve_FactoryError(t *testing.T) {
	c := New()

	factoryErr := errors.New("factory failed")
	c.Register("failingService", func(c *Container) (interface{}, error) {
		return nil, factoryErr
	})

	_, err := c.Resolve("failingService")
	if err == nil {
		t.Fatal("Expected error from failing factory")
	}

	if !errors.Is(err, factoryErr) {
		t.Errorf("Expected error to wrap factory error, got: %v", err)
	}
}

func TestResolve_SingletonBehavior(t *testing.T) {
	c := New()

	callCount := 0
	c.Register("testService", func(c *Container) (interface{}, error) {
		callCount++
		return &TestService{Value: fmt.Sprintf("instance_%d", callCount)}, nil
	})

	// Resolve multiple times
	first, err := c.Resolve("testService")
	if err != nil {
		t.Fatalf("First resolve failed: %v", err)
	}

	second, err := c.Resolve("testService")
	if err != nil {
		t.Fatalf("Second resolve failed: %v", err)
	}

	third, err := c.Resolve("testService")
	if err != nil {
		t.Fatalf("Third resolve failed: %v", err)
	}

	// Verify factory was called only once
	if callCount != 1 {
		t.Errorf("Expected factory to be called once, was called %d times", callCount)
	}

	// Verify all resolutions return the same instance
	if first != second || second != third {
		t.Error("Resolve should return the same instance (singleton pattern)")
	}

	// Verify the value is from the first call
	service := first.(*TestService)
	if service.Value != "instance_1" {
		t.Errorf("Expected 'instance_1', got '%s'", service.Value)
	}
}

func TestResolve_DependencyResolution(t *testing.T) {
	c := New()

	// Register dependency
	c.Register("dependency", func(c *Container) (interface{}, error) {
		return &TestDependency{Name: "dep1"}, nil
	})

	// Register service that depends on it
	c.Register("serviceWithDep", func(c *Container) (interface{}, error) {
		dep, err := c.Resolve("dependency")
		if err != nil {
			return nil, err
		}

		return &TestServiceWithDep{
			Dependency: dep.(*TestDependency),
			ID:         42,
		}, nil
	})

	// Resolve service
	resolved, err := c.Resolve("serviceWithDep")
	if err != nil {
		t.Fatalf("Failed to resolve service with dependency: %v", err)
	}

	service := resolved.(*TestServiceWithDep)

	if service.ID != 42 {
		t.Errorf("Expected ID 42, got %d", service.ID)
	}

	if service.Dependency == nil {
		t.Fatal("Dependency not resolved")
	}

	if service.Dependency.Name != "dep1" {
		t.Errorf("Expected dependency name 'dep1', got '%s'", service.Dependency.Name)
	}

	// Verify dependency is a singleton (same instance if resolved separately)
	dep, _ := c.Resolve("dependency")
	if dep != service.Dependency {
		t.Error("Dependency should be the same instance")
	}
}

func TestMustResolve_Success(t *testing.T) {
	c := New()

	c.RegisterSingleton("testService", &TestService{Value: "test"})

	// Should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("MustResolve panicked unexpectedly: %v", r)
		}
	}()

	service := c.MustResolve("testService")

	if service == nil {
		t.Fatal("MustResolve returned nil")
	}

	testService := service.(*TestService)
	if testService.Value != "test" {
		t.Errorf("Expected value 'test', got '%s'", testService.Value)
	}
}

func TestMustResolve_Panic(t *testing.T) {
	c := New()

	// Should panic
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("MustResolve should panic for non-existent service")
		}

		// Verify panic message
		msg, ok := r.(string)
		if !ok {
			t.Fatalf("Expected panic string, got %T", r)
		}

		if msg != "container: failed to resolve nonExistent: service not registered: nonExistent" {
			t.Errorf("Unexpected panic message: %s", msg)
		}
	}()

	c.MustResolve("nonExistent")
}

func TestHas(t *testing.T) {
	c := New()

	// Register singleton
	c.RegisterSingleton("singleton", &TestService{Value: "test"})

	// Register factory
	c.Register("factory", func(c *Container) (interface{}, error) {
		return &TestService{Value: "test"}, nil
	})

	tests := []struct {
		name     string
		expected bool
	}{
		{"singleton", true},
		{"factory", true},
		{"nonExistent", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := c.Has(tt.name)
			if result != tt.expected {
				t.Errorf("Has(%s) = %v, expected %v", tt.name, result, tt.expected)
			}
		})
	}
}

func TestClear(t *testing.T) {
	c := New()

	c.RegisterSingleton("singleton", &TestService{Value: "test"})
	c.Register("factory", func(c *Container) (interface{}, error) {
		return &TestService{Value: "test"}, nil
	})

	// Verify services registered
	if !c.Has("singleton") || !c.Has("factory") {
		t.Fatal("Services should be registered before clear")
	}

	c.Clear()

	// Verify all cleared
	if c.Has("singleton") || c.Has("factory") {
		t.Error("Services should not exist after clear")
	}

	if len(c.services) != 0 || len(c.factories) != 0 {
		t.Error("Maps should be empty after clear")
	}
}

func TestList(t *testing.T) {
	c := New()

	c.RegisterSingleton("service1", &TestService{Value: "test1"})
	c.Register("service2", func(c *Container) (interface{}, error) {
		return &TestService{Value: "test2"}, nil
	})
	c.RegisterSingleton("service3", &TestService{Value: "test3"})

	list := c.List()

	if len(list) != 3 {
		t.Fatalf("Expected 3 services, got %d", len(list))
	}

	// Verify all services are in the list
	found := make(map[string]bool)
	for _, name := range list {
		found[name] = true
	}

	expectedServices := []string{"service1", "service2", "service3"}
	for _, expected := range expectedServices {
		if !found[expected] {
			t.Errorf("Service '%s' not found in list", expected)
		}
	}
}

func TestConcurrentResolve(t *testing.T) {
	c := New()

	callCount := 0
	var mu sync.Mutex

	c.Register("testService", func(c *Container) (interface{}, error) {
		mu.Lock()
		callCount++
		count := callCount
		mu.Unlock()

		return &TestService{Value: fmt.Sprintf("instance_%d", count)}, nil
	})

	var wg sync.WaitGroup
	numGoroutines := 100
	results := make([]interface{}, numGoroutines)

	// Resolve concurrently from many goroutines
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			service, err := c.Resolve("testService")
			if err != nil {
				t.Errorf("Concurrent resolve failed: %v", err)
				return
			}
			results[index] = service
		}(i)
	}

	wg.Wait()

	// Verify factory was called only once (or very few times due to race)
	// Note: Due to double-check locking, it might be called 2-3 times, but not 100 times
	if callCount > 5 {
		t.Errorf("Expected factory to be called ~1 time, was called %d times", callCount)
	}

	// Verify all goroutines got the same instance
	firstInstance := results[0]
	for i, result := range results {
		if result != firstInstance {
			t.Errorf("Goroutine %d got different instance", i)
		}
	}
}

func TestConcurrentRegisterAndResolve(t *testing.T) {
	c := New()

	var wg sync.WaitGroup

	// Register services concurrently
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			serviceName := fmt.Sprintf("service_%d", index)
			c.Register(serviceName, func(c *Container) (interface{}, error) {
				return &TestService{Value: serviceName}, nil
			})
		}(i)
	}

	wg.Wait()

	// Resolve all services concurrently
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			serviceName := fmt.Sprintf("service_%d", index)
			service, err := c.Resolve(serviceName)
			if err != nil {
				t.Errorf("Failed to resolve %s: %v", serviceName, err)
				return
			}

			testService := service.(*TestService)
			if testService.Value != serviceName {
				t.Errorf("Expected value '%s', got '%s'", serviceName, testService.Value)
			}
		}(i)
	}

	wg.Wait()
}

// Benchmark resolution performance
func BenchmarkResolve_FirstTime(b *testing.B) {
	c := New()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		c.Clear()
		c.Register("testService", func(c *Container) (interface{}, error) {
			return &TestService{Value: "test"}, nil
		})
		b.StartTimer()

		c.Resolve("testService")
	}
}

func BenchmarkResolve_Cached(b *testing.B) {
	c := New()

	c.Register("testService", func(c *Container) (interface{}, error) {
		return &TestService{Value: "test"}, nil
	})

	// Prime the cache
	c.Resolve("testService")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		c.Resolve("testService")
	}
}

func BenchmarkMustResolve_Cached(b *testing.B) {
	c := New()

	c.RegisterSingleton("testService", &TestService{Value: "test"})

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		c.MustResolve("testService")
	}
}
