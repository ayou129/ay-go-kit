package money

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/shopspring/decimal"
)

var (
	amountStrictPattern = regexp.MustCompile(`^-?\d+(\.\d{1,2})?$`)
	defaultScale        = int32(8)
)

// ResidualMode 控制分账后的舍入差额处理方式
type ResidualMode string

const (
	// ResidualModeExtract 差额单独返回，不补到任何一方
	ResidualModeExtract ResidualMode = "extract"
	// ResidualModeAllocateToUser 差额补到用户侧
	ResidualModeAllocateToUser ResidualMode = "allocate_to_user"
	// ResidualModeAllocateToPlatform 差额补到平台侧
	ResidualModeAllocateToPlatform ResidualMode = "allocate_to_platform"
)

// ParseAmount 校验最多两位小数的字符串（如 "123.45"），转为 decimal.Decimal
func ParseAmount(raw string) (decimal.Decimal, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return decimal.Zero, errors.New("amount is empty")
	}
	if !amountStrictPattern.MatchString(raw) {
		return decimal.Zero, fmt.Errorf("amount %q must have at most two decimal places", raw)
	}
	val, err := decimal.NewFromString(raw)
	if err != nil {
		return decimal.Zero, fmt.Errorf("parse amount %q: %w", raw, err)
	}
	return val, nil
}

// MustParseAmount 与 ParseAmount 相同，但遇到错误会 panic，只适用于初始化阶段
func MustParseAmount(raw string) decimal.Decimal {
	val, err := ParseAmount(raw)
	if err != nil {
		panic(err)
	}
	return val
}

// FormatAmount 使用银行家舍入保留两位小数，以字符串格式返回
func FormatAmount(val decimal.Decimal) string {
	return val.RoundBank(2).StringFixed(2)
}

// RoundToCent 四舍五入保留两位小数
func RoundToCent(val decimal.Decimal) decimal.Decimal {
	return val.Round(2)
}

// RoundUpToCent 向上取整至两位小数，常用于手续费
func RoundUpToCent(val decimal.Decimal) decimal.Decimal {
	return val.RoundCeil(2)
}

// RoundDownToCent 向零截断到两位小数
func RoundDownToCent(val decimal.Decimal) decimal.Decimal {
	return val.RoundFloor(2)
}

// AddAmounts 返回 a + b
func AddAmounts(a, b decimal.Decimal) decimal.Decimal {
	return a.Add(b)
}

// SubtractAmounts 返回 a - b
func SubtractAmounts(a, b decimal.Decimal) decimal.Decimal {
	return a.Sub(b)
}

// MultiplyAmount 计算 a * b，结果保持内部高精度（8 位小数）
func MultiplyAmount(a, b decimal.Decimal) decimal.Decimal {
	return a.Mul(b).Round(defaultScale)
}

// DivideAmount 计算 numerator / denominator，通过 roundingRule 控制两位小数的处理方式；
// 分母为 0 时返回错误。roundingRule 为 nil 时默认四舍五入。
func DivideAmount(numerator, denominator decimal.Decimal, roundingRule func(decimal.Decimal) decimal.Decimal) (decimal.Decimal, error) {
	if denominator.IsZero() {
		return decimal.Zero, errors.New("divide amount: denominator is zero")
	}
	quotient := numerator.DivRound(denominator, defaultScale)
	if roundingRule == nil {
		roundingRule = RoundToCent
	}
	quotient = roundingRule(quotient)
	return quotient, nil
}

// ToMinorUnits 将金额转换为最小货币单位（如分），四舍五入返回 int64
func ToMinorUnits(val decimal.Decimal) int64 {
	return val.Mul(decimal.NewFromInt(100)).Round(0).IntPart()
}

// FromMinorUnits 将最小货币单位（如分）转换回金额 decimal
func FromMinorUnits(units int64) decimal.Decimal {
	return decimal.NewFromInt(units).Div(decimal.NewFromInt(100))
}

// SettleUserPlatform 根据两方的高精度金额计算最终结算金额，返回舍入差额。
// mode 控制差额分配：extract=单独返回 / allocate_to_user=补到用户 / allocate_to_platform=补到平台
func SettleUserPlatform(userShare, platformShare decimal.Decimal, mode ResidualMode) (decimal.Decimal, decimal.Decimal, decimal.Decimal) {
	userRounded := RoundToCent(userShare)
	platformRounded := RoundToCent(platformShare)
	totalRounded := RoundToCent(userShare.Add(platformShare))
	sumRounded := userRounded.Add(platformRounded)
	residual := totalRounded.Sub(sumRounded)

	if residual.IsZero() {
		return userRounded, platformRounded, decimal.Zero
	}

	switch mode {
	case ResidualModeAllocateToUser:
		userRounded = RoundToCent(userRounded.Add(residual))
		residual = decimal.Zero
	case ResidualModeAllocateToPlatform:
		platformRounded = RoundToCent(platformRounded.Add(residual))
		residual = decimal.Zero
	case ResidualModeExtract:
		// 保留差额给调用方
	default:
		// 未知模式默认当作 extract 处理
	}

	return userRounded, platformRounded, residual
}
