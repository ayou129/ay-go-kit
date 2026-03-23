package timex_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/ay/go-kit/timex"
)

// ==================== Scan ====================

func TestTimeModelScanTime(t *testing.T) {
	shanghai, _ := time.LoadLocation("Asia/Shanghai")

	tests := []struct {
		name       string
		input      time.Time
		wantString string
	}{
		{
			"上海时区时间（pgx ScanLocation 生效后的实际输入）",
			time.Date(2026, 3, 3, 14, 30, 0, 0, shanghai),
			"2026-03-03 14:30:00",
		},
		{
			"跨日时间",
			time.Date(2026, 12, 31, 23, 59, 59, 0, shanghai),
			"2026-12-31 23:59:59",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ct timex.TimeModel
			if err := ct.Scan(tt.input); err != nil {
				t.Fatalf("Scan() error = %v", err)
			}
			got := ct.String()
			if got != tt.wantString {
				t.Errorf("Scan(%v).String() = %q, want %q", tt.input, got, tt.wantString)
			}
		})
	}
}

func TestTimeModelScanString(t *testing.T) {
	var ct timex.TimeModel
	if err := ct.Scan("2026-03-03 14:30:00"); err != nil {
		t.Fatalf("Scan(string) error = %v", err)
	}
	if ct.String() != "2026-03-03 14:30:00" {
		t.Errorf("got %q, want %q", ct.String(), "2026-03-03 14:30:00")
	}
}

func TestTimeModelScanBytes(t *testing.T) {
	var ct timex.TimeModel
	if err := ct.Scan([]byte("2026-03-03 14:30:00")); err != nil {
		t.Fatalf("Scan([]byte) error = %v", err)
	}
	if ct.String() != "2026-03-03 14:30:00" {
		t.Errorf("got %q, want %q", ct.String(), "2026-03-03 14:30:00")
	}
}

func TestTimeModelScanNil(t *testing.T) {
	var ct timex.TimeModel
	if err := ct.Scan(nil); err != nil {
		t.Fatalf("Scan(nil) error = %v", err)
	}
	if !ct.Time().IsZero() {
		t.Error("Scan(nil) should produce zero time")
	}
}

func TestTimeModelScanUnsupportedType(t *testing.T) {
	var ct timex.TimeModel
	if err := ct.Scan(12345); err == nil {
		t.Error("Scan(int) should return error")
	}
}

func TestTimeModelScanInvalidString(t *testing.T) {
	var ct timex.TimeModel
	if err := ct.Scan("not-a-date"); err == nil {
		t.Error("Scan(invalid string) should return error")
	}
}

// ==================== MarshalJSON ====================

func TestTimeModelMarshalJSON(t *testing.T) {
	shanghai, _ := time.LoadLocation("Asia/Shanghai")

	tests := []struct {
		name     string
		input    time.Time
		wantJSON string
	}{
		{
			"正常时间",
			time.Date(2026, 3, 3, 14, 30, 0, 0, shanghai),
			`"2026-03-03 14:30:00"`,
		},
		{
			"零值 → 空字符串",
			time.Time{},
			`""`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ct := timex.TimeModel(tt.input)
			data, err := json.Marshal(ct)
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

func TestTimeModelUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantString string
		wantZero   bool
	}{
		{"完整时间", `"2026-03-03 14:30:00"`, "2026-03-03 14:30:00", false},
		{"纯日期", `"2026-03-03"`, "2026-03-03 00:00:00", false},
		{"null", `null`, "", true},
		{"空字符串", `""`, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ct timex.TimeModel
			if err := json.Unmarshal([]byte(tt.input), &ct); err != nil {
				t.Fatalf("UnmarshalJSON(%s) error = %v", tt.input, err)
			}
			if tt.wantZero {
				if !ct.Time().IsZero() {
					t.Error("expected zero time")
				}
			} else if ct.String() != tt.wantString {
				t.Errorf("got %q, want %q", ct.String(), tt.wantString)
			}
		})
	}
}

func TestTimeModelUnmarshalInvalid(t *testing.T) {
	var ct timex.TimeModel
	if err := json.Unmarshal([]byte(`"not-a-date"`), &ct); err == nil {
		t.Error("expected error for invalid date")
	}
}

// ==================== Value (driver.Valuer) ====================

func TestTimeModelValue(t *testing.T) {
	shanghai, _ := time.LoadLocation("Asia/Shanghai")

	t.Run("正常时间 → time.Time", func(t *testing.T) {
		ct := timex.TimeModel(time.Date(2026, 3, 3, 14, 30, 0, 0, shanghai))
		v, err := ct.Value()
		if err != nil {
			t.Fatalf("Value() error = %v", err)
		}
		if v == nil {
			t.Fatal("Value() should not be nil for non-zero time")
		}
	})

	t.Run("零值 → nil", func(t *testing.T) {
		ct := timex.TimeModel(time.Time{})
		v, err := ct.Value()
		if err != nil {
			t.Fatalf("Value() error = %v", err)
		}
		if v != nil {
			t.Error("Value() should be nil for zero time")
		}
	})
}

// ==================== Scan + MarshalJSON 联动（模拟 DB→JSON 链路） ====================

func TestTimeModelScanThenMarshal(t *testing.T) {
	shanghai, _ := time.LoadLocation("Asia/Shanghai")
	dbTime := time.Date(2026, 3, 3, 14, 30, 0, 0, shanghai)

	var ct timex.TimeModel
	if err := ct.Scan(dbTime); err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	data, err := json.Marshal(ct)
	if err != nil {
		t.Fatalf("MarshalJSON() error = %v", err)
	}

	want := `"2026-03-03 14:30:00"`
	if string(data) != want {
		t.Errorf("DB→Scan→JSON = %s, want %s", data, want)
	}
}

// ==================== String ====================

func TestTimeModelString(t *testing.T) {
	shanghai, _ := time.LoadLocation("Asia/Shanghai")
	ct := timex.TimeModel(time.Date(2026, 6, 15, 9, 0, 0, 0, shanghai))
	want := "2026-06-15 09:00:00"
	if ct.String() != want {
		t.Errorf("String() = %q, want %q", ct.String(), want)
	}
}

// ==================== Time ====================

func TestTimeModelTime(t *testing.T) {
	original := time.Date(2026, 3, 3, 14, 30, 0, 0, time.UTC)
	ct := timex.TimeModel(original)
	if !ct.Time().Equal(original) {
		t.Error("Time() should return the underlying time.Time")
	}
}
