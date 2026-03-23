# go-kit 协调指南

本文件用于指导 AI 助手（Claude / ChatGPT 等）在修改 go-kit 或使用 go-kit 的项目时，遵循正确的协调流程。

## 仓库关系

```
go-kit（基础设施库）
├── 项目 A：/Users/ay/Desktop/app/ay-tianchang56/backend_go
├── 项目 B：/Users/ay/Desktop/app/ay-flowi/backend
└── 未来项目...
```

go-kit 是多项目共享的底层库，任何改动都会影响所有下游项目。

## 协调流程 （必须按顺序执行）

### 第一步：审计 go-kit 自身

在对任何项目做迁移或改动之前，先确认 go-kit 当前状态：

1. **阅读 README.md** — 了解所有包的用途和 API
2. **运行测试** — `cd /Users/ay/Desktop/app/ay-go-kit && go test ./... && go vet ./...`
3. **检查一致性**：
   - 包之间的依赖关系是否合理（dbx 不依赖 ginx，ginx 可依赖 dbx）
   - 导出的 API 命名是否统一
   - 是否有重复或冲突的功能
4. 特别检查 多项目结合Kit在DB的使用交错是否产生特殊的问题
5. **向用户报告**审计结果，等待确认后再进入第二步

### 第二步：检查 相关协调项目 覆盖情况，并且更新 使用现状的进度

### 第三步：协调项目改动

确认 go-kit 无误后，再根据实际任务修改目标项目：

- **新增 go-kit 能力**：先改 go-kit（加功能 + 测试），再改项目
- **项目迁移到 go-kit**：识别项目中可替换为 go-kit 调用的代码，逐个替换
- **修复 bug**：判断 bug 在 go-kit 还是项目层，在正确的层修复

### 第三步：反思 项目中有哪些能力是可以继续提出到本Kit中的

### 第四步：保证本文档和README.md文档内容描述和Kit代码一致即可

## go-kit 设计原则

- **只收通用代码**：至少两个项目会用到的才放进 kit，业务特有的留在项目里
- **不依赖业务**：kit 的包不能 import 任何项目代码
- **接口优先**：对外暴露接口（如 TxRunner），实现可替换
- **context 传播**：事务、TraceID、UserID 等通过 context 传递，不用全局变量
- **测试覆盖**：kit 的每个导出函数必须有测试，因为下游项目依赖它的正确性

## 各项目的 go-kit 使用现状

### tianchang56（迁移中）

- **已完成（Role 试点）**：repo 去掉 `db *gorm.DB` 参数，service 用 `dbx.Transaction()` 包函数，controller 用 `ginx.BindJSON` + `ginx.PageQueryDTO`
- 其余实体仍使用旧模式 → 逐个迁移：repo 内部 `dbx.GetDB(ctx)` / service 不注入 db / `ginx.PageQueryDTO` + `dbx.FindByPage`
- 清洗用项目自己的 `utils.BindJSON` → 目标：切换到 `ginx.BindJSON`
- 密码用项目自己的 `utils.HashPassword` → 目标：切换到 `cryptox.HashPassword`

### flowi（待迁移）

- repo 层使用 `dbx.BaseRepository`（已废弃）→ 目标：改用 `dbx.GetDB(ctx)` 直接调用
- 分页手动 `parsePage()` → 目标：切换到 `ginx.PageQueryDTO` + `dbx.FindByPage`
- 请求绑定用 `c.ShouldBindJSON` 无清洗 → 目标：切换到 `ginx.BindJSON`
- 密码用裸 bcrypt → 目标：切换到 `cryptox.HashPassword`

## 关键 API 速查

| 场景 | go-kit API |
|------|-----------|
| 获取 DB 连接 | `dbx.GetDB(ctx)` |
| 事务（推荐） | `dbx.Transaction(ctx, dbx.GetDB(ctx), fn)` 包函数直接调用 |
| 事务（可选） | `dbx.TxRunner` 接口（需 mock 事务的场景） |
| 分页查询 | `dbx.FindByPage[T](ctx, pq, scopes...)` |
| 分页请求绑定 | `ginx.PageQueryDTO.Bind(c)` |
| 筛选转 Scope | `ginx.PageQueryDTO.ToScopes(ctx, allowed)` |
| 请求绑定+清洗 | `ginx.BindJSON(c, &req)` |
| 密码哈希 | `cryptox.HashPassword(password)` |
| 密码校验 | `cryptox.CheckPassword(password, hash)` |
| 记录不存在 | `dbx.IsRecordNotFound(err)` |
| 金额计算 | `money.ParseAmount / FormatAmount / AddAmounts / ...` |
| 树形构建 | `treex.BuildTree[T](nodes)` |
| 错误响应 | `ginx.NewError(code)` / `ginx.NewInternal(err)` |
| 成功响应 | `ginx.Success(c, data)` |
