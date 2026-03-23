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

// TxRunner 事务执行接口，支持依赖注入，替代直接调用全局 Transaction 函数。
// Service 层注入 TxRunner 后，事务内的 fn 通过 database.GetDB(txCtx) 自动获取事务连接。
type TxRunner interface {
	Transaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type gormTxRunner struct{}

// NewTxRunner 创建基于全局 DB 的事务执行器
func NewTxRunner() TxRunner {
	return &gormTxRunner{}
}

func (r *gormTxRunner) Transaction(ctx context.Context, fn func(ctx context.Context) error) error {
	db := GetDB(ctx)
	return Transaction(ctx, db, fn)
}

// Deprecated: BaseRepository 已废弃，请在 repo 方法内直接调用 database.GetDB(ctx) 替代。
// 将在下个大版本移除。
type BaseRepository struct {
	DB *gorm.DB
}

// Deprecated: 使用 database.GetDB(ctx) 替代。
func (r *BaseRepository) GetDB(ctx context.Context) *gorm.DB {
	if tx := TxFromContext(ctx); tx != nil {
		return tx
	}
	return r.DB.WithContext(ctx)
}
