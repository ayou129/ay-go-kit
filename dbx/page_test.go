package dbx

import (
	"encoding/json"
	"testing"
)

func TestPageQuery_Normalize(t *testing.T) {
	tests := []struct {
		name         string
		input        PageQuery
		wantPage     int
		wantPageSize int
	}{
		{"valid", PageQuery{Page: 2, PageSize: 10}, 2, 10},
		{"page < 1", PageQuery{Page: 0, PageSize: 10}, 1, 10},
		{"negative page", PageQuery{Page: -5, PageSize: 10}, 1, 10},
		{"pageSize < 1", PageQuery{Page: 1, PageSize: 0}, 1, 20},
		{"negative pageSize", PageQuery{Page: 1, PageSize: -1}, 1, 20},
		{"pageSize > 100", PageQuery{Page: 1, PageSize: 200}, 1, 100},
		{"both invalid", PageQuery{Page: -1, PageSize: -1}, 1, 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := tt.input
			q.Normalize()
			if q.Page != tt.wantPage {
				t.Errorf("Page = %d, want %d", q.Page, tt.wantPage)
			}
			if q.PageSize != tt.wantPageSize {
				t.Errorf("PageSize = %d, want %d", q.PageSize, tt.wantPageSize)
			}
		})
	}
}

func TestPageQuery_Offset(t *testing.T) {
	tests := []struct {
		page     int
		pageSize int
		want     int
	}{
		{1, 20, 0},
		{2, 20, 20},
		{3, 10, 20},
		{1, 100, 0},
	}

	for _, tt := range tests {
		q := PageQuery{Page: tt.page, PageSize: tt.pageSize}
		if got := q.Offset(); got != tt.want {
			t.Errorf("PageQuery{%d,%d}.Offset() = %d, want %d", tt.page, tt.pageSize, got, tt.want)
		}
	}
}

func TestPageResult_JSONTags(t *testing.T) {
	result := PageResult[struct{}]{
		Total:    100,
		Page:     1,
		PageSize: 20,
		List:     []struct{}{},
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	expectedKeys := []string{"total", "page", "page_size", "list"}
	for _, key := range expectedKeys {
		if _, ok := m[key]; !ok {
			t.Errorf("expected JSON key %q not found", key)
		}
	}
}
