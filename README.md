# go-kit

多项目共享基础设施库。

## 包一览

| 包 | 说明 |
|---|---|
| `auth` | Token 鉴权（创建/校验/刷新/删除），Redis + Lua 实现 |
| `cryptox` | 加解密工具（bcrypt 密码哈希） |
| `ctxutil` | Context 工具（UserID / TenantID / TraceID / Lang / Token） |
| `dbx` | PostgreSQL 连接 + 事务 + 分页 + 筛选 + JSONB 类型（GORM + pgx） |
| `ginx` | Gin 扩展（错误处理 / 响应封装 / 请求绑定清洗 / 分页查询 DTO） |
| `i18n` | 国际化（错误码 → 多语言文案） |
| `logger` | 结构化日志（TraceID 自动注入） |
| `money` | 高精度金额计算（shopspring/decimal） |
| `ratelimit` | 滑动窗口限流（Redis ZSET + Lua，含 Gin 中间件） |
| `rediscli` | Redis 客户端（连接管理 + Lua 脚本缓存） |
| `timex` | 时间类型（TimeModel / DateModel）+ 上海时区 |
| `token` | Token / TraceID / 随机字符串生成 |
| `treex` | 通用树形结构构建 |

---

## dbx

### 连接管理

```go
import "github.com/ay/go-kit/dbx"

// 启动时初始化
db, _ := dbx.Open(dbx.ConfigFromEnv())
dbx.SetGlobal(db)

// 任意位置获取（自动感知事务）
db := dbx.GetDB(ctx)
```

### 事务

事务内 `dbx.GetDB(txCtx)` 自动返回事务连接，repo 无感知。

**推荐：包函数直接调用**（适合集成测试 + txdb 的项目）

```go
func (s *orderService) Create(ctx context.Context, req CreateReq) error {
    return dbx.Transaction(ctx, dbx.GetDB(ctx), func(txCtx context.Context) error {
        // dbx.GetDB(txCtx) 自动返回 tx，repo 无感知
        return s.repo.Create(txCtx, &model.Order{...})
    })
}
```

**可选：TxRunner 接口注入**（适合需要 mock 事务的单元测试场景）

```go
type orderService struct {
    repo repository.OrderRepository
    tx   dbx.TxRunner
}

func (s *orderService) Create(ctx context.Context, req CreateReq) error {
    return s.tx.Transaction(ctx, func(txCtx context.Context) error {
        return s.repo.Create(txCtx, &model.Order{...})
    })
}
```

### 分页 — FindByPage

泛型分页查询，消灭 repo 层 Count + Offset + Limit + Find 样板代码。

```go
// 类型
type PageQuery struct {
    Page     int `json:"page"`
    PageSize int `json:"page_size"`
}

type PageResult[T any] struct {
    Total    int64 `json:"total"`
    Page     int   `json:"page"`
    PageSize int   `json:"page_size"`
    List     []T   `json:"list"`
}

// 使用：Service 层直接调用
func (s *roleService) ListPage(ctx context.Context, q ginx.PageQueryDTO) (*dbx.PageResult[model.Role], error) {
    scopes := q.ToScopes(ctx, nil)
    return dbx.FindByPage[model.Role](ctx, q.PageQuery, scopes...)
}
```

### 筛选排序 — ToScopes

将前端传来的 `FilterQuery` + `SortOption` 转为 GORM Scope。支持 14 种操作符。

```go
// 操作符：eq / neq / gt / gte / lt / lte / between / not_between
//         like / not_like / starts_with / ends_with / in / not_in
//         is_null / is_not_null

scopes := dbx.ToScopes(ctx, filters, sorts, allowedFields)
// allowedFields: nil=不限制, map[string]bool=白名单
```

### 错误判断

```go
if dbx.IsRecordNotFound(err) {
    // 记录不存在
}
```

### JSONB 类型

PostgreSQL JSONB 列的自动序列化/反序列化，实现 `driver.Valuer` + `sql.Scanner`。

```go
// model 中直接使用
type Order struct {
    AddressSnapshot dbx.JSONBObject  `gorm:"type:jsonb" json:"address_snapshot"`
    Tags            dbx.JSONBArrayStr `gorm:"type:jsonb" json:"tags"`
    Extra           dbx.JSONBArray   `gorm:"type:jsonb" json:"extra"`
}
```

