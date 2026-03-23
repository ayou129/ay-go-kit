package database

import (
	"context"
	"fmt"
	"strings"

	"github.com/ay/go-kit/logger"

	"gorm.io/gorm"
)

// ==================== 操作符常量 ====================

const (
	// 相等类
	FilterOpEqual    = "eq"  // 等于
	FilterOpNotEqual = "neq" // 不等于

	// 比较类
	FilterOpGT  = "gt"  // 大于
	FilterOpGTE = "gte" // 大于等于
	FilterOpLT  = "lt"  // 小于
	FilterOpLTE = "lte" // 小于等于

	// 范围类
	FilterOpBetween    = "between"     // 区间范围
	FilterOpNotBetween = "not_between" // 不在区间范围

	// 模糊匹配类
	FilterOpLike       = "like"        // 包含
	FilterOpNotLike    = "not_like"    // 不包含
	FilterOpStartsWith = "starts_with" // 前缀匹配
	FilterOpEndsWith   = "ends_with"   // 后缀匹配

	// 集合类
	FilterOpIn    = "in"     // 在集合中
	FilterOpNotIn = "not_in" // 不在集合中

	// 特殊类
	FilterOpIsNull    = "is_null"     // 是空值
	FilterOpIsNotNull = "is_not_null" // 不是空值
)

// ==================== 值类型常量 ====================

const (
	ValueTypeString  = "string"
	ValueTypeNumber  = "number"
	ValueTypeBoolean = "boolean"
	ValueTypeDate    = "date"
	ValueTypeArray   = "array"
)

// ==================== 数据结构 ====================

// FilterQuery 单个筛选条件
type FilterQuery struct {
	Field    string `json:"field"`    // 字段名
	Operator string `json:"operator"` // 操作符
	Value    any    `json:"value"`    // 值
}

// SortOption 排序选项
type SortOption struct {
	Field string `json:"sort_field"` // 排序字段
	Order string `json:"sort_order"` // 排序方向 (asc/desc)
}

// ==================== 核心转换 ====================

// ToScopes 将筛选和排序条件转为 GORM Scope 列表。
// allowed 为字段白名单，nil 表示不限制（开发阶段），非 nil 时跳过非白名单字段（生产安全）。
func ToScopes(ctx context.Context, filters []FilterQuery, sorts []SortOption, allowed map[string]bool) []Scope {
	var scopes []Scope

	for _, f := range filters {
		if allowed != nil && !allowed[f.Field] {
			continue
		}
		f := f // capture loop variable
		scopes = append(scopes, filterToScope(ctx, f))
	}

	for _, s := range sorts {
		if allowed != nil && !allowed[s.Field] {
			continue
		}
		order := s.Field + " " + strings.ToUpper(s.Order)
		scopes = append(scopes, func(db *gorm.DB) *gorm.DB {
			return db.Order(order)
		})
	}

	return scopes
}

func filterToScope(ctx context.Context, f FilterQuery) Scope {
	return func(db *gorm.DB) *gorm.DB {
		switch f.Operator {
		case FilterOpEqual:
			return db.Where(f.Field+" = ?", f.Value)
		case FilterOpNotEqual:
			return db.Where(f.Field+" != ?", f.Value)
		case FilterOpLike:
			if strVal, ok := f.Value.(string); ok {
				return db.Where(f.Field+" LIKE ?", "%"+strVal+"%")
			}
			logger.Error(ctx, "LIKE 操作符值类型错误: field=%s, value=%v", f.Field, f.Value)
			return db
		case FilterOpNotLike:
			if strVal, ok := f.Value.(string); ok {
				return db.Where(f.Field+" NOT LIKE ?", "%"+strVal+"%")
			}
			logger.Error(ctx, "NOT LIKE 操作符值类型错误: field=%s, value=%v", f.Field, f.Value)
			return db
		case FilterOpStartsWith:
			if strVal, ok := f.Value.(string); ok {
				return db.Where(f.Field+" LIKE ?", strVal+"%")
			}
			logger.Error(ctx, "STARTS WITH 操作符值类型错误: field=%s, value=%v", f.Field, f.Value)
			return db
		case FilterOpEndsWith:
			if strVal, ok := f.Value.(string); ok {
				return db.Where(f.Field+" LIKE ?", "%"+strVal)
			}
			logger.Error(ctx, "ENDS WITH 操作符值类型错误: field=%s, value=%v", f.Field, f.Value)
			return db
		case FilterOpGT:
			return db.Where(f.Field+" > ?", f.Value)
		case FilterOpGTE:
			return db.Where(f.Field+" >= ?", f.Value)
		case FilterOpLT:
			return db.Where(f.Field+" < ?", f.Value)
		case FilterOpLTE:
			return db.Where(f.Field+" <= ?", f.Value)
		case FilterOpBetween:
			if arr, ok := f.Value.([]any); ok && len(arr) >= 2 {
				return db.Where(f.Field+" BETWEEN ? AND ?", arr[0], arr[1])
			}
			logger.Error(ctx, "BETWEEN 操作符值类型错误: field=%s, value=%v", f.Field, f.Value)
			return db
		case FilterOpNotBetween:
			if arr, ok := f.Value.([]any); ok && len(arr) >= 2 {
				return db.Where(f.Field+" NOT BETWEEN ? AND ?", arr[0], arr[1])
			}
			logger.Error(ctx, "NOT BETWEEN 操作符值类型错误: field=%s, value=%v", f.Field, f.Value)
			return db
		case FilterOpIn:
			return db.Where(f.Field+" IN ?", f.Value)
		case FilterOpNotIn:
			return db.Where(f.Field+" NOT IN ?", f.Value)
		case FilterOpIsNull:
			return db.Where(f.Field + " IS NULL")
		case FilterOpIsNotNull:
			return db.Where(f.Field + " IS NOT NULL")
		default:
			logger.Error(ctx, "不支持的操作符: %s, 已忽略", f.Operator)
			return db
		}
	}
}

// ==================== 辅助函数 ====================

// ShouldShowValueInput 判断操作符是否需要输入值（前端表单用）
func ShouldShowValueInput(operator string) bool {
	return operator != FilterOpIsNull && operator != FilterOpIsNotNull
}

// GetAvailableOperatorsByType 根据值类型获取可用的操作符列表（前端表单用）
func GetAvailableOperatorsByType(valueType string) []string {
	switch valueType {
	case ValueTypeBoolean:
		return []string{FilterOpEqual, FilterOpNotEqual}
	case ValueTypeDate, ValueTypeNumber:
		return []string{
			FilterOpEqual, FilterOpNotEqual,
			FilterOpGT, FilterOpGTE, FilterOpLT, FilterOpLTE,
			FilterOpBetween,
			FilterOpIsNull, FilterOpIsNotNull,
		}
	case ValueTypeString:
		return []string{
			FilterOpEqual, FilterOpNotEqual,
			FilterOpLike, FilterOpNotLike,
			FilterOpStartsWith, FilterOpEndsWith,
			FilterOpIsNull, FilterOpIsNotNull,
		}
	case ValueTypeArray:
		return []string{FilterOpIn, FilterOpNotIn}
	default:
		return []string{
			FilterOpEqual, FilterOpNotEqual,
			FilterOpLike, FilterOpNotLike,
			FilterOpStartsWith, FilterOpEndsWith,
			FilterOpIsNull, FilterOpIsNotNull,
		}
	}
}

// FormatFilterDesc 格式化筛选条件用于日志（调试用）
func FormatFilterDesc(f FilterQuery) string {
	return fmt.Sprintf("%s %s %v", f.Field, f.Operator, f.Value)
}
