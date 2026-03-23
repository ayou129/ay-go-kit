package timex

import (
	"database/sql/driver"
	"fmt"
	"time"
)

// 统一时间格式
const TimeFormat = "2006-01-02 15:04:05"

// 日期格式（用于文件名等场景）
const DateFormat = "2006-01-02"

// ShanghaiLocation 上海时区
var ShanghaiLocation *time.Location

func init() {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		ShanghaiLocation = time.FixedZone("CST", 8*3600)
	} else {
		ShanghaiLocation = loc
	}
}

// TimeModel 自定义时间类型，统一 JSON 序列化格式
type TimeModel time.Time

// MarshalJSON 实现 JSON 序列化
func (ct TimeModel) MarshalJSON() ([]byte, error) {
	if time.Time(ct).IsZero() {
		return []byte(`""`), nil
	}
	formatted := time.Time(ct).In(ShanghaiLocation).Format(TimeFormat)
	return []byte(`"` + formatted + `"`), nil
}

// UnmarshalJSON 实现 JSON 反序列化
func (ct *TimeModel) UnmarshalJSON(data []byte) error {
	// 移除引号
	str := string(data)
	if str == "null" || str == `""` {
		*ct = TimeModel(time.Time{})
		return nil
	}

	// 去掉前后引号
	if len(str) >= 2 && str[0] == '"' && str[len(str)-1] == '"' {
		str = str[1 : len(str)-1]
	}

	// 优先尝试完整格式 YYYY-MM-DD HH:mm:ss，不匹配则尝试纯日期 YYYY-MM-DD
	t, err := time.ParseInLocation(TimeFormat, str, ShanghaiLocation)
	if err != nil {
		t, err = time.ParseInLocation(DateFormat, str, ShanghaiLocation)
		if err != nil {
			return err
		}
	}
	*ct = TimeModel(t)
	return nil
}

// Value 实现 driver.Valuer 接口，用于数据库写入
func (ct TimeModel) Value() (driver.Value, error) {
	t := time.Time(ct)
	if t.IsZero() {
		return nil, nil
	}
	return t, nil
}

// Scan 实现 sql.Scanner 接口，用于数据库读取
func (ct *TimeModel) Scan(value any) error {
	if value == nil {
		*ct = TimeModel(time.Time{})
		return nil
	}

	switch v := value.(type) {
	case time.Time:
		*ct = TimeModel(v)
		return nil
	case []byte:
		t, err := time.ParseInLocation(TimeFormat, string(v), ShanghaiLocation)
		if err != nil {
			return err
		}
		*ct = TimeModel(t)
		return nil
	case string:
		t, err := time.ParseInLocation(TimeFormat, v, ShanghaiLocation)
		if err != nil {
			return err
		}
		*ct = TimeModel(t)
		return nil
	default:
		return fmt.Errorf("cannot scan type %T into TimeModel", value)
	}
}

// Time 转换为标准 time.Time
func (ct TimeModel) Time() time.Time {
	return time.Time(ct)
}

// String 实现 Stringer 接口
func (ct TimeModel) String() string {
	return time.Time(ct).In(ShanghaiLocation).Format(TimeFormat)
}
