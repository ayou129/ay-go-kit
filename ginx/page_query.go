package ginx

import (
	"context"

	"github.com/ay/go-kit/dbx"

	"github.com/gin-gonic/gin"
)

// PageQueryDTO 通用分页查询请求（含筛选+排序），提供 gin 绑定层。
// 嵌入 dbx.PageQuery 获得基础分页能力。
type PageQueryDTO struct {
	dbx.PageQuery
	Filters []dbx.FilterQuery `json:"filters" form:"filters" swaggertype:"array,object"` // 筛选条件列表
	Sorts   []dbx.SortOption  `json:"sorts" form:"sorts" swaggertype:"array,object"`     // 多字段排序
}

// Bind 绑定 JSON 请求参数并自动修正到合法范围
func (q *PageQueryDTO) Bind(c *gin.Context) error {
	if err := c.ShouldBindJSON(q); err != nil {
		return err
	}
	q.normalize()
	return nil
}

// normalize 自动修正所有字段到合法范围
func (q *PageQueryDTO) normalize() {
	q.PageQuery.Normalize()

	// 修正 sorts
	if len(q.Sorts) == 0 {
		q.Sorts = []dbx.SortOption{{Field: "id", Order: "desc"}}
	}
	for i := range q.Sorts {
		if q.Sorts[i].Order != "asc" && q.Sorts[i].Order != "desc" {
			q.Sorts[i].Order = "desc"
		}
		if q.Sorts[i].Field == "" {
			q.Sorts[i].Field = "id"
		}
	}

	// 修正 filters
	if q.Filters == nil {
		q.Filters = []dbx.FilterQuery{}
	}
}

// ToScopes 将筛选和排序条件转为 GORM Scope 列表。
// allowed 为字段白名单，nil 表示不限制。
func (q *PageQueryDTO) ToScopes(ctx context.Context, allowed map[string]bool) []dbx.Scope {
	return dbx.ToScopes(ctx, q.Filters, q.Sorts, allowed)
}
