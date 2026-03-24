package ginx

import (
	"reflect"
	"strings"
	"unicode"

	"github.com/gin-gonic/gin"
)

// BindJSON 绑定 JSON 请求体并自动清洗所有 string 字段。
// 替代 c.ShouldBindJSON(&req)，绑定成功后自动调用 SanitizeStrings。
func BindJSON(c *gin.Context, obj any) error {
	if err := c.ShouldBindJSON(obj); err != nil {
		return err
	}
	SanitizeStrings(obj)
	return nil
}

// BindJSONMap 绑定 JSON 请求体到 map 并自动清洗所有 string 值。
// 替代 c.ShouldBindJSON(&updates) + SanitizeMap(updates)。
func BindJSONMap(c *gin.Context, m *map[string]any) error {
	if err := c.ShouldBindJSON(m); err != nil {
		return err
	}
	SanitizeMap(*m)
	return nil
}

// SanitizeMap 清洗 map[string]any 中所有 string 类型的 value
func SanitizeMap(m map[string]any) {
	for k, v := range m {
		if s, ok := v.(string); ok {
			m[k] = CleanString(s)
		}
	}
}

// SanitizeStrings 递归清洗 struct 中所有 string 字段：
//   - 去除零宽字符和不可见控制字符
//   - trim 首尾空白
//   - 合并连续空格为单个
//
// 支持嵌套 struct 和指针字段，跳过 nil 指针。
// 通过 sanitize tag 控制行为：
//   - sanitize:"-"         跳过清洗（密码等敏感字段）
//   - sanitize:"multiline" 保留换行符，按行清洗（Markdown / 富文本）
func SanitizeStrings(v any) {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return
	}
	sanitizeStruct(rv)
}

func sanitizeStruct(rv reflect.Value) {
	rt := rv.Type()
	for i := 0; i < rv.NumField(); i++ {
		field := rv.Field(i)
		structField := rt.Field(i)

		// 跳过未导出字段
		if !structField.IsExported() {
			continue
		}

		tag := structField.Tag.Get("sanitize")
		if tag == "-" {
			continue
		}

		switch field.Kind() {
		case reflect.String:
			if field.CanSet() {
				if tag == "multiline" {
					field.SetString(CleanMultilineString(field.String()))
				} else {
					field.SetString(CleanString(field.String()))
				}
			}
		case reflect.Struct:
			sanitizeStruct(field)
		case reflect.Ptr:
			if !field.IsNil() && field.Elem().Kind() == reflect.Struct {
				sanitizeStruct(field.Elem())
			}
		case reflect.Interface:
			// 处理 any 类型字段（如 Value any），底层是 string 时清洗
			if !field.IsNil() && field.Elem().Kind() == reflect.String && field.CanSet() {
				field.Set(reflect.ValueOf(CleanString(field.Elem().String())))
			}
		}
	}
}

// CleanString 清洗单个字符串：
//  1. 去除零宽字符（U+200B, U+200C, U+200D, U+FEFF, U+00AD 等）
//  2. 去除不可见控制字符（保留普通空格 U+0020）
//  3. 全角空格转半角空格，tab 转空格
//  4. trim 首尾空白
//  5. 合并连续空格为单个
func CleanString(s string) string {
	if s == "" {
		return s
	}

	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if isZeroWidth(r) {
			continue
		}
		if unicode.IsControl(r) && r != '\t' {
			continue
		}
		if r == '\u3000' {
			b.WriteByte(' ')
			continue
		}
		if r == '\t' {
			b.WriteByte(' ')
			continue
		}
		b.WriteRune(r)
	}

	result := strings.TrimSpace(b.String())
	result = collapseSpaces(result)
	return result
}

// CleanMultilineString 清洗多行字符串，保留换行符结构：
//  1. 按 \n 分行，每行执行 CleanString 清洗
//  2. 去除末尾连续空行
//
// 适用于 Markdown、富文本等需要保留换行的场景。
func CleanMultilineString(s string) string {
	if s == "" {
		return s
	}
	// 统一换行符为 \n
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")

	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = CleanString(line)
	}

	// 去除末尾连续空行
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	return strings.Join(lines, "\n")
}

func isZeroWidth(r rune) bool {
	switch r {
	case '\u200B', // zero width space
		'\u200C', // zero width non-joiner
		'\u200D', // zero width joiner
		'\uFEFF', // BOM / zero width no-break space
		'\u00AD', // soft hyphen
		'\u200E', // left-to-right mark
		'\u200F', // right-to-left mark
		'\u2060', // word joiner
		'\u2061', // function application
		'\u2062', // invisible times
		'\u2063', // invisible separator
		'\u2064': // invisible plus
		return true
	}
	return false
}

func collapseSpaces(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	prevSpace := false
	for _, r := range s {
		if r == ' ' {
			if prevSpace {
				continue
			}
			prevSpace = true
		} else {
			prevSpace = false
		}
		b.WriteRune(r)
	}
	return b.String()
}
