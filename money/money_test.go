package money

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestParseAmount_Valid(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"123.45", "123.45"},
		{"0.01", "0.01"},
		{"100", "100.00"},
		{"-50.99", "-50.99"},
		{" 12.34 ", "12.34"},
	}
	for _, tt := range tests {
		val, err := ParseAmount(tt.input)
		if err != nil {
			t.Errorf("ParseAmount(%q) error: %v", tt.input, err)
			continue
		}
		if val.StringFixed(2) != tt.want {
			t.Errorf("ParseAmount(%q) = %s, want %s", tt.input, val.StringFixed(2), tt.want)
		}
	}
}

func TestParseAmount_Invalid(t *testing.T) {
	invalids := []string{"", "abc", "12.345", "12.3.4"}
	for _, input := range invalids {
		_, err := ParseAmount(input)
		if err == nil {
			t.Errorf("ParseAmount(%q) should return error", input)
		}
	}
}

func TestMustParseAmount_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustParseAmount should panic on invalid input")
		}
	}()
	MustParseAmount("invalid")
}

func TestFormatAmount(t *testing.T) {
	tests := []struct {
		input decimal.Decimal
		want  string
	}{
		{decimal.NewFromFloat(1.005), "1.00"}, // 银行家舍入
		{decimal.NewFromFloat(1.015), "1.02"}, // 银行家舍入
		{decimal.NewFromFloat(100.0), "100.00"},
	}
	for _, tt := range tests {
		got := FormatAmount(tt.input)
		if got != tt.want {
			t.Errorf("FormatAmount(%s) = %s, want %s", tt.input, got, tt.want)
		}
	}
}

func TestRoundVariants(t *testing.T) {
	val := decimal.NewFromFloat(1.235)

	if got := RoundToCent(val).StringFixed(2); got != "1.24" {
		t.Errorf("RoundToCent(1.235) = %s, want 1.24", got)
	}
	if got := RoundUpToCent(val).StringFixed(2); got != "1.24" {
		t.Errorf("RoundUpToCent(1.235) = %s, want 1.24", got)
	}
	if got := RoundDownToCent(val).StringFixed(2); got != "1.23" {
		t.Errorf("RoundDownToCent(1.235) = %s, want 1.23", got)
	}
}

func TestArithmetic(t *testing.T) {
	a := decimal.NewFromFloat(10.50)
	b := decimal.NewFromFloat(3.25)

	if got := AddAmounts(a, b).StringFixed(2); got != "13.75" {
		t.Errorf("Add = %s, want 13.75", got)
	}
	if got := SubtractAmounts(a, b).StringFixed(2); got != "7.25" {
		t.Errorf("Subtract = %s, want 7.25", got)
	}
}

func TestMultiplyAmount(t *testing.T) {
	a := decimal.NewFromFloat(10.00)
	b := decimal.NewFromFloat(0.15)
	got := MultiplyAmount(a, b)
	if got.StringFixed(2) != "1.50" {
		t.Errorf("Multiply = %s, want 1.50", got.StringFixed(2))
	}
}

func TestDivideAmount_Normal(t *testing.T) {
	num := decimal.NewFromFloat(10.00)
	den := decimal.NewFromFloat(3.00)
	got, err := DivideAmount(num, den, nil)
	if err != nil {
		t.Fatalf("DivideAmount error: %v", err)
	}
	if got.StringFixed(2) != "3.33" {
		t.Errorf("Divide = %s, want 3.33", got.StringFixed(2))
	}
}

func TestDivideAmount_DivideByZero(t *testing.T) {
	_, err := DivideAmount(decimal.NewFromInt(10), decimal.Zero, nil)
	if err == nil {
		t.Error("DivideAmount by zero should return error")
	}
}

