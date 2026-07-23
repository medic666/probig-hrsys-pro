# AGENTS.md

## Build & Run

```bash
# 完整构建（先前端，再 Go 二进制）：
cd web && bun run build && cd .. && go build -o probig .

# 开发模式（两个终端）：
# 终端1：go run .
# 终端2：cd web && bun run dev    # /api 代理到 :8080

# 运行：
./probig            # 默认 :8080，首次运行自动创建 probig.db

# 代码检查：
go vet ./...        # Go 静态分析
cd web && bun run build    # 包含 tsc 类型检查 + vite build
```

**构建顺序不可逆**：Go 二进制通过 `//go:embed` 嵌入 `web/dist/`。发布前必须 `bun run build` → `go build`。

## 环境

- **Go module**：`github.com/medic666/probig`（Go 1.22+）
- **包管理器**：Bun（做 `vite build` 即可，npm/yarn 也能用）
- **数据库**：纯 Go SQLite（`modernc.org/sqlite`），无 CGO。WAL 模式，`SetMaxOpenConns(1)` 单连接
- **配置**：工作目录下的 `config.yaml`。字段：`server.port`(8080)、`jwt.secret`、`database.path`、`upload.dir`
- **默认账号**：`admin` / `admin123`（`database.RunSeed` 中 bcrypt 哈希，幂等）

## 架构

### 分层设计
```
main.go  →  handlers  →  services  →  models (sqlx + DB)
                ↓
           middleware (JWT, RBAC)
```

- `handlers/`：Gin handler，做参数绑定和响应格式化。**不要在这里写业务逻辑**。
- `services/`：所有写操作必须使用显式 `tx.Beginx()` / `tx.Commit()` 事务。审计日志在同一事务内写入。
- `models/`：数据库结构体（`db:` 标签）、JSON 请求/响应类型、快照数据结构。
- `middleware/`：`AuthRequired` 从 JWT 提取 `user_id`、`username`、`role_id`、`perms` 到 `gin.Context`。`RBAC(module, action)` 检查 `module:action` 是否在权限列表中。

### 事件驱动快照引擎（核心架构，不可改）

人员/组织的**所有字段变更**都必须通过事件。快照引擎在同一事务内从头重建完整快照链：

1. 插入/更新/删除 `person_events` 或 `org_events`
2. 调用 `RebuildPersonSnapshots(tx, personID)` 或 `RebuildOrgSnapshots(tx, orgID)`
3. 引擎 DELETE 该实体的全部现有快照，按 `effective_date ASC` 重放所有事件，将每个事件的 payload 合并到累积状态，每个事件插入一行快照

**推论**：最新快照（按 `effective_date DESC, id DESC`）即为实体当前状态。查询"当前状态"只需取最新快照，不要在查询时跨事件 JOIN。

**合并规则**（`services/snapshot.go`）：非零值覆盖，零值跳过。数值字段 `!= 0` 才更新，字符串字段 `!= ""` 才更新。这意味着一旦将数值字段设为非零，无法通过事件将其重置为 0。

### 考勤 → 工资计算顺序

工资计算依赖 `attendance_summaries`，所以**必须先算考勤再算工资**：
```
POST /api/attendance/calculate  →  写入 attendance_summaries
POST /api/salary/calculate      →  读取 attendance_summaries + person_snapshots → 写入 salary_summaries
```
两者均为手动触发（管理员点击按钮），不会在事件变更时自动计算。

### SPA 前端路由回退

Go 后端必须为所有非 API、非静态文件路径返回 `index.html`，否则刷新页面时浏览器直接向服务器请求 `/person/1` 这种 SPA 路由会 404。此逻辑在 `router.go` 的 `NoRoute` 处理器中实现：先尝试从 `web/dist/` 提供请求路径对应的文件，失败则回退到 `index.html`。

### SQLite 注意事项
- 最大连接数：**1**（SQLite 锁约束），不要增大。
- DSN 包含 `?_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=on`
- 所有迁移均为幂等（`CREATE TABLE IF NOT EXISTS`、`INSERT OR IGNORE`）
- `internal/database/migrations/` 下的 `.sql` 文件在每次启动时按文件名顺序执行
- 给 summary/aggregate 表添加列时，务必核对该 service 中 `INSERT OR REPLACE` 语句的 `?` 占位符数量与新列数一致。SQLite 会抛出含义模糊的 "X values for Y columns" 错误。

## 前端

- React 18 + Vite 5 + Ant Design 5 + react-router-dom v6
- `web/src/services/api.ts`：axios 封装，带 JWT 拦截器（自动附加 Bearer token，401 时跳转 `/login`）
- `web/src/hooks/useAuth.tsx`：**React Context 模式**（不是独立 useState）。`AuthProvider` 包裹整个应用，子组件通过 `useAuth()` 共享同一份认证状态。包括：`user`、`perms`、`loading`、`login()`、`logout()`、`hasPermission()`。
- RBAC：前端通过 `hasPermission(module, action)` 检查权限列表，控制菜单项和操作按钮的显示
- 导出：前端用 `xlsx` 库做客户端 Excel 生成；服务端有 `/api/export/*` 端点用 `excelize` 做备选

### 人员事件表单（EventForm）注意事项

`Person/EventForm.tsx` 和 `Organization/Detail.tsx` 的表单提交构造 JSON payload 时，**数值字段不能发空字符串 `""`**，这在 Go 端会因 `json: cannot unmarshal string into Go struct field ... of type float64` 导致 `ShouldBindJSON` 失败。应区分字符串字段和数值字段：只有实际填写了的数值字段才包含在 payload 中；未填的字段应返回 `undefined`（JSON 序列化时自动省略），由快照引擎的"非零才合并"规则处理。

### 列表查询中的名称映射

以下 model 结构体包含 `PersonName` 或 `Username` 字段（带 `db:` 和 `json:` 标签）：
- `AttendanceEvent`、`AttendanceSummary`、`SalaryEvent`、`SalarySummary`：`person_name` 来自 `LEFT JOIN entities`
- `AuditLog`：`username` 来自 `LEFT JOIN users`

对应的 service 方法（`ListEvents`、`ListSummaries`、`List`）的 SQL 查询已包含 JOIN。前端可直接使用这些字段显示名称，导出时也应导出名称而非 ID。

## 测试

暂无测试框架。手动验证：
```bash
# 启动服务
./probig &
# 登录获取 token
TOKEN=$(curl -s -X POST http://localhost:8080/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"admin123"}' | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['token'])")
# 测试接口：-H "Authorization: Bearer $TOKEN"
```

## RBAC 模型

- 权限在 `roles` 表和 `role_permissions`（module + action）中定义
- 模块：`person`、`organization`、`attendance`、`salary`、`file`、`audit`
- 操作：`read`、`write`、`delete`
- 权限以 `["module:action", ...]` 格式嵌入 JWT claims
- 管理员角色在种子数据中获得全部 18 个权限组合
- 新增模块时需要：种子 SQL、`router.go` 中的 RBAC 中间件注册、前端菜单过滤
