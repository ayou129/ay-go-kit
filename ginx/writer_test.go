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
