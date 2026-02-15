package broker

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"log"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

// Typed mock use case for testing RunUseCase
type mockTypedInput struct {
	UserID int
	Name   string
}

type mockTypedOutput struct {
	Result  string
	Success bool
}

type mockTypedUseCase struct {
	output     mockTypedOutput
	err        error
	requiresTx bool
	executeFn  func(ctx context.Context, tx *sql.Tx, input mockTypedInput) (mockTypedOutput, error)
}

func (m *mockTypedUseCase) Execute(ctx context.Context, tx *sql.Tx, input mockTypedInput) (mockTypedOutput, error) {
	if m.executeFn != nil {
		return m.executeFn(ctx, tx, input)
	}
	if m.err != nil {
		return mockTypedOutput{}, m.err
	}
	return m.output, nil
}

func (m *mockTypedUseCase) RequiresTransaction() bool {
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

func TestWithLogger(t *testing.T) {
	broker, _, cleanup := setupTestBroker(t)
	defer cleanup()

	// Should return the same broker instance
	result := broker.WithLogger(nil)
	if result != broker {
		t.Error("expected WithLogger to return same broker instance")
	}
}

func TestRunUseCase_Success_NonTransactional(t *testing.T) {
	broker, _, cleanup := setupTestBroker(t)
	defer cleanup()

	// No transaction expected for non-transactional use case
	useCase := &mockTypedUseCase{
		output:     mockTypedOutput{Result: "success", Success: true},
		requiresTx: false,
	}

	input := mockTypedInput{UserID: 123, Name: "test"}
	result, err := RunUseCase(broker, context.Background(), useCase, input)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Result != "success" {
		t.Errorf("expected result 'success', got %q", result.Result)
	}
	if !result.Success {
		t.Error("expected Success to be true")
	}
}

func TestRunUseCase_Success_Transactional(t *testing.T) {
	broker, mock, cleanup := setupTestBroker(t)
	defer cleanup()

	// Expect transaction for transactional use case
	mock.ExpectBegin()
	mock.ExpectCommit()

	useCase := &mockTypedUseCase{
		output:     mockTypedOutput{Result: "created", Success: true},
		requiresTx: true,
	}

	input := mockTypedInput{UserID: 456, Name: "create"}
	result, err := RunUseCase(broker, context.Background(), useCase, input)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Result != "created" {
		t.Errorf("expected result 'created', got %q", result.Result)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestRunUseCase_Error_Rollback(t *testing.T) {
	broker, mock, cleanup := setupTestBroker(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectRollback()

	expectedErr := errors.New("use case failed")
	useCase := &mockTypedUseCase{
		err:        expectedErr,
		requiresTx: true,
	}

	input := mockTypedInput{UserID: 789, Name: "fail"}
	_, err := RunUseCase(broker, context.Background(), useCase, input)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected error to wrap use case error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestRunUseCase_Timeout(t *testing.T) {
	broker, mock, cleanup := setupTestBroker(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectRollback()

	// Use case that takes longer than timeout
	useCase := &mockTypedUseCase{
		requiresTx: true,
		executeFn: func(ctx context.Context, tx *sql.Tx, input mockTypedInput) (mockTypedOutput, error) {
			select {
			case <-time.After(200 * time.Millisecond):
				return mockTypedOutput{Result: "completed"}, nil
			case <-ctx.Done():
				return mockTypedOutput{}, ctx.Err()
			}
		},
	}

	input := mockTypedInput{UserID: 123, Name: "slow"}
	_, err := RunUseCase(broker, context.Background(), useCase, input, WithTimeout(50*time.Millisecond))

	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if err.Error() != "use case timed out after 50ms" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestRunUseCase_TransactionBeginError(t *testing.T) {
	broker, mock, cleanup := setupTestBroker(t)
	defer cleanup()

	mock.ExpectBegin().WillReturnError(errors.New("connection failed"))

	useCase := &mockTypedUseCase{
		output:     mockTypedOutput{Result: "success"},
		requiresTx: true,
	}

	input := mockTypedInput{UserID: 123, Name: "test"}
	_, err := RunUseCase(broker, context.Background(), useCase, input)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestRunUseCase_CommitError(t *testing.T) {
	broker, mock, cleanup := setupTestBroker(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectCommit().WillReturnError(errors.New("commit failed"))

	useCase := &mockTypedUseCase{
		output:     mockTypedOutput{Result: "success"},
		requiresTx: true,
	}

	input := mockTypedInput{UserID: 123, Name: "test"}
	_, err := RunUseCase(broker, context.Background(), useCase, input)

	if err == nil {
		t.Fatal("expected commit error, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestRunUseCase_WithOptions(t *testing.T) {
	broker, mock, cleanup := setupTestBroker(t)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectCommit()

	useCase := &mockTypedUseCase{
		output:     mockTypedOutput{Result: "success"},
		requiresTx: true,
	}

	input := mockTypedInput{UserID: 123, Name: "test"}
	_, err := RunUseCase(
		broker,
		context.Background(),
		useCase,
		input,
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

// Benchmark tests
func BenchmarkRunUseCase_NonTransactional(b *testing.B) {
	db, _, err := sqlmock.New()
	if err != nil {
		b.Fatalf("failed to create mock db: %v", err)
	}
	defer db.Close()

	broker := NewBroker(db).WithLogger(log.New(io.Discard, "", 0))
	useCase := &mockTypedUseCase{
		output:     mockTypedOutput{Result: "success"},
		requiresTx: false,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RunUseCase(broker, context.Background(), useCase, mockTypedInput{UserID: 1})
	}
}

func BenchmarkRunUseCase_Transactional(b *testing.B) {
	db, mock, err := sqlmock.New()
	if err != nil {
		b.Fatalf("failed to create mock db: %v", err)
	}
	defer db.Close()

	broker := NewBroker(db).WithLogger(log.New(io.Discard, "", 0))
	useCase := &mockTypedUseCase{
		output:     mockTypedOutput{Result: "success"},
		requiresTx: true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mock.ExpectBegin()
		mock.ExpectCommit()

		RunUseCase(broker, context.Background(), useCase, mockTypedInput{UserID: 1})
	}
}