func TestDivideAmount_CustomRounding(t *testing.T) {
	num := decimal.NewFromFloat(10.00)
	den := decimal.NewFromFloat(3.00)
	got, err := DivideAmount(num, den, RoundUpToCent)
	if err != nil {
		t.Fatalf("DivideAmount error: %v", err)
	}
	if got.StringFixed(2) != "3.34" {
		t.Errorf("Divide with RoundUp = %s, want 3.34", got.StringFixed(2))
	}
}

func TestToMinorUnits_Roundtrip(t *testing.T) {
	val := decimal.NewFromFloat(12.34)
	units := ToMinorUnits(val)
	if units != 1234 {
		t.Errorf("ToMinorUnits(12.34) = %d, want 1234", units)
	}
	back := FromMinorUnits(units)
	if !back.Equal(val) {
		t.Errorf("FromMinorUnits(1234) = %s, want 12.34", back)
	}
}

func TestSettleUserPlatform_NoResidual(t *testing.T) {
	user := decimal.NewFromFloat(7.00)
	platform := decimal.NewFromFloat(3.00)
	u, p, r := SettleUserPlatform(user, platform, ResidualModeExtract)
	if !r.IsZero() {
		t.Errorf("residual = %s, want 0", r)
	}
	if u.StringFixed(2) != "7.00" || p.StringFixed(2) != "3.00" {
		t.Errorf("user=%s platform=%s, want 7.00 and 3.00", u, p)
	}
}

func TestSettleUserPlatform_AllocateToUser(t *testing.T) {
	// 构造一个有舍入差额的场景
	user := decimal.NewFromFloat(3.333)
	platform := decimal.NewFromFloat(6.667)
	u, p, r := SettleUserPlatform(user, platform, ResidualModeAllocateToUser)
	if !r.IsZero() {
		t.Errorf("residual should be 0 when allocated, got %s", r)
	}
	total := u.Add(p)
	expected := RoundToCent(user.Add(platform))
	if !total.Equal(expected) {
		t.Errorf("user+platform=%s, want %s", total, expected)
	}
}

func TestSettleUserPlatform_AllocateToPlatform(t *testing.T) {
	user := decimal.NewFromFloat(3.333)
	platform := decimal.NewFromFloat(6.667)
	u, p, r := SettleUserPlatform(user, platform, ResidualModeAllocateToPlatform)
	if !r.IsZero() {
		t.Errorf("residual should be 0 when allocated, got %s", r)
	}
	total := u.Add(p)
	expected := RoundToCent(user.Add(platform))
	if !total.Equal(expected) {
		t.Errorf("user+platform=%s, want %s", total, expected)
	}
}

func TestSettleUserPlatform_Extract(t *testing.T) {
	// 构造有差额的场景，extract 模式下 residual 应非零
	user := decimal.NewFromFloat(3.335)
	platform := decimal.NewFromFloat(6.669)
	u, p, r := SettleUserPlatform(user, platform, ResidualModeExtract)
	total := u.Add(p).Add(r)
	expected := RoundToCent(user.Add(platform))
	if !total.Equal(expected) {
		t.Errorf("user+platform+residual=%s, want %s", total, expected)
	}
}

func TestSettleUserPlatform_UnknownMode(t *testing.T) {
	// 未知 mode 默认当 extract 处理
	user := decimal.NewFromFloat(3.335)
	platform := decimal.NewFromFloat(6.669)
	u1, p1, r1 := SettleUserPlatform(user, platform, ResidualModeExtract)
	u2, p2, r2 := SettleUserPlatform(user, platform, "unknown_mode")
	if !u1.Equal(u2) || !p1.Equal(p2) || !r1.Equal(r2) {
		t.Errorf("unknown mode should behave like extract")
	}
}

func TestSettleUserPlatform_NegativeAmounts(t *testing.T) {
	user := decimal.NewFromFloat(-3.333)
	platform := decimal.NewFromFloat(-6.667)
	u, p, _ := SettleUserPlatform(user, platform, ResidualModeExtract)
	if u.IsPositive() || p.IsPositive() {
		t.Error("negative inputs should produce negative outputs")
	}
}
