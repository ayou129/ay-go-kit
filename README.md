# go-kit

多项目共享基础设施库。

## 包一览

| 包 | 说明 |
|---|---|
| `auth` | Token 鉴权（创建/校验/刷新/删除），Redis + Lua 实现 |
| `cryptox` | 加解密工具（bcrypt 密码哈希） |
| `ctxutil` | Context 工具（UserID / TenantID / TraceID / Lang / Token） |
| `dbx` | PostgreSQL 连接 + 事务 + 分页 + 筛选（GORM + pgx） |
| `ginx` | Gin 扩展（错误处理 / 响应封装 / 请求绑定清洗 / 分页查询 DTO） |
| `i18n` | 国际化（错误码 → 多语言文案） |
| `logger` | 结构化日志（TraceID 自动注入） |
| `money` | 高精度金额计算（shopspring/decimal） |
| `rediscli` | Redis 客户端（连接管理 + Lua 脚本缓存） |
| `timex` | 时间类型（TimeModel / DateModel）+ 上海时区 |
| `token` | Token / TraceID 生成 |
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
ginx.SuccessWithMsg(c, code, data)
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

## 包依赖关系

```
ctxutil  ←  零依赖
   ↑
logger   ←  ctxutil
   ↑
dbx ←  logger, gorm          (不依赖 gin)
   ↑
ginx     ←  dbx, i18n, gin
   ↑
auth     ←  ginx, rediscli

cryptox  ←  golang.org/x/crypto   (独立)
money    ←  shopspring/decimal     (独立)
treex    ←  零依赖                 (独立)
timex    ←  零依赖                 (独立)
token    ←  零依赖                 (独立)
```
