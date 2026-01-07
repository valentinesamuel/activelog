package broker

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"log"
	"testing"
)

// Test: All transactional use cases - should create one transaction
func TestTransactionBoundary_AllTransactional(t *testing.T) {
	broker, mock, cleanup := setupTestBroker(t)
	defer cleanup()
	broker.WithLogger(log.New(io.Discard, "", 0)) // Suppress logs for cleaner test output

	// Expect single transaction for all use cases
	mock.ExpectBegin()
	mock.ExpectCommit()

	useCase1 := &mockUseCase{
		name:       "TransactionalUC1",
		output:     map[string]interface{}{"step1": "completed"},
		requiresTx: true,
	}
	useCase2 := &mockUseCase{
		name:       "TransactionalUC2",
		output:     map[string]interface{}{"step2": "completed"},
		requiresTx: true,
	}
	useCase3 := &mockUseCase{
		name:       "TransactionalUC3",
		output:     map[string]interface{}{"step3": "completed"},
		requiresTx: true,
	}

	useCases := []UseCase{useCase1, useCase2, useCase3}
	result, err := broker.RunUseCases(context.Background(), useCases, map[string]interface{}{})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result["step1"] != "completed" || result["step2"] != "completed" || result["step3"] != "completed" {
		t.Error("expected all steps to complete")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// Test: All non-transactional use cases - should not create any transaction
func TestTransactionBoundary_AllNonTransactional(t *testing.T) {
	broker, mock, cleanup := setupTestBroker(t)
	defer cleanup()
	broker.WithLogger(log.New(io.Discard, "", 0))

	// No transaction expected
	// (mock.ExpectBegin() NOT called)

	useCase1 := &mockUseCase{
		name:       "NonTransactionalUC1",
		output:     map[string]interface{}{"step1": "completed"},
		requiresTx: false, // Non-transactional
	}
	useCase2 := &mockUseCase{
		name:       "NonTransactionalUC2",
		output:     map[string]interface{}{"step2": "completed"},
		requiresTx: false, // Non-transactional
	}

	useCases := []UseCase{useCase1, useCase2}
	result, err := broker.RunUseCases(context.Background(), useCases, map[string]interface{}{})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result["step1"] != "completed" || result["step2"] != "completed" {
		t.Error("expected all steps to complete")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// Test: Mixed chain - tx → tx → non-tx → tx
// Should create two separate transactions with boundary break
func TestTransactionBoundary_MixedChain(t *testing.T) {
	broker, mock, cleanup := setupTestBroker(t)
	defer cleanup()
	broker.WithLogger(log.New(io.Discard, "", 0))

	// First transaction (UC1, UC2)
	mock.ExpectBegin()
	mock.ExpectCommit() // Committed when UC3 (non-tx) encountered

	// Second transaction (UC4)
	mock.ExpectBegin()
	mock.ExpectCommit() // Committed at end

	useCase1 := &mockUseCase{
		name:       "TransactionalUC1",
		output:     map[string]interface{}{"step1": "completed"},
		requiresTx: true, // Transaction starts here
	}
	useCase2 := &mockUseCase{
		name:       "TransactionalUC2",
		output:     map[string]interface{}{"step2": "completed"},
		requiresTx: true, // Continues same transaction
	}
	useCase3 := &mockUseCase{
		name:       "NonTransactionalUC3",
		output:     map[string]interface{}{"step3": "completed"},
		requiresTx: false, // BOUNDARY BREAK - commits TX1
	}
	useCase4 := &mockUseCase{
		name:       "TransactionalUC4",
		output:     map[string]interface{}{"step4": "completed"},
		requiresTx: true, // New transaction starts here
	}

	useCases := []UseCase{useCase1, useCase2, useCase3, useCase4}
	result, err := broker.RunUseCases(context.Background(), useCases, map[string]interface{}{})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// All steps should complete
	if result["step1"] != "completed" || result["step2"] != "completed" ||
		result["step3"] != "completed" || result["step4"] != "completed" {
		t.Error("expected all steps to complete")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// Test: Error in non-transactional use case
// Previous transaction should have been committed (not rolled back)
func TestTransactionBoundary_ErrorInNonTransactional(t *testing.T) {
	broker, mock, cleanup := setupTestBroker(t)
	defer cleanup()
	broker.WithLogger(log.New(io.Discard, "", 0))

	// Transaction for UC1 and UC2
	mock.ExpectBegin()
	mock.ExpectCommit() // Should commit when UC3 (non-tx) encountered

	// No transaction for UC3 (it's non-transactional)

	useCase1 := &mockUseCase{
		name:       "TransactionalUC1",
		output:     map[string]interface{}{"step1": "completed"},
		requiresTx: true,
	}
	useCase2 := &mockUseCase{
		name:       "TransactionalUC2",
		output:     map[string]interface{}{"step2": "completed"},
		requiresTx: true,
	}
	useCase3 := &mockUseCase{
		name:       "NonTransactionalUC3",
		err:        errors.New("non-tx use case failed"),
		requiresTx: false, // Non-transactional - should not rollback previous tx
	}

	useCases := []UseCase{useCase1, useCase2, useCase3}
	result, err := broker.RunUseCases(context.Background(), useCases, map[string]interface{}{})

	if err == nil {
		t.Fatal("expected error from UC3, got nil")
	}

	if result != nil {
		t.Errorf("expected nil result on error, got %v", result)
	}

	// IMPORTANT: Transaction should have been committed before UC3 failed
	// This proves the boundary break works correctly
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// Test: Error in transactional use case
// Current transaction should be rolled back
func TestTransactionBoundary_ErrorInTransactional(t *testing.T) {
	broker, mock, cleanup := setupTestBroker(t)
	defer cleanup()
	broker.WithLogger(log.New(io.Discard, "", 0))

	// Transaction should begin and rollback
	mock.ExpectBegin()
	mock.ExpectRollback() // Should rollback when UC2 fails

	useCase1 := &mockUseCase{
		name:       "TransactionalUC1",
		output:     map[string]interface{}{"step1": "completed"},
		requiresTx: true,
	}
	useCase2 := &mockUseCase{
		name:       "TransactionalUC2",
		err:        errors.New("transactional use case failed"),
		requiresTx: true, // Transaction should rollback
	}

	useCases := []UseCase{useCase1, useCase2}
	result, err := broker.RunUseCases(context.Background(), useCases, map[string]interface{}{})

	if err == nil {
		t.Fatal("expected error from UC2, got nil")
	}

	if result != nil {
		t.Errorf("expected nil result on error, got %v", result)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// Test: Complex mixed chain - tx → non-tx → tx → non-tx → tx
func TestTransactionBoundary_ComplexMixedChain(t *testing.T) {
	broker, mock, cleanup := setupTestBroker(t)
	defer cleanup()
	broker.WithLogger(log.New(io.Discard, "", 0))

	// TX1 (UC1)
	mock.ExpectBegin()
	mock.ExpectCommit() // Commits before UC2 (non-tx)

	// No tx for UC2 (non-tx)

	// TX2 (UC3)
	mock.ExpectBegin()
	mock.ExpectCommit() // Commits before UC4 (non-tx)

	// No tx for UC4 (non-tx)

	// TX3 (UC5)
	mock.ExpectBegin()
	mock.ExpectCommit() // Commits at end

	useCase1 := &mockUseCase{
		name:       "TransactionalUC1",
		output:     map[string]interface{}{"step1": "completed"},
		requiresTx: true,
	}
	useCase2 := &mockUseCase{
		name:       "NonTransactionalUC2",
		output:     map[string]interface{}{"step2": "completed"},
		requiresTx: false,
	}
	useCase3 := &mockUseCase{
		name:       "TransactionalUC3",
		output:     map[string]interface{}{"step3": "completed"},
		requiresTx: true,
	}
	useCase4 := &mockUseCase{
		name:       "NonTransactionalUC4",
		output:     map[string]interface{}{"step4": "completed"},
		requiresTx: false,
	}
	useCase5 := &mockUseCase{
		name:       "TransactionalUC5",
		output:     map[string]interface{}{"step5": "completed"},
		requiresTx: true,
	}

	useCases := []UseCase{useCase1, useCase2, useCase3, useCase4, useCase5}
	result, err := broker.RunUseCases(context.Background(), useCases, map[string]interface{}{})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// All steps should complete
	if len(result) != 5 {
		t.Errorf("expected 5 results, got %d", len(result))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// Test: Result chaining works across transaction boundaries
func TestTransactionBoundary_ResultChainingAcrossBoundaries(t *testing.T) {
	broker, mock, cleanup := setupTestBroker(t)
	defer cleanup()
	broker.WithLogger(log.New(io.Discard, "", 0))

	// TX1 (UC1)
	mock.ExpectBegin()
	mock.ExpectCommit()

	// TX2 (UC3)
	mock.ExpectBegin()
	mock.ExpectCommit()

	var uc2Input, uc3Input map[string]interface{}

	useCase1 := &mockUseCase{
		name:       "TransactionalUC1",
		output:     map[string]interface{}{"created_id": 123},
		requiresTx: true,
	}
	useCase2 := &mockUseCase{
		name:       "NonTransactionalUC2",
		requiresTx: false,
		executeFn: func(ctx context.Context, tx *sql.Tx, input map[string]interface{}) (map[string]interface{}, error) {
			uc2Input = make(map[string]interface{})
			for k, v := range input {
				uc2Input[k] = v
			}
			return map[string]interface{}{"validated": true}, nil
		},
	}
	useCase3 := &mockUseCase{
		name:       "TransactionalUC3",
		requiresTx: true,
		executeFn: func(ctx context.Context, tx *sql.Tx, input map[string]interface{}) (map[string]interface{}, error) {
			uc3Input = make(map[string]interface{})
			for k, v := range input {
				uc3Input[k] = v
			}
			return map[string]interface{}{"finalized": true}, nil
		},
	}

	useCases := []UseCase{useCase1, useCase2, useCase3}
	initialInput := map[string]interface{}{"user_id": 456}

	result, err := broker.RunUseCases(context.Background(), useCases, initialInput)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// UC2 should receive output from UC1 + initial input
	if uc2Input["user_id"] != 456 {
		t.Error("UC2 should receive initial input")
	}
	if uc2Input["created_id"] != 123 {
		t.Error("UC2 should receive output from UC1")
	}

	// UC3 should receive outputs from UC1 and UC2 + initial input
	if uc3Input["user_id"] != 456 {
		t.Error("UC3 should receive initial input")
	}
	if uc3Input["created_id"] != 123 {
		t.Error("UC3 should receive output from UC1")
	}
	if uc3Input["validated"] != true {
		t.Error("UC3 should receive output from UC2")
	}

	// Final result should exclude initial input
	if _, exists := result["user_id"]; exists {
		t.Error("final result should not include initial input")
	}

	// Final result should include all use case outputs
	if result["created_id"] != 123 || result["validated"] != true || result["finalized"] != true {
		t.Error("final result should include all use case outputs")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// Test: Single transactional use case followed by non-transactional
func TestTransactionBoundary_SingleTxFollowedByNonTx(t *testing.T) {
	broker, mock, cleanup := setupTestBroker(t)
	defer cleanup()
	broker.WithLogger(log.New(io.Discard, "", 0))

	// TX for UC1
	mock.ExpectBegin()
	mock.ExpectCommit() // Commits before UC2 (non-tx)

	useCase1 := &mockUseCase{
		name:       "TransactionalUC1",
		output:     map[string]interface{}{"step1": "completed"},
		requiresTx: true,
	}
	useCase2 := &mockUseCase{
		name:       "NonTransactionalUC2",
		output:     map[string]interface{}{"step2": "completed"},
		requiresTx: false,
	}

	useCases := []UseCase{useCase1, useCase2}
	result, err := broker.RunUseCases(context.Background(), useCases, map[string]interface{}{})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result["step1"] != "completed" || result["step2"] != "completed" {
		t.Error("expected all steps to complete")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// Test: Single non-transactional use case followed by transactional
func TestTransactionBoundary_SingleNonTxFollowedByTx(t *testing.T) {
	broker, mock, cleanup := setupTestBroker(t)
	defer cleanup()
	broker.WithLogger(log.New(io.Discard, "", 0))

	// TX for UC2
	mock.ExpectBegin()
	mock.ExpectCommit()

	useCase1 := &mockUseCase{
		name:       "NonTransactionalUC1",
		output:     map[string]interface{}{"step1": "completed"},
		requiresTx: false,
	}
	useCase2 := &mockUseCase{
		name:       "TransactionalUC2",
		output:     map[string]interface{}{"step2": "completed"},
		requiresTx: true,
	}

	useCases := []UseCase{useCase1, useCase2}
	result, err := broker.RunUseCases(context.Background(), useCases, map[string]interface{}{})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result["step1"] != "completed" || result["step2"] != "completed" {
		t.Error("expected all steps to complete")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}
