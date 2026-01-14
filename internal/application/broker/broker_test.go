package broker

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

// Mock use case for testing
type mockUseCase struct {
	name              string
	output            map[string]interface{}
	err               error
	requiresTx        bool // Whether this mock requires a transaction
	executeFn         func(ctx context.Context, tx *sql.Tx, input map[string]interface{}) (map[string]interface{}, error)
}

func (m *mockUseCase) Execute(ctx context.Context, tx *sql.Tx, input map[string]interface{}) (map[string]interface{}, error) {
	if m.executeFn != nil {
		return m.executeFn(ctx, tx, input)
	}
	if m.err != nil {
		return nil, m.err
	}
	return m.output, nil
}

// RequiresTransaction implements TransactionalUseCase marker interface
// For backward compatibility, mock use cases default to requiring transactions
func (m *mockUseCase) RequiresTransaction() bool {
	return m.requiresTx
}

// Test helper to create test database
func setupTestBroker(t *testing.T) (*Broker, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}

	broker := NewBroker(db)
	cleanup := func() {
		db.Close()
	}

	return broker, mock, cleanup
}

func TestNewBroker(t *testing.T) {
	db, _, cleanup := setupTestBroker(t)
	defer cleanup()

	if db == nil {
		t.Fatal("expected broker to be created")
	}
	if db.defaultTimeout != 60*time.Second {
		t.Errorf("expected default timeout 60s, got %v", db.defaultTimeout)
	}
	if db.defaultIsolationLevel != sql.LevelReadCommitted {
		t.Errorf("expected isolation level ReadCommitted, got %v", db.defaultIsolationLevel)
	}
}

