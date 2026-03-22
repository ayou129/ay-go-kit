package database

import (
	"context"
	"testing"

	"gorm.io/gorm"
)

func TestSetGlobalAndDB(t *testing.T) {
	old := globalDB
	defer func() { globalDB = old }()

	fakeDB := &gorm.DB{Config: &gorm.Config{}}
	SetGlobal(fakeDB)

	if DB() != fakeDB {
		t.Error("DB() should return the global instance set by SetGlobal")
	}
}

func TestGetDB_ReturnsNilWhenNotInitialized(t *testing.T) {
	old := globalDB
	defer func() { globalDB = old }()

	globalDB = nil
	got := GetDB(context.Background())
	if got != nil {
		t.Errorf("GetDB should return nil when global not set, got %v", got)
	}
}

func TestGetDB_ReturnsTxFromContext(t *testing.T) {
	old := globalDB
	defer func() { globalDB = old }()

	fakeDB := &gorm.DB{Config: &gorm.Config{SkipDefaultTransaction: false}}
	fakeTx := &gorm.DB{Config: &gorm.Config{SkipDefaultTransaction: true}}
	SetGlobal(fakeDB)

	ctx := WithTx(context.Background(), fakeTx)
	got := GetDB(ctx)
	if got != fakeTx {
		t.Error("GetDB should return tx from context when present")
	}
}

func TestOverrideGetDB(t *testing.T) {
	old := globalDB
	defer func() { globalDB = old }()

	original := &gorm.DB{Config: &gorm.Config{SkipDefaultTransaction: false}}
	override := &gorm.DB{Config: &gorm.Config{SkipDefaultTransaction: true}}
	SetGlobal(original)

	restore := OverrideGetDB(override)
	if DB() != override {
		t.Error("after OverrideGetDB, DB() should return override")
	}

	restore()
	if DB() != original {
		t.Error("after restore, DB() should return original")
	}
}
