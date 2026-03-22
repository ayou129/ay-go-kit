package database

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

type txKey struct{}

// WithTx stores a transaction in context
func WithTx(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

// TxFromContext retrieves a transaction from context, nil if none
func TxFromContext(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx
	}
	return nil
}

// Transaction runs fn within a transaction, auto commit/rollback
func Transaction(ctx context.Context, db *gorm.DB, fn func(ctx context.Context) error) error {
	tx := db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("begin transaction: %w", tx.Error)
	}
	txCtx := WithTx(ctx, tx)
	if err := fn(txCtx); err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

// BaseRepository provides DB injection and transaction awareness
type BaseRepository struct {
	DB *gorm.DB
}

// GetDB returns the transaction from context if present, otherwise the injected DB
func (r *BaseRepository) GetDB(ctx context.Context) *gorm.DB {
	if tx := TxFromContext(ctx); tx != nil {
		return tx
	}
	return r.DB.WithContext(ctx)
}