- `JSONBObject` — `map[string]any`，存储 JSON 对象
- `JSONBArray` — `[]any`，存储任意类型数组
- `JSONBArrayStr` — `[]string`，存储字符串数组（无需类型断言）

### 测试

```go
// txdb 事务回滚隔离
restore := dbx.OverrideGetDB(txdbDB)
defer restore()
```

---

## ginx

### 请求绑定 + 字符串清洗

替代 `c.ShouldBindJSON`，自动清洗所有 string 字段（零宽字符 / 控制字符 / 全角空格 / 连续空格 / trim）。

```go
// struct 绑定
ginx.BindJSON(c, &req)

// map 绑定
ginx.BindJSONMap(c, &updates)

// 跳过清洗（密码等）
type LoginReq struct {
    Phone    string `json:"phone"`
    Password string `json:"password" sanitize:"-"`
}

// 保留换行（Markdown / 富文本）
type ProfileReq struct {
    Content string `json:"content" sanitize:"multiline"`
}
```

### 分页查询 DTO

封装 `dbx.PageQuery` + 筛选排序 + gin 绑定。

```go
var query ginx.PageQueryDTO
if err := query.Bind(c); err != nil { ... }
// 自动修正：page<1→1, pageSize<1→20, 默认排序 id desc

scopes := query.ToScopes(ctx, allowedFields)
result, err := dbx.FindByPage[model.Role](ctx, query.PageQuery, scopes...)
```

### 响应

```go
ginx.Success(c, data)           // {code:0, msg:"成功", data: ...}
ginx.SuccessWithMsg(c, msgCode, data) // code 永远为 0，msgCode 仅用于查 i18n 消息
```

### 错误

```go
ginx.NewError(i18n.CodeXxx)     // 业务错误 → 400
ginx.NewInternal(err)           // 基础设施错误 → 500
ginx.NewForbidden()             // 权限不足 → 401
// Controller: _ = c.Error(err) → ErrorHandlerMiddleware 统一处理
```

---

## cryptox

```go
import "github.com/ay/go-kit/cryptox"

hash, err := cryptox.HashPassword("secret")    // bcrypt, cost=12
ok := cryptox.CheckPassword("secret", hash)    // true
```

---

## money

高精度金额计算，基于 `shopspring/decimal`。

```go
import "github.com/ay/go-kit/money"

val, _ := money.ParseAmount("123.45")          // 校验最多两位小数
s := money.FormatAmount(val)                    // 银行家舍入 → "123.45"

// 四则运算（内部 8 位精度）
sum := money.AddAmounts(a, b)
diff := money.SubtractAmounts(a, b)
prod := money.MultiplyAmount(a, b)
quot, _ := money.DivideAmount(a, b, money.RoundUpToCent)

// 元 ↔ 分
cents := money.ToMinorUnits(val)               // 12345
val = money.FromMinorUnits(cents)              // 123.45

// 分账结算
user, platform, residual := money.SettleUserPlatform(
    userShare, platformShare, money.ResidualModeAllocateToUser,
)
```

---

## treex

通用树形结构构建。

```go
import "github.com/ay/go-kit/treex"

// Model 实现 TreeNode 接口
type Menu struct {
    ID       int64
    ParentID int64
    Name     string
    Children []treex.TreeNode
}
func (m *Menu) GetID() int64             { return m.ID }
func (m *Menu) GetParentID() int64       { return m.ParentID }
func (m *Menu) SetChildren(c []treex.TreeNode) { m.Children = c }
func (m *Menu) GetChildren() []treex.TreeNode  { return m.Children }

// 构建（parentID=0 为根节点）
tree := treex.BuildTree(menus)
```

---

## timex

```go
import "github.com/ay/go-kit/timex"

// 时间戳字段
CreatedAt timex.TimeModel `gorm:"column:created_at" json:"created_at"`

// 纯日期字段
Birthday *timex.DateModel `gorm:"column:birthday" json:"birthday"`

// 常量
timex.TimeFormat       // "2006-01-02 15:04:05"
timex.DateFormat       // "2006-01-02"
timex.ShanghaiLocation // *time.Location
```