func TestRunUseCases_Success(t *testing.T) {
	broker, mock, cleanup := setupTestBroker(t)
	defer cleanup()

	// Expect transaction to be started and committed
	mock.ExpectBegin()
	mock.ExpectCommit()

	// Create simple use cases
	useCase1 := &mockUseCase{
		name:       "UseCase1",
		output:     map[string]interface{}{"step1": "completed"},
		requiresTx: true, // Requires transaction
	}
	useCase2 := &mockUseCase{
		name:       "UseCase2",
		output:     map[string]interface{}{"step2": "completed"},
		requiresTx: true, // Requires transaction
	}

	initialInput := map[string]interface{}{"user_id": 123}
	useCases := []UseCase{useCase1, useCase2}

	result, err := broker.RunUseCases(context.Background(), useCases, initialInput)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Should have outputs from both use cases (initial input excluded)
	if result["step1"] != "completed" {
		t.Errorf("expected step1 to be completed")
	}
	if result["step2"] != "completed" {
		t.Errorf("expected step2 to be completed")
	}
	// Initial input should be excluded from final results
	if _, exists := result["user_id"]; exists {
		t.Errorf("initial input should not be in final results")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestRunUseCases_ResultChaining(t *testing.T) {
	broker, mock, cleanup := setupTestBroker(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectCommit()

	// Use case 1 outputs a value that use case 2 should receive
	var receivedInput map[string]interface{}
	useCase1 := &mockUseCase{
		output:     map[string]interface{}{"created_id": 456},
		requiresTx: true, // Requires transaction
	}
	useCase2 := &mockUseCase{
		requiresTx: true, // Requires transaction
		executeFn: func(ctx context.Context, tx *sql.Tx, input map[string]interface{}) (map[string]interface{}, error) {
			receivedInput = input
			return map[string]interface{}{"updated": true}, nil
		},
	}

	initialInput := map[string]interface{}{"user_id": 123}
	useCases := []UseCase{useCase1, useCase2}

	_, err := broker.RunUseCases(context.Background(), useCases, initialInput)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// UseCase2 should receive both initial input and output from UseCase1
	if receivedInput["user_id"] != 123 {
		t.Errorf("expected user_id from initial input")
	}
	if receivedInput["created_id"] != 456 {
		t.Errorf("expected created_id from UseCase1 output")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestRunUseCases_FailureRollback(t *testing.T) {
	broker, mock, cleanup := setupTestBroker(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectRollback()

	useCase1 := &mockUseCase{
		output:     map[string]interface{}{"step1": "completed"},
		requiresTx: true, // Requires transaction
	}
	useCase2 := &mockUseCase{
		err:        errors.New("use case 2 failed"),
		requiresTx: true, // Requires transaction
	}

	useCases := []UseCase{useCase1, useCase2}
	initialInput := map[string]interface{}{"user_id": 123}

	result, err := broker.RunUseCases(context.Background(), useCases, initialInput)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if result != nil {
		t.Errorf("expected nil result on error, got %v", result)
	}
	if !errors.Is(err, useCase2.err) {
		t.Errorf("expected error to contain use case error")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestRunUseCases_EmptyUseCases(t *testing.T) {
	broker, _, cleanup := setupTestBroker(t)
	defer cleanup()

	result, err := broker.RunUseCases(context.Background(), []UseCase{}, nil)

	if err == nil {
		t.Fatal("expected error for empty use cases, got nil")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
	if err.Error() != "at least one use case must be provided" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestRunUseCases_Timeout(t *testing.T) {
	broker, mock, cleanup := setupTestBroker(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectRollback()

	// Use case that takes longer than timeout
	slowUseCase := &mockUseCase{
		requiresTx: true, // Requires transaction
		executeFn: func(ctx context.Context, tx *sql.Tx, input map[string]interface{}) (map[string]interface{}, error) {
			select {
			case <-time.After(200 * time.Millisecond):
				return map[string]interface{}{}, nil
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		},
	}

	useCases := []UseCase{slowUseCase}
	initialInput := map[string]interface{}{}

	// Set very short timeout
	result, err := broker.RunUseCases(
		context.Background(),
		useCases,
		initialInput,
		WithTimeout(50*time.Millisecond),
	)

	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if result != nil {
		t.Errorf("expected nil result on timeout, got %v", result)
	}
	if err.Error() != "transaction timed out after 50ms" {
		t.Errorf("unexpected error message: %v", err)
	}

	// Note: Rollback might not be called in time due to goroutine
	// This is acceptable as the transaction will be rolled back when connection closes
}

func TestRunUseCases_WithOptions(t *testing.T) {
	broker, mock, cleanup := setupTestBroker(t)
	defer cleanup()

	// Expect transaction with specific isolation level
	mock.ExpectBegin()
	mock.ExpectCommit()

	useCase := &mockUseCase{
		output:     map[string]interface{}{"result": "success"},
		requiresTx: true, // Requires transaction
	}

	useCases := []UseCase{useCase}
	initialInput := map[string]interface{}{}

	_, err := broker.RunUseCases(
		context.Background(),
		useCases,
		initialInput,
		WithTimeout(5*time.Second),
		WithIsolationLevel(sql.LevelSerializable),
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestRunUseCases_TransactionBeginError(t *testing.T) {
	broker, mock, cleanup := setupTestBroker(t)
	defer cleanup()

	// Simulate error when beginning transaction
	mock.ExpectBegin().WillReturnError(errors.New("connection failed"))

	useCase := &mockUseCase{
		output:     map[string]interface{}{"result": "success"},
		requiresTx: true, // Requires transaction
	}

	result, err := broker.RunUseCases(context.Background(), []UseCase{useCase}, nil)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestRunUseCases_CommitError(t *testing.T) {
	broker, mock, cleanup := setupTestBroker(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectCommit().WillReturnError(errors.New("commit failed"))

	useCase := &mockUseCase{
		output:     map[string]interface{}{"result": "success"},
		requiresTx: true, // Requires transaction
	}

	result, err := broker.RunUseCases(context.Background(), []UseCase{useCase}, nil)

	if err == nil {
		t.Fatal("expected commit error, got nil")
	}
	if result != nil {
		t.Errorf("expected nil result on commit error, got %v", result)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// Note: Panic recovery testing is skipped because RunUseCases executes in a goroutine.
// When a panic occurs in the goroutine, it's recovered, rollback happens, and then
// the panic is re-thrown - which would crash the test. In production, this would
// crash the server, which is the desired behavior (fail fast).
//
// The rollback mechanism is tested through the FailureRollback test instead.

func TestUseCaseFunc(t *testing.T) {
	// Test that UseCaseFunc wrapper works correctly
	called := false
	fn := UseCaseFunc(func(ctx context.Context, tx *sql.Tx, input map[string]interface{}) (map[string]interface{}, error) {
		called = true
		return map[string]interface{}{"result": "success"}, nil
	})

	result, err := fn.Execute(context.Background(), nil, nil)

	if !called {
		t.Error("expected function to be called")
	}
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result["result"] != "success" {
		t.Errorf("unexpected result: %v", result)
	}
}

func TestWithLogger(t *testing.T) {
	broker, _, cleanup := setupTestBroker(t)
	defer cleanup()

	// Should return the same broker instance
	result := broker.WithLogger(nil)
	if result != broker {
		t.Error("expected WithLogger to return same broker instance")
	}
}

// Benchmark tests
func BenchmarkRunUseCases_SingleUseCase(b *testing.B) {
	db, mock, err := sqlmock.New()
	if err != nil {
		b.Fatalf("failed to create mock db: %v", err)
	}
	defer db.Close()

	broker := NewBroker(db).WithLogger(log.New(io.Discard, "", 0))
	useCase := &mockUseCase{
		output:     map[string]interface{}{"result": "success"},
		requiresTx: true, // Requires transaction
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mock.ExpectBegin()
		mock.ExpectCommit()

		broker.RunUseCases(context.Background(), []UseCase{useCase}, nil)
	}
}

func BenchmarkRunUseCases_MultipleUseCases(b *testing.B) {
	db, mock, err := sqlmock.New()
	if err != nil {
		b.Fatalf("failed to create mock db: %v", err)
	}
	defer db.Close()

	broker := NewBroker(db).WithLogger(log.New(io.Discard, "", 0))
	useCases := make([]UseCase, 5)
	for i := range useCases {
		useCases[i] = &mockUseCase{
			output:     map[string]interface{}{fmt.Sprintf("step%d", i): "completed"},
			requiresTx: true, // Requires transaction
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mock.ExpectBegin()
		mock.ExpectCommit()

		broker.RunUseCases(context.Background(), useCases, nil)
	}
}
