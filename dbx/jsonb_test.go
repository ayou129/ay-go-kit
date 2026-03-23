package dbx

import (
	"testing"
)

func TestJSONBObject_ValueScan(t *testing.T) {
	obj := JSONBObject{"name": "test", "count": float64(42)}
	v, err := obj.Value()
	if err != nil {
		t.Fatalf("Value() error: %v", err)
	}

	var got JSONBObject
	if err := got.Scan(v); err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	if got["name"] != "test" || got["count"] != float64(42) {
		t.Fatalf("got %v, want name=test count=42", got)
	}
}

func TestJSONBObject_Nil(t *testing.T) {
	var obj JSONBObject
	v, err := obj.Value()
	if err != nil || v != nil {
		t.Fatalf("nil Value() = %v, %v", v, err)
	}

	if err := obj.Scan(nil); err != nil || obj != nil {
		t.Fatalf("Scan(nil) = %v, %v", obj, err)
	}
}

func TestJSONBArray_ValueScan(t *testing.T) {
	arr := JSONBArray{"a", float64(1), true}
	v, err := arr.Value()
	if err != nil {
		t.Fatalf("Value() error: %v", err)
	}

	var got JSONBArray
	if err := got.Scan(v); err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	if len(got) != 3 || got[0] != "a" {
		t.Fatalf("got %v", got)
	}
}

func TestJSONBArray_Nil(t *testing.T) {
	var arr JSONBArray
	v, err := arr.Value()
	if err != nil || v != nil {
		t.Fatalf("nil Value() = %v, %v", v, err)
	}
}

func TestJSONBArrayStr_ValueScan(t *testing.T) {
	arr := JSONBArrayStr{"选项A", "选项B"}
	v, err := arr.Value()
	if err != nil {
		t.Fatalf("Value() error: %v", err)
	}

	var got JSONBArrayStr
	if err := got.Scan(v); err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	if len(got) != 2 || got[0] != "选项A" || got[1] != "选项B" {
		t.Fatalf("got %v", got)
	}
}

func TestJSONBArrayStr_Nil(t *testing.T) {
	var arr JSONBArrayStr
	v, err := arr.Value()
	if err != nil || v != nil {
		t.Fatalf("nil Value() = %v, %v", v, err)
	}
}

func TestJSONB_ScanInvalidType(t *testing.T) {
	var obj JSONBObject
	if err := obj.Scan(123); err == nil {
		t.Fatal("expected error for non-[]byte input")
	}

	var arr JSONBArray
	if err := arr.Scan(123); err == nil {
		t.Fatal("expected error for non-[]byte input")
	}

	var str JSONBArrayStr
	if err := str.Scan(123); err == nil {
		t.Fatal("expected error for non-[]byte input")
	}
}