零值处理：JSON 输出 `""`，DB 输出 `nil`。

---

## token

Token / TraceID / 随机字符串生成，基于 `crypto/rand`。

```go
import "github.com/ay/go-kit/token"

tok := token.GenerateToken()     // 48 字符 (12 时间 + 36 随机)
tid := token.GenerateTraceID()   // 32 字符 (8 时间 + 24 随机)
s := token.RandomString(6)       // 6 位随机字符串 (a-zA-Z0-9)
custom := token.Generate(4, 8)   // 自定义：4 位时间 + 8 位随机
```

---

## ctxutil

Context 值传播工具，所有 With/Get 成对使用。

```go
import "github.com/ay/go-kit/ctxutil"

ctx = ctxutil.WithUid(ctx, 42)
uid := ctxutil.GetUid(ctx)              // 42（未设置返回 0）

ctx = ctxutil.WithTenantID(ctx, 1)
ctx = ctxutil.WithTraceID(ctx, "abc123")
ctx = ctxutil.WithLang(ctx, "en")       // 默认 ctxutil.DefaultLang = "zh"
ctx = ctxutil.WithAccessToken(ctx, "tok_xxx")
ctx = ctxutil.WithRefreshToken(ctx, "ref_xxx")
```

---

## logger

结构化日志，自动从 context 提取 TraceID。

```go
import "github.com/ay/go-kit/logger"

// 初始化
log, _ := logger.New(logger.ConfigFromEnv())  // LOG_LEVEL + LOG_PATH
logger.SetGlobal(log)
defer log.Close()

// 实例方法
log.Info(ctx, "user %d logged in", uid)

// 全局便捷函数
logger.Debug(ctx, "detail: %v", v)
logger.Info(ctx, "ok")
logger.Warn(ctx, "slow query: %dms", ms)
logger.Error(ctx, "failed: %v", err)
```

---

## i18n

错误码 → 多语言文案映射。

```go
import "github.com/ay/go-kit/i18n"

// 初始化
cat := i18n.NewCatalog("zh")
cat.Register(i18n.CodeParamInvalid, map[string]string{
    "zh": "参数不合法",
    "en": "Invalid parameter",
})
i18n.SetGlobal(cat)

// 获取文案（自动 fallback 到默认语言）
msg := i18n.GetLangMsg(i18n.CodeParamInvalid, "en")  // "Invalid parameter"
```

预定义错误码：`CodeInternalError`(10001) / `CodeParamInvalid`(10002) / `CodeTokenInvalid`(20001) / `CodeUserNotFound`(40001) 等 20+ 个。

---

## rediscli

Redis 客户端，支持 Lua 脚本 SHA 缓存和分布式锁。

```go
import "github.com/ay/go-kit/rediscli"

// 连接
client, _ := rediscli.Open(rediscli.ConfigFromEnv())  // REDIS_ADDR / REDIS_PASSWORD / REDIS_DB
defer client.Close()

// 原生客户端
rdb := client.Redis()

// Lua 脚本（自动 EVALSHA + 回退 EVAL）
result, _ := client.ExecuteLuaScript(script, keys, args...)

// 分布式锁（单 key）
lock := rediscli.NewDistributedLock(rdb, "order:123", 10*time.Second)
_ = lock.Lock(ctx, 5*time.Second)   // 阻塞等待
defer lock.Unlock(ctx)
_ = lock.Extend(ctx, 5*time.Second) // 续期

// 批量锁（多 key 原子获取）
batch := rediscli.NewBatchLock(rdb, []string{"a", "b"}, 10*time.Second)
_ = batch.Lock(ctx, 5*time.Second)
defer batch.Unlock(ctx)
```

---

## auth

Token 鉴权服务，Redis + Lua 实现。

