# go-kit

项目共享基础设施库。

## 模块

| 包 | 说明 |
|---|---|
| `auth` | 鉴权工具 |
| `ctxutil` | Context 工具函数 |
| `database` | PostgreSQL 连接管理（GORM + pgx，时区 Asia/Shanghai） |
| `ginx` | Gin 扩展（错误处理、响应封装） |
| `i18n` | 国际化基础设施 |
| `logger` | 结构化日志 |
| `rediscli` | Redis 客户端封装 |
| `timex` | 时间类型（TimeModel）+ 常量（TimeFormat / DateFormat）+ 上海时区 |
| `token` | Token 管理 |

## timex

自定义时间类型，统一 JSON 序列化（`"2006-01-02 15:04:05"`）和数据库读写行为。

```go
import "github.com/ay/go-kit/timex"

// 类型
var t timex.TimeModel

// 常量
timex.TimeFormat       // "2006-01-02 15:04:05"
timex.DateFormat       // "2006-01-02"
timex.ShanghaiLocation // *time.Location (Asia/Shanghai)
```

零值处理：JSON 输出 `""`，DB 输出 `nil`。
