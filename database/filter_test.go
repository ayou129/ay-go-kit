package database

import (
	"context"
	"testing"

	"gorm.io/gorm"
)

// ==================== 操作符常量 ====================

func TestFilterOpConstants(t *testing.T) {
	ops := []string{
		FilterOpEqual, FilterOpNotEqual,
		FilterOpGT, FilterOpGTE, FilterOpLT, FilterOpLTE,
		FilterOpBetween, FilterOpNotBetween,
		FilterOpLike, FilterOpNotLike, FilterOpStartsWith, FilterOpEndsWith,
		FilterOpIn, FilterOpNotIn,
		FilterOpIsNull, FilterOpIsNotNull,
	}
	for _, op := range ops {
		if op == "" {
			t.Error("filter operator constant is empty")
		}
	}
	if len(ops) != 16 {
		t.Errorf("expected 16 operators, got %d", len(ops))
	}
}

// ==================== filterToScope 操作符测试 ====================
// 验证每个操作符生成的 Scope 函数非 nil 且可调用（不 panic）

func scopeFromFilter(op string, value any) Scope {
	ctx := context.Background()
	scopes := ToScopes(ctx, []FilterQuery{{Field: "name", Operator: op, Value: value}}, nil, nil)
	if len(scopes) != 1 {
		return nil
	}
	return scopes[0]
}

func assertScopeNotNil(t *testing.T, op string, value any) {
	t.Helper()
	s := scopeFromFilter(op, value)
	if s == nil {
		t.Errorf("scope for operator %q is nil", op)
	}
}

func TestFilterToScope_AllOperators_ProduceScope(t *testing.T) {
	// 每个操作符都应该生成一个非 nil 的 Scope
	tests := []struct {
		op    string
		value any
	}{
		{FilterOpEqual, "test"},
		{FilterOpNotEqual, "test"},
		{FilterOpGT, 10},
		{FilterOpGTE, 10},
		{FilterOpLT, 10},
		{FilterOpLTE, 10},
		{FilterOpLike, "test"},
		{FilterOpNotLike, "test"},
		{FilterOpStartsWith, "pre"},
		{FilterOpEndsWith, "suf"},
		{FilterOpBetween, []any{1, 10}},
		{FilterOpNotBetween, []any{1, 10}},
		{FilterOpIn, []any{1, 2, 3}},
		{FilterOpNotIn, []any{1, 2, 3}},
		{FilterOpIsNull, nil},
		{FilterOpIsNotNull, nil},
	}

	for _, tt := range tests {
		t.Run(tt.op, func(t *testing.T) {
			assertScopeNotNil(t, tt.op, tt.value)
		})
	}
}

func TestFilterToScope_LikeWrongType_NoError(t *testing.T) {
	// Value 不是 string 时应该生成 scope 但不 panic
	ops := []string{FilterOpLike, FilterOpNotLike, FilterOpStartsWith, FilterOpEndsWith}
	for _, op := range ops {
		t.Run(op, func(t *testing.T) {
			s := scopeFromFilter(op, 123) // int instead of string
			if s == nil {
				t.Errorf("scope for %q with wrong type should still exist", op)
			}
			// 验证 scope 应用到 DB 不 panic
			db := &gorm.DB{Config: &gorm.Config{}, Statement: &gorm.Statement{}}
			s(db) // 不应 panic
		})
	}
}

func TestFilterToScope_BetweenWrongType_NoError(t *testing.T) {
	ops := []string{FilterOpBetween, FilterOpNotBetween}
	for _, op := range ops {
		t.Run(op, func(t *testing.T) {
			s := scopeFromFilter(op, "not_array")
			if s == nil {
				t.Errorf("scope for %q with wrong type should still exist", op)
			}
			db := &gorm.DB{Config: &gorm.Config{}, Statement: &gorm.Statement{}}
			s(db) // 不应 panic
		})
	}
}

