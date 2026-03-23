package ginx

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/ay/go-kit/database"
)

func TestPageQueryDTO_Normalize_DefaultSort(t *testing.T) {
	q := &PageQueryDTO{}
	q.normalize()

	if q.Page != 1 {
		t.Errorf("Page = %d, want 1", q.Page)
	}
	if q.PageSize != 20 {
		t.Errorf("PageSize = %d, want 20", q.PageSize)
	}
	if len(q.Sorts) != 1 || q.Sorts[0].Field != "id" || q.Sorts[0].Order != "desc" {
		t.Errorf("default sort = %+v, want [{id desc}]", q.Sorts)
	}
	if q.Filters == nil {
		t.Error("Filters should be initialized to empty slice, not nil")
	}
}

func TestPageQueryDTO_Normalize_FixInvalidSort(t *testing.T) {
	q := &PageQueryDTO{
		Sorts: []database.SortOption{
			{Field: "", Order: "invalid"},
		},
	}
	q.normalize()

	if q.Sorts[0].Field != "id" {
		t.Errorf("empty sort field should default to 'id', got %q", q.Sorts[0].Field)
	}
	if q.Sorts[0].Order != "desc" {
		t.Errorf("invalid sort order should default to 'desc', got %q", q.Sorts[0].Order)
	}
}

func TestPageQueryDTO_Normalize_PreservesValidSort(t *testing.T) {
	q := &PageQueryDTO{
		Sorts: []database.SortOption{
			{Field: "created_at", Order: "asc"},
		},
	}
	q.PageQuery = database.PageQuery{Page: 2, PageSize: 50}
	q.normalize()

	if q.Sorts[0].Field != "created_at" || q.Sorts[0].Order != "asc" {
		t.Errorf("valid sort should be preserved, got %+v", q.Sorts[0])
	}
	if q.Page != 2 || q.PageSize != 50 {
		t.Errorf("valid page params should be preserved, got page=%d pageSize=%d", q.Page, q.PageSize)
	}
}

func TestPageQueryDTO_JSONTags(t *testing.T) {
	typ := reflect.TypeOf(PageQueryDTO{})

	// Check Filters field
	f, ok := typ.FieldByName("Filters")
	if !ok {
		t.Fatal("Filters field not found")
	}
	if tag := f.Tag.Get("json"); tag != "filters" {
		t.Errorf("Filters json tag = %q, want %q", tag, "filters")
	}

	// Check Sorts field
	s, ok := typ.FieldByName("Sorts")
	if !ok {
		t.Fatal("Sorts field not found")
	}
	if tag := s.Tag.Get("json"); tag != "sorts" {
		t.Errorf("Sorts json tag = %q, want %q", tag, "sorts")
	}
}

func TestPageQueryDTO_JSONRoundtrip(t *testing.T) {
	input := `{"page":2,"page_size":15,"filters":[{"field":"name","operator":"like","value":"test"}],"sorts":[{"sort_field":"id","sort_order":"asc"}]}`

	var q PageQueryDTO
	if err := json.Unmarshal([]byte(input), &q); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if q.Page != 2 {
		t.Errorf("Page = %d, want 2", q.Page)
	}
	if q.PageSize != 15 {
		t.Errorf("PageSize = %d, want 15", q.PageSize)
	}
	if len(q.Filters) != 1 || q.Filters[0].Field != "name" {
		t.Errorf("Filters = %+v, want [{name like test}]", q.Filters)
	}
	if len(q.Sorts) != 1 || q.Sorts[0].Field != "id" {
		t.Errorf("Sorts = %+v, want [{id asc}]", q.Sorts)
	}
}