```go
import "github.com/ay/go-kit/auth"

// 初始化（读取 PROJECT_NAME 环境变量，如 "tc"）
cfg := auth.ConfigFromEnv()  // PROJECT_NAME / TOKEN_ACCESS_EXPIRE / TOKEN_REFRESH_EXPIRE
repo := auth.NewRedisRepository(cfg, redisClient, logFn)
svc := auth.NewService(repo, logFn)
// Redis key 格式: {project}_auth_{scene}_{key_type}:{value}
// 例: tc_auth_admin_user:123, tc_auth_admin_access_token:abc...

// 创建 token 对
tokens, _ := svc.Create(ctx, userID, "admin")  // → *auth.Tokens{AccessToken, RefreshToken}

// 校验（返回 userID）
uid, _ := svc.Validate(ctx, accessToken, refreshToken, "admin")

// 刷新
newTokens, _ := svc.Refresh(ctx, accessToken, refreshToken, "admin")

// 登出
_ = svc.Delete(ctx, userID, "admin")

// 在线状态
status, _ := svc.GetUserOnlineStatus(ctx, []int64{1, 2}, "admin")
// map[int64]auth.UserOnlineInfo{1: {LastAccess: ts, IsOnline: true}}
```

---

## ratelimit

滑动窗口限流，Redis ZSET + Lua 原子操作，含 Gin 中间件。

```go
import "github.com/ay/go-kit/ratelimit"

// 创建限流器（project 隔离多项目共享 Redis）
limiter := ratelimit.New(redisClient, "tc")  // Redis key: tc_ratelimit_{业务key}

// 直接调用（Service 层可用）
result, _ := limiter.Allow(ctx, "login:1.2.3.4", ratelimit.Rate{
    Limit:  5,
    Window: time.Minute,
})
// result.Allowed / result.Remaining / result.RetryAfter / result.Total

// 只读查询（不计入请求）
status, _ := limiter.Peek(ctx, "login:1.2.3.4", ratelimit.Rate{Limit: 5, Window: time.Minute})

// 读取窗口内所有请求时间戳（监控/调试）
entries, _ := limiter.Entries(ctx, "login:1.2.3.4", ratelimit.Rate{Limit: 5, Window: time.Minute})

// 重置
limiter.Reset(ctx, "login:1.2.3.4")
```

### Gin 中间件

```go
// 按 IP 限流（全局）
r.Use(ratelimit.GinMiddleware(limiter, ratelimit.GinConfig{
    KeyFunc:   ratelimit.KeyByIP("api"),
    Rate:      ratelimit.Rate{Limit: 60, Window: time.Minute},
    ErrorCode: i18n.CodeRateLimited,
}))

// 按请求头限流（API Key 维度）
v1.Use(ratelimit.GinMiddleware(limiter, ratelimit.GinConfig{
    KeyFunc:   ratelimit.KeyByHeader("rk", "X-Api-Key"),
    Rate:      ratelimit.Rate{Limit: 30, Window: time.Minute},
    ErrorCode: i18n.CodeRateLimited,
    Skip:      func(c *gin.Context) bool { return false }, // 可选跳过条件
}))
```

内置 KeyFunc（独立使用）：`KeyByIP` / `KeyByHeader` / `KeyByParam`

组合使用：`KeyCompose` + 维度提取器 `PartByIP` / `PartByHeader` / `PartByParam`

```go
// 组合多维度限流（IP + 用户名）
ratelimit.KeyCompose("login", ratelimit.PartByIP(), ratelimit.PartByHeader("X-User"))
// 生成 key: "login:ip:1.2.3.4:hdr:alice"
```

响应头：
- 通过时：`X-RateLimit-Limit` + `X-RateLimit-Remaining`
- 拒绝时：额外 `Retry-After`（秒）

Redis 故障时自动放行，不阻塞业务。

---

## 包依赖关系

```
ctxutil  ←  零依赖
   ↑
logger   ←  ctxutil
   ↑
dbx ←  logger, gorm          (不依赖 gin)
   ↑
ginx      ←  dbx, i18n, gin, token, ctxutil
   ↑
auth      ←  ginx, rediscli
ratelimit ←  ginx, rediscli

cryptox  ←  golang.org/x/crypto   (独立)
money    ←  shopspring/decimal     (独立)
treex    ←  零依赖                 (独立)
timex    ←  零依赖                 (独立)
token    ←  零依赖                 (独立)
```