func TestFilterToScope_UnknownOperator_NoError(t *testing.T) {
	s := scopeFromFilter("unknown_op", "value")
	if s == nil {
		t.Error("scope for unknown operator should still exist (no-op)")
	}
	db := &gorm.DB{Config: &gorm.Config{}, Statement: &gorm.Statement{}}
	s(db) // 不应 panic
}

// ==================== Sort 测试 ====================

func TestToScopes_SortGeneratesScope(t *testing.T) {
	ctx := context.Background()
	scopes := ToScopes(ctx, nil, []SortOption{
		{Field: "name", Order: "asc"},
		{Field: "id", Order: "desc"},
	}, nil)
	if len(scopes) != 2 {
		t.Errorf("expected 2 sort scopes, got %d", len(scopes))
	}
}

// ==================== 白名单 + 空输入 ====================

func TestToScopes_AllowedFieldsFilter(t *testing.T) {
	ctx := context.Background()
	filters := []FilterQuery{
		{Field: "name", Operator: FilterOpEqual, Value: "test"},
		{Field: "secret", Operator: FilterOpEqual, Value: "hack"},
	}
	sorts := []SortOption{
		{Field: "name", Order: "asc"},
		{Field: "hidden", Order: "desc"},
	}
	allowed := map[string]bool{"name": true}

	scopes := ToScopes(ctx, filters, sorts, allowed)
	if len(scopes) != 2 {
		t.Errorf("ToScopes with allowed fields returned %d scopes, want 2", len(scopes))
	}
}

func TestToScopes_NilAllowed_PassesAll(t *testing.T) {
	ctx := context.Background()
	filters := []FilterQuery{
		{Field: "a", Operator: FilterOpEqual, Value: 1},
		{Field: "b", Operator: FilterOpLike, Value: "x"},
	}
	sorts := []SortOption{{Field: "c", Order: "desc"}}

	scopes := ToScopes(ctx, filters, sorts, nil)
	if len(scopes) != 3 {
		t.Errorf("ToScopes with nil allowed returned %d scopes, want 3", len(scopes))
	}
}

func TestToScopes_Empty(t *testing.T) {
	ctx := context.Background()
	scopes := ToScopes(ctx, nil, nil, nil)
	if len(scopes) != 0 {
		t.Errorf("ToScopes with empty inputs returned %d scopes, want 0", len(scopes))
	}
}

// ==================== 辅助函数 ====================

func TestShouldShowValueInput(t *testing.T) {
	tests := []struct {
		operator string
		want     bool
	}{
		{FilterOpEqual, true},
		{FilterOpLike, true},
		{FilterOpBetween, true},
		{FilterOpIn, true},
		{FilterOpIsNull, false},
		{FilterOpIsNotNull, false},
	}

	for _, tt := range tests {
		if got := ShouldShowValueInput(tt.operator); got != tt.want {
			t.Errorf("ShouldShowValueInput(%q) = %v, want %v", tt.operator, got, tt.want)
		}
	}
}

func TestGetAvailableOperatorsByType(t *testing.T) {
	tests := []struct {
		valueType string
		minCount  int
	}{
		{ValueTypeString, 6},
		{ValueTypeNumber, 7},
		{ValueTypeDate, 7},
		{ValueTypeBoolean, 2},
		{ValueTypeArray, 2},
		{"unknown", 6},
	}

	for _, tt := range tests {
		ops := GetAvailableOperatorsByType(tt.valueType)
		if len(ops) < tt.minCount {
			t.Errorf("GetAvailableOperatorsByType(%q) returned %d operators, want at least %d", tt.valueType, len(ops), tt.minCount)
		}
	}
}

func TestFormatFilterDesc(t *testing.T) {
	f := FilterQuery{Field: "name", Operator: FilterOpLike, Value: "test"}
	desc := FormatFilterDesc(f)
	if desc != "name like test" {
		t.Errorf("FormatFilterDesc = %q, want %q", desc, "name like test")
	}
}
