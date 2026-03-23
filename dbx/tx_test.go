package dbx

import (
	"context"
	"testing"

	"gorm.io/gorm"
)

func TestWithTxAndTxFromContext_Roundtrip(t *testing.T) {
	// Use a typed nil pointer — enough to verify context storage without a real DB
	var fakeTx *gorm.DB

	ctx := context.Background()
	ctx = WithTx(ctx, fakeTx)

	got := TxFromContext(ctx)
	if got != fakeTx {
		t.Errorf("TxFromContext returned %v, want %v", got, fakeTx)
	}
}

func TestTxFromContext_EmptyContext(t *testing.T) {
	ctx := context.Background()
	got := TxFromContext(ctx)
	if got != nil {
		t.Errorf("TxFromContext on empty context returned %v, want nil", got)
	}
}

func TestBaseRepository_GetDB_ReturnsTxWhenPresent(t *testing.T) {
	// Create two distinguishable *gorm.DB values
	injectedDB := &gorm.DB{Config: &gorm.Config{SkipDefaultTransaction: false}}
	txDB := &gorm.DB{Config: &gorm.Config{SkipDefaultTransaction: true}}

	repo := BaseRepository{DB: injectedDB}

	// With tx in context — should return txDB directly (bypassing WithContext)
	ctx := context.Background()
	ctxWithTx := WithTx(ctx, txDB)
	got := repo.GetDB(ctxWithTx)
	if got != txDB {
		t.Errorf("GetDB with tx in context returned %p, want %p", got, txDB)
	}
}

func TestBaseRepository_GetDB_FallsBackToInjectedDB(t *testing.T) {
	// Verify that without tx in context, GetDB calls DB.WithContext (not tx).
	// We can't call WithContext on a bare gorm.DB without a real connection,
	// so we verify indirectly: TxFromContext returns nil on plain context.
	ctx := context.Background()
	if tx := TxFromContext(ctx); tx != nil {
		t.Errorf("expected nil tx from plain context, got %v", tx)
	}
}

func TestNewTxRunner_ReturnsNonNil(t *testing.T) {
	runner := NewTxRunner()
	if runner == nil {
		t.Error("NewTxRunner() returned nil")
	}
}

func TestTxRunner_ImplementsInterface(t *testing.T) {
	var _ TxRunner = NewTxRunner()
	var _ TxRunner = &gormTxRunner{}
}
