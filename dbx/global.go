package dbx

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

// Scope 是 GORM 原生的查询修饰函数，用于组合 WHERE/ORDER/JOIN 等条件
type Scope = func(*gorm.DB) *gorm.DB

// IsRecordNotFound 检查是否是记录不存在错误（支持 errors.Is 链式判断）
func IsRecordNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

var globalDB *gorm.DB

// SetGlobal sets the global database instance (call once at startup after Open)
func SetGlobal(db *gorm.DB) {
	globalDB = db
}

// DB returns the global database instance
func DB() *gorm.DB {
	return globalDB
}

// GetDB returns a context-aware session from the global instance.
// If ctx contains a transaction (via WithTx), the transaction is returned.
func GetDB(ctx context.Context) *gorm.DB {
	if tx := TxFromContext(ctx); tx != nil {
		return tx
	}
	if globalDB == nil {
		return nil
	}
	session := globalDB.Session(&gorm.Session{})
	if ctx != nil {
		return session.WithContext(ctx)
	}
	return session
}

// OverrideGetDB replaces the global DB instance temporarily (for testing with txdb).
// Returns a restore function.
func OverrideGetDB(override *gorm.DB) func() {
	old := globalDB
	globalDB = override
	return func() { globalDB = old }
}
