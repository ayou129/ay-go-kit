package i18n

import (
	"fmt"
	"strings"
	"sync"
)

const (
	LangZh = "zh"
	LangEn = "en"
)

// Common error codes (system level, shared across all projects)
const (
	CodeSuccess       = 0
	CodeWelcome       = 1
	CodeInternalError = 10001
	CodeParamInvalid  = 10002
	CodeParamFilterFieldNotAllowed = 10003
	CodeRouteNotFound = 10011
	CodeForbidden     = 10012
	CodeRateLimit     = 10013
	CodeMaintenance   = 10020

	CodeTokenInvalid       = 20001
	CodeTokenExpired       = 20002
	CodeTokenCreateFailed  = 20003
	CodeTokenRefreshFailed = 20004

	// Captcha 验证码
	CodeCaptchaIncorrect    = 30001
	CodeCaptchaTooManyTries = 30002
	CodeCaptchaDoneInvalid  = 30003

	// SMS 短信
	CodeSmsBlacklisted = 30101
	CodeSmsBusy        = 30102
	CodeSmsCodeError   = 30103
	CodeSmsVerifyLimit = 30104

	// User 用户（auth 的自然延伸）
	CodeUserNotFound     = 40001
	CodeLoginFailed      = 40002
	CodeAccountDisabled  = 40003
	CodeOldPasswordWrong = 40004

	// Data 通用数据校验
	CodeDataNotFound           = 40101
	CodeDataExists             = 40102
	CodeDataConflict           = 40103
	CodeCannotDeleteUsedRecord = 40104
)

// Catalog holds code->lang->message mappings
type Catalog struct {
	mu          sync.RWMutex
	messages    map[int]map[string]string
	defaultLang string
}

// defaultMessages are shared system-level messages
var defaultMessages = map[int]map[string]string{
	CodeSuccess:            {LangZh: "操作成功", LangEn: "Success"},
	CodeWelcome:            {LangZh: "欢迎", LangEn: "Welcome"},
	CodeInternalError:      {LangZh: "系统内部错误", LangEn: "Internal server error"},
	CodeParamInvalid:       {LangZh: "参数无效", LangEn: "Invalid parameter"},
	CodeParamFilterFieldNotAllowed: {LangZh: "不允许的筛选字段", LangEn: "Filter field not allowed"},
	CodeRouteNotFound:      {LangZh: "页面不存在", LangEn: "Page not found"},
	CodeForbidden:          {LangZh: "您还未登录", LangEn: "You are not logged in"},
	CodeRateLimit:          {LangZh: "请求过于频繁，请稍后再试", LangEn: "Too many requests, please try again later"},
	CodeMaintenance:        {LangZh: "系统维护中，请稍后再试", LangEn: "System maintenance, please try again later"},
	CodeTokenInvalid:       {LangZh: "Token 无效", LangEn: "Invalid token"},
	CodeTokenExpired:       {LangZh: "Token 已过期", LangEn: "Token expired"},
	CodeTokenCreateFailed:  {LangZh: "Token 创建失败", LangEn: "Failed to create token"},
	CodeTokenRefreshFailed: {LangZh: "Token 刷新失败", LangEn: "Failed to refresh token"},
	CodeCaptchaIncorrect:       {LangZh: "验证码错误", LangEn: "Incorrect captcha"},
	CodeCaptchaTooManyTries:    {LangZh: "验证码尝试次数过多", LangEn: "Too many captcha attempts"},
	CodeCaptchaDoneInvalid:     {LangZh: "验证码凭证已失效", LangEn: "Captcha credential expired"},
	CodeSmsBlacklisted:         {LangZh: "该手机号已被限制", LangEn: "Phone number is blacklisted"},
	CodeSmsBusy:                {LangZh: "短信发送过于频繁，请稍后再试", LangEn: "SMS sent too frequently, please try later"},
	CodeSmsCodeError:           {LangZh: "短信验证码错误", LangEn: "Incorrect SMS code"},
	CodeSmsVerifyLimit:         {LangZh: "短信验证码尝试次数过多", LangEn: "Too many SMS verification attempts"},
	CodeUserNotFound:           {LangZh: "用户不存在", LangEn: "User not found"},
	CodeLoginFailed:            {LangZh: "密码错误", LangEn: "Login failed"},
	CodeAccountDisabled:        {LangZh: "账号已被禁用", LangEn: "Account is disabled"},
	CodeOldPasswordWrong:       {LangZh: "旧密码错误", LangEn: "Old password is incorrect"},
	CodeDataNotFound:           {LangZh: "数据不存在", LangEn: "Data not found"},
	CodeDataExists:             {LangZh: "数据已存在", LangEn: "Data already exists"},
	CodeDataConflict:           {LangZh: "数据已被修改（%s），请刷新后重试", LangEn: "Data has been modified (%s), please refresh"},
	CodeCannotDeleteUsedRecord: {LangZh: "该记录已被引用，无法删除", LangEn: "Cannot delete a record that is in use"},
}

// NewCatalog creates a catalog with default system messages
func NewCatalog(defaultLang string) *Catalog {
	c := &Catalog{
		messages:    make(map[int]map[string]string),
		defaultLang: defaultLang,
	}
	for code, msgs := range defaultMessages {
		c.messages[code] = msgs
	}
	return c
}

// Register adds or overwrites messages for a code
func (c *Catalog) Register(code int, messages map[string]string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.messages[code] = messages
}

// RegisterBatch adds multiple code->messages mappings
func (c *Catalog) RegisterBatch(batch map[int]map[string]string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for code, msgs := range batch {
		c.messages[code] = msgs
	}
}

// GetDefaultLang returns the default language
func (c *Catalog) GetDefaultLang() string { return c.defaultLang }

// GetMsg returns localized message for code + lang
func (c *Catalog) GetMsg(code int, lang string, params ...any) string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if lang == "" {
		lang = c.defaultLang
	}

	messages, exists := c.messages[code]
	if !exists {
		return fmt.Sprintf("UNKNOWN_ERROR_%d", code)
	}

	message, exists := messages[lang]
	if !exists {
		message, exists = messages[c.defaultLang]
		if !exists {
			return fmt.Sprintf("UNKNOWN_ERROR_%d", code)
		}
	}

	if len(params) > 0 {
		cleaned := strings.ReplaceAll(message, "%%", "")
		if strings.Contains(cleaned, "%") {
			return fmt.Sprintf(message, params...)
		}
		if detail, ok := params[0].(string); ok && detail != "" {
			return message + "：" + detail
		}
	}

	return message
}

// Global convenience (projects set this once at startup)
var globalCatalog = NewCatalog(LangZh)

func SetGlobal(c *Catalog)   { globalCatalog = c }
func Global() *Catalog       { return globalCatalog }
func GetDefaultLang() string { return globalCatalog.GetDefaultLang() }
func GetLangMsg(code int, lang string, params ...any) string {
	return globalCatalog.GetMsg(code, lang, params...)
}
