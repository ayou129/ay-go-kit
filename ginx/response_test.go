package ginx

import (
	"reflect"
	"testing"
)

func TestApiResponse_JsonTags(t *testing.T) {
	typ := reflect.TypeOf(ApiResponse{})
	expected := map[string]string{
		"Code": `json:"code" example:"0"`,
		"Msg":  `json:"msg" example:"成功"`,
		"Data": `json:"data"`,
	}
	for field, wantTag := range expected {
		f, ok := typ.FieldByName(field)
		if !ok {
			t.Errorf("field %s not found", field)
			continue
		}
		got := string(f.Tag)
		if got != wantTag {
			t.Errorf("field %s: tag = %q, want %q", field, got, wantTag)
		}
	}
}

func TestPageResponse_JsonTags(t *testing.T) {
	type Item struct{ ID int }
	typ := reflect.TypeOf(PageResponse[Item]{})
	expected := map[string]string{
		"Total":    `json:"total" example:"100"`,
		"Page":     `json:"page" example:"1"`,
		"PageSize": `json:"page_size" example:"20"`,
		"List":     `json:"list"`,
	}
	for field, wantTag := range expected {
		f, ok := typ.FieldByName(field)
		if !ok {
			t.Errorf("field %s not found", field)
			continue
		}
		got := string(f.Tag)
		if got != wantTag {
			t.Errorf("field %s: tag = %q, want %q", field, got, wantTag)
		}
	}
}

func TestPageResponse_ListType(t *testing.T) {
	p := PageResponse[string]{
		Total:    10,
		Page:     1,
		PageSize: 5,
		List:     []string{"a", "b"},
	}
	if len(p.List) != 2 {
		t.Errorf("expected 2 items, got %d", len(p.List))
	}
	if p.List[0] != "a" {
		t.Errorf("expected first item 'a', got %s", p.List[0])
	}
}
