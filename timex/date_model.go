package timex

import (
	"database/sql/driver"
	"fmt"
	"time"
)

// DateModel 纯日期类型，JSON/String 输出 YYYY-MM-DD，DB 存储为 DATE
type DateModel time.Time

// MarshalJSON 实现 JSON 序列化
func (d DateModel) MarshalJSON() ([]byte, error) {
	if time.Time(d).IsZero() {
		return []byte(`""`), nil
	}
	formatted := time.Time(d).In(ShanghaiLocation).Format(DateFormat)
	return []byte(`"` + formatted + `"`), nil
}

// UnmarshalJSON 实现 JSON 反序列化
func (d *DateModel) UnmarshalJSON(data []byte) error {
	str := string(data)
	if str == "null" || str == `""` {
		*d = DateModel(time.Time{})
		return nil
	}

	if len(str) >= 2 && str[0] == '"' && str[len(str)-1] == '"' {
		str = str[1 : len(str)-1]
	}

	t, err := time.ParseInLocation(DateFormat, str, ShanghaiLocation)
	if err != nil {
		return err
	}
	*d = DateModel(t)
	return nil
}

// Value 实现 driver.Valuer 接口，用于数据库写入
func (d DateModel) Value() (driver.Value, error) {
	t := time.Time(d)
	if t.IsZero() {
		return nil, nil
	}
	return t, nil
}

// Scan 实现 sql.Scanner 接口，用于数据库读取
func (d *DateModel) Scan(value any) error {
	if value == nil {
		*d = DateModel(time.Time{})
		return nil
	}

	switch v := value.(type) {
	case time.Time:
		*d = DateModel(v)
		return nil
	case []byte:
		t, err := time.ParseInLocation(DateFormat, string(v), ShanghaiLocation)
		if err != nil {
			// 兼容 TIMESTAMP 类型扫描（部分驱动返回完整时间字符串）
			t, err = time.ParseInLocation(TimeFormat, string(v), ShanghaiLocation)
			if err != nil {
				return err
			}
		}
		*d = DateModel(t)
		return nil
	case string:
		t, err := time.ParseInLocation(DateFormat, v, ShanghaiLocation)
		if err != nil {
			t, err = time.ParseInLocation(TimeFormat, v, ShanghaiLocation)
			if err != nil {
				return err
			}
		}
		*d = DateModel(t)
		return nil
	default:
		return fmt.Errorf("cannot scan type %T into DateModel", value)
	}
}

// Time 转换为标准 time.Time
func (d DateModel) Time() time.Time {
	return time.Time(d)
}

// String 实现 Stringer 接口
func (d DateModel) String() string {
	return time.Time(d).In(ShanghaiLocation).Format(DateFormat)
}
