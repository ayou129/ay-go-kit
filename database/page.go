package database

import (
	"context"
	"fmt"
)

// PageQuery 通用分页参数
type PageQuery struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

// Normalize 自动修正分页参数到合法范围
func (q *PageQuery) Normalize() {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.PageSize < 1 {
		q.PageSize = 20
	}
	if q.PageSize > 100 {
		q.PageSize = 100
	}
}

// Offset 计算 SQL OFFSET 值
func (q PageQuery) Offset() int {
	return (q.Page - 1) * q.PageSize
}

// PageResult 通用分页响应
type PageResult[T any] struct {
	Total    int64 `json:"total" example:"100"`
	Page     int   `json:"page" example:"1"`
	PageSize int   `json:"page_size" example:"20"`
	List     []T   `json:"list"`
}

// FindByPage 泛型分页查询，消灭 repo 层的重复 Count+Offset+Limit+Find 样板代码。
// T 必须是 GORM 模型（有对应数据表）。
// scopes 用于附加 WHERE/ORDER 等条件。
func FindByPage[T any](ctx context.Context, pq PageQuery, scopes ...Scope) (*PageResult[T], error) {
	db := GetDB(ctx)
	if db == nil {
		return nil, fmt.Errorf("database: global DB not initialized")
	}
	db = db.Model(new(T))

	for _, s := range scopes {
		db = s(db)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("database.FindByPage count: %w", err)
	}

	if total == 0 {
		return &PageResult[T]{
			Total:    0,
			Page:     pq.Page,
			PageSize: pq.PageSize,
			List:     []T{},
		}, nil
	}

	var list []T
	if err := db.Offset(pq.Offset()).Limit(pq.PageSize).Find(&list).Error; err != nil {
		return nil, fmt.Errorf("database.FindByPage find: %w", err)
	}

	return &PageResult[T]{
		Total:    total,
		Page:     pq.Page,
		PageSize: pq.PageSize,
		List:     list,
	}, nil
}
