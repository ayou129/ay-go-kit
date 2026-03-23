package timex_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/ay/go-kit/timex"
)

// ==================== Scan ====================

func TestDateModelScanTime(t *testing.T) {
	shanghai, _ := time.LoadLocation("Asia/Shanghai")

	tests := []struct {
		name       string
		input      time.Time
		wantString string
	}{
		{
			"正常日期",
			time.Date(2026, 3, 23, 0, 0, 0, 0, shanghai),
			"2026-03-23",
		},
		{
			"带时分秒的时间（只保留日期）",
			time.Date(2026, 3, 23, 14, 30, 0, 0, shanghai),
			"2026-03-23",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var d timex.DateModel
			if err := d.Scan(tt.input); err != nil {
				t.Fatalf("Scan() error = %v", err)
			}
			got := d.String()
			if got != tt.wantString {
				t.Errorf("Scan(%v).String() = %q, want %q", tt.input, got, tt.wantString)
			}
		})
	}
}

func TestDateModelScanString(t *testing.T) {
	var d timex.DateModel
	if err := d.Scan("2026-03-23"); err != nil {
		t.Fatalf("Scan(string) error = %v", err)
	}
	if d.String() != "2026-03-23" {
		t.Errorf("got %q, want %q", d.String(), "2026-03-23")
	}
}

func TestDateModelScanFullTimeString(t *testing.T) {
	var d timex.DateModel
	if err := d.Scan("2026-03-23 14:30:00"); err != nil {
		t.Fatalf("Scan(full time string) error = %v", err)
	}
	if d.String() != "2026-03-23" {
		t.Errorf("got %q, want %q", d.String(), "2026-03-23")
	}
}

func TestDateModelScanBytes(t *testing.T) {
	var d timex.DateModel
	if err := d.Scan([]byte("2026-03-23")); err != nil {
		t.Fatalf("Scan([]byte) error = %v", err)
	}
	if d.String() != "2026-03-23" {
		t.Errorf("got %q, want %q", d.String(), "2026-03-23")
	}
}

func TestDateModelScanNil(t *testing.T) {
	var d timex.DateModel
	if err := d.Scan(nil); err != nil {
		t.Fatalf("Scan(nil) error = %v", err)
	}
	if !d.Time().IsZero() {
		t.Error("Scan(nil) should produce zero time")
	}
}

func TestDateModelScanUnsupportedType(t *testing.T) {
	var d timex.DateModel
	if err := d.Scan(12345); err == nil {
		t.Error("Scan(int) should return error")
	}
}

func TestDateModelScanInvalidString(t *testing.T) {
	var d timex.DateModel
	if err := d.Scan("not-a-date"); err == nil {
		t.Error("Scan(invalid string) should return error")
	}
}

// ==================== MarshalJSON ====================

func TestDateModelMarshalJSON(t *testing.T) {
	shanghai, _ := time.LoadLocation("Asia/Shanghai")

	tests := []struct {
		name     string
		input    time.Time
		wantJSON string
	}{
		{
			"正常日期",
			time.Date(2026, 3, 23, 0, 0, 0, 0, shanghai),
			`"2026-03-23"`,
		},
		{
			"带时分秒（只输出日期）",
			time.Date(2026, 3, 23, 14, 30, 0, 0, shanghai),
			`"2026-03-23"`,
		},
		{
			"零值 → 空字符串",
			time.Time{},
			`""`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := timex.DateModel(tt.input)
			data, err := json.Marshal(d)
			if err != nil {
				t.Fatalf("MarshalJSON() error = %v", err)
			}
			if string(data) != tt.wantJSON {
				t.Errorf("MarshalJSON() = %s, want %s", data, tt.wantJSON)
			}
		})
	}
}

// ==================== UnmarshalJSON ====================

func TestDateModelUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantString string
		wantZero   bool
	}{
		{"纯日期", `"2026-03-23"`, "2026-03-23", false},
		{"null", `null`, "", true},
		{"空字符串", `""`, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var d timex.DateModel
			if err := json.Unmarshal([]byte(tt.input), &d); err != nil {
				t.Fatalf("UnmarshalJSON(%s) error = %v", tt.input, err)
			}
			if tt.wantZero {
				if !d.Time().IsZero() {
					t.Error("expected zero time")
				}
			} else if d.String() != tt.wantString {
				t.Errorf("got %q, want %q", d.String(), tt.wantString)
			}
		})
	}
}

func TestDateModelUnmarshalInvalidFormat(t *testing.T) {
	var d timex.DateModel
	if err := json.Unmarshal([]byte(`"2026-03-23 14:30:00"`), &d); err == nil {
		t.Error("UnmarshalJSON should reject full datetime format")
	}
}

func TestDateModelUnmarshalInvalid(t *testing.T) {
	var d timex.DateModel
	if err := json.Unmarshal([]byte(`"not-a-date"`), &d); err == nil {
		t.Error("expected error for invalid date")
	}
}

// ==================== Value (driver.Valuer) ====================

func TestDateModelValue(t *testing.T) {
	shanghai, _ := time.LoadLocation("Asia/Shanghai")

	t.Run("正常日期 → time.Time", func(t *testing.T) {
		d := timex.DateModel(time.Date(2026, 3, 23, 0, 0, 0, 0, shanghai))
		v, err := d.Value()
		if err != nil {
			t.Fatalf("Value() error = %v", err)
		}
		if v == nil {
			t.Fatal("Value() should not be nil for non-zero date")
		}
	})

	t.Run("零值 → nil", func(t *testing.T) {
		d := timex.DateModel(time.Time{})
		v, err := d.Value()
		if err != nil {
			t.Fatalf("Value() error = %v", err)
		}
		if v != nil {
			t.Error("Value() should be nil for zero date")
		}
	})
}

// ==================== Scan + MarshalJSON 联动 ====================

func TestDateModelScanThenMarshal(t *testing.T) {
	shanghai, _ := time.LoadLocation("Asia/Shanghai")
	dbTime := time.Date(2026, 3, 23, 0, 0, 0, 0, shanghai)

	var d timex.DateModel
	if err := d.Scan(dbTime); err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	data, err := json.Marshal(d)
	if err != nil {
		t.Fatalf("MarshalJSON() error = %v", err)
	}

	want := `"2026-03-23"`
	if string(data) != want {
		t.Errorf("DB→Scan→JSON = %s, want %s", data, want)
	}
}

// ==================== String ====================

func TestDateModelString(t *testing.T) {
	shanghai, _ := time.LoadLocation("Asia/Shanghai")
	d := timex.DateModel(time.Date(2026, 6, 15, 9, 0, 0, 0, shanghai))
	want := "2026-06-15"
	if d.String() != want {
		t.Errorf("String() = %q, want %q", d.String(), want)
	}
}

// ==================== Time ====================

func TestDateModelTime(t *testing.T) {
	original := time.Date(2026, 3, 23, 0, 0, 0, 0, time.UTC)
	d := timex.DateModel(original)
	if !d.Time().Equal(original) {
		t.Error("Time() should return the underlying time.Time")
	}
}
