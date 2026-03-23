package ginx

import (
	"testing"
)

func TestCleanString_ZeroWidthChars(t *testing.T) {
	input := "hello\u200Bworld\uFEFF"
	want := "helloworld"
	if got := CleanString(input); got != want {
		t.Errorf("CleanString(%q) = %q, want %q", input, got, want)
	}
}

func TestCleanString_ControlChars(t *testing.T) {
	input := "hello\x00\x01\x02world"
	want := "helloworld"
	if got := CleanString(input); got != want {
		t.Errorf("CleanString(%q) = %q, want %q", input, got, want)
	}
}

func TestCleanString_FullWidthSpace(t *testing.T) {
	input := "hello\u3000world"
	want := "hello world"
	if got := CleanString(input); got != want {
		t.Errorf("CleanString(%q) = %q, want %q", input, got, want)
	}
}

func TestCleanString_ConsecutiveSpaces(t *testing.T) {
	input := "hello   world"
	want := "hello world"
	if got := CleanString(input); got != want {
		t.Errorf("CleanString(%q) = %q, want %q", input, got, want)
	}
}

func TestCleanString_Tab(t *testing.T) {
	input := "hello\tworld"
	want := "hello world"
	if got := CleanString(input); got != want {
		t.Errorf("CleanString(%q) = %q, want %q", input, got, want)
	}
}

func TestCleanString_Trim(t *testing.T) {
	input := "  hello  "
	want := "hello"
	if got := CleanString(input); got != want {
		t.Errorf("CleanString(%q) = %q, want %q", input, got, want)
	}
}

func TestCleanString_Empty(t *testing.T) {
	if got := CleanString(""); got != "" {
		t.Errorf("CleanString empty = %q, want empty", got)
	}
}

func TestCleanString_Combined(t *testing.T) {
	input := "  \u200Bhello\u3000\t  world\uFEFF  "
	want := "hello world"
	if got := CleanString(input); got != want {
		t.Errorf("CleanString combined = %q, want %q", got, want)
	}
}

func TestSanitizeStrings_SimpleStruct(t *testing.T) {
	type req struct {
		Name  string
		Email string
	}
	r := &req{Name: "  hello  ", Email: "a\u200Bb"}
	SanitizeStrings(r)
	if r.Name != "hello" {
		t.Errorf("Name = %q, want %q", r.Name, "hello")
	}
	if r.Email != "ab" {
		t.Errorf("Email = %q, want %q", r.Email, "ab")
	}
}

func TestSanitizeStrings_SkipTag(t *testing.T) {
	type req struct {
		Name     string
		Password string `sanitize:"-"`
	}
	r := &req{Name: "  hello  ", Password: "  secret  "}
	SanitizeStrings(r)
	if r.Name != "hello" {
		t.Errorf("Name = %q, want %q", r.Name, "hello")
	}
	if r.Password != "  secret  " {
		t.Errorf("Password should be untouched, got %q", r.Password)
	}
}

func TestSanitizeStrings_NestedStruct(t *testing.T) {
	type inner struct {
		Value string
	}
	type outer struct {
		Inner inner
	}
	o := &outer{Inner: inner{Value: "  test  "}}
	SanitizeStrings(o)
	if o.Inner.Value != "test" {
		t.Errorf("Inner.Value = %q, want %q", o.Inner.Value, "test")
	}
}

func TestSanitizeStrings_PointerStruct(t *testing.T) {
	type inner struct {
		Value string
	}
	type outer struct {
		Inner *inner
	}
	o := &outer{Inner: &inner{Value: "  test  "}}
	SanitizeStrings(o)
	if o.Inner.Value != "test" {
		t.Errorf("Inner.Value = %q, want %q", o.Inner.Value, "test")
	}
}

func TestSanitizeStrings_NilPointerStruct(t *testing.T) {
	type inner struct {
		Value string
	}
	type outer struct {
		Inner *inner
	}
	o := &outer{Inner: nil}
	// Should not panic
	SanitizeStrings(o)
}

func TestSanitizeMap(t *testing.T) {
	m := map[string]any{
		"name":  "  hello  ",
		"count": 42,
		"email": "a\u200Bb",
	}
	SanitizeMap(m)
	if m["name"] != "hello" {
		t.Errorf("name = %q, want %q", m["name"], "hello")
	}
	if m["count"] != 42 {
		t.Errorf("count should be unchanged, got %v", m["count"])
	}
	if m["email"] != "ab" {
		t.Errorf("email = %q, want %q", m["email"], "ab")
	}
}
