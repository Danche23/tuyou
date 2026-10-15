# tuyou

户外旅行平台后端服务 —— 基于 Go + Gin 构建，同时提供**前台会员端**与**后台运营管理端**两套 API。

覆盖笔记/攻略/活动的内容发布与互动、会员体系、评论与敏感词治理、消息通知（SSE 实时推送）、
文件上传（阿里云 OSS）以及完整的 RBAC 权限管理。

---

## 功能模块

### 前台（会员端 `/api/v1`）

| 模块 | 说明 |
|---|---|
| 认证 | 手机号验证码登录/注册、密码登录、图形验证码、注销账号 |
| 笔记 | 发布、编辑、详情、列表、点赞、收藏、标签、图片 |
| 攻略 | 与笔记同构的内容体系（发布、互动、标签） |
| 活动 | 活动发布、报名、详情、收藏、开始前提醒 |
| 社区首页 | 信息流聚合、轮播图 |
| 评论 | 两级评论模型、点赞、热评排序（热度后台异步重算） |
| 个人中心 | 资料、我的发布、收藏、关注/粉丝、参与记录 |
| 消息 | 互动通知、SSE 实时推送、未读数 |
| 搜索 | 全量搜索、热搜词、搜索历史 |
| 文件上传 | 图片批量上传（魔数校验 + 压缩）、视频上传（自动截取封面） |

### 后台（运营管理端 `/api/v1/admin`）

| 模块 | 说明 |
|---|---|
| 权限管理 | 管理员、角色、菜单、接口权限（RBAC，接口级细粒度鉴权） |
| 内容管理 | 笔记、攻略、活动、标签的审核与维护 |
| 会员管理 | 会员列表、状态调整、关注关系清理 |
| 评论管理 | 评论筛选、状态查看、硬删除清理 |
| 敏感词管理 | 增删改查，**修改后热更新内存词库，无需重启** |
| 消息管理 | 模板管理、群发消息 |
| 工作台 | 运营数据概览、Excel 导出 |

---

## 技术栈

| 分类 | 选型 | 版本 |
|---|---|---|
| 语言 | Go | 1.25 |
| Web 框架 | Gin | v1.11.0 |
| ORM | GORM + MySQL Driver | v1.31.2 / v1.6.0 |
| 数据库 | MySQL | 8.0 |
| 缓存 | Redis（go-redis） | v9.21.0 |
| 配置 | Viper | v1.21.0 |
| 日志 | Zap + Lumberjack（自动切割） | v1.28.0 |
| 鉴权 | JWT | v5.3.1 |
| 密码哈希 | golang.org/x/crypto/bcrypt | — |
| 短信 | 阿里云号码认证服务（DyPNS） | — |
| 对象存储 | 阿里云 OSS（SDK v2） | v1.5.3 |
| 图形验证码 | base64Captcha | v1.3.8 |
| 文件类型校验 | mimetype（读魔数） | v1.4.12 |
| 视频处理 | ffmpeg-go | v0.5.0 |
| Excel 导出 | excelize | v2.11.0 |

---

## 架构

标准三层架构，依赖单向流动：`Handler → Service → Repository`。

```
cmd/server/          # 程序入口
configs/             # 配置文件（*.yaml 不入库，只入库 config.example.yaml）
internal/
├── api/             # 路由注册 + HTTP Handler（按 admin / web 分组）
├── app/             # 应用装配与生命周期（DI 容器、优雅停机）
├── middleware/      # Recovery / Logger / RequestLogger / CORS / Auth / MemberAuth / PermissionAuth / UploadAuth
├── model/
│   ├── entity/      # GORM 实体
│   └── dto/         # 请求/响应结构体
├── repository/      # 数据访问层（按 admin / web 分组）
├── service/         # 业务逻辑层（按 admin / web 分组）
├── storage/         # 文件存储抽象（oss | local）
├── stream/          # 浏览量异步消费 + SSE Hub
└── task/            # 定时任务
pkg/                 # 可复用基础组件
├── config/  database/  errors/  excel/
├── jwt/     logger/    response/  sensitive/  utils/
sql/                 # 种子数据（建表后初始化用）
nginx/               # Nginx 反向代理配置
docker-compose.yml       # 本地开发编排
docker-compose.prod.yml  # 生产编排（MySQL/Redis 不暴露端口）
```

### 关键设计说明

- **浏览量异步落库**：前台浏览量先写 Redis，由 `stream.Consumer` 异步批量回写 MySQL，避免高频写打满数据库。
- **热度后台重算**：点赞/评论热度不在请求链路里计算，由定时任务周期性校准（`task/heat.go`、`task/comment_heat.go`）。
- **定时校准任务**：发布数校准（`member_publish.go`）、活动开始前提醒（`activity_remind.go`）——用周期性校准替代分布式计数，避免计数漂移。
- **敏感词热更新**：启动时从 `sensitive_word` 表加载启用词构建 Trie；后台改动后原子替换内存词库，**加载失败直接 `Fatal`，不带病上线**。
- **SSE 实时推送**：`stream.SSEHub` 按会员 ID 维护多端连接，非阻塞投递（对端离线或队列满则丢弃，不拖累调用方）。
- **优雅停机**：收到 `SIGINT`/`SIGTERM` 后依次停止 HTTP 服务、消费协程、各定时任务，再关闭数据库与 Redis 连接。

---

## 快速开始

### 前置依赖

- Go 1.25+
- MySQL 8.0+
- Redis 7+（未启动时服务仍可运行，仅 Redis 相关增强功能降级）

### 1. 准备配置

```bash
cp configs/config.example.yaml configs/config.yaml
```

然后按需修改 `configs/config.yaml`。**所有敏感值建议改用环境变量注入，不要写回文件**（见下节）。

> `configs/*.yaml` 已在 `.gitignore` 中排除，不会被提交；只有 `config.example.yaml` 会入库。

### 2. 环境变量（推荐的生产做法）

以下环境变量存在时，会**覆盖**配置文件中的同名字段：

| 环境变量 | 覆盖的配置项 |
|---|---|
| `MYSQL_PASSWORD` | `database.mysql.password` |
| `REDIS_PASSWORD` | `database.redis.password` |
| `JWT_SECRET` | `jwt.secret` |
| `SMS_ACCESS_KEY_ID` | `sms.access_key_id` |
| `SMS_ACCESS_KEY_SECRET` | `sms.access_key_secret` |
| `OSS_ACCESS_KEY_ID` | `storage.oss.access_key_id` |
| `OSS_ACCESS_KEY_SECRET` | `storage.oss.access_key_secret` |

例如：

```bash
export JWT_SECRET="$(openssl rand -base64 48)"
export MYSQL_PASSWORD='your-strong-password'
go run ./cmd/server
```

未设置这些变量时，配置文件中的值原样生效（本地开发不受影响）。

### 3. 初始化数据库

服务启动时会自动执行表结构迁移（`AutoMigrate`），但**不会自动插入敏感词种子数据**。仅在全新空库中手动导入下方种子；随后检查 `sensitive_word` 表中启用词数量，再启动服务。若已有业务数据，不要执行 `sql/seed_data.sql`：它会清空多张业务表。

```bash
mysql -u root -p tuyou < sql/sensitive_word_seed.sql
```

### 4. 启动

```bash
make run                 # 等同于 go run cmd/server/main.go
# 或指定配置文件
go run ./cmd/server -c configs/config.yaml
```

服务默认监听 `:9300`。

### 5. 健康检查

```bash
curl http://127.0.0.1:9300/api/v1/health
# {"status":"ok","message":"TuYou API is running"}
```

---

## API 接口

### 健康检查
- `GET /api/v1/health`

### 前台 `/api/v1`

| 模块 | 认证 | 接口 |
|------|------|------|
| auth | 公开/JWT | `GET /auth/captcha` `POST /auth/register` `POST /auth/login` `POST /auth/logout` `POST /auth/reset-password` `POST /auth/send-code` `POST /auth/change-phone`(JWT) `PUT /auth/password`(JWT) `POST /auth/cancel`(JWT) |
| notes | JWT | `POST /notes` `GET /notes/:id` `DELETE /notes/:id` `POST /notes/:id/like` `POST /notes/:id/favorite` |
| guides | JWT | `POST /guides` `GET /guides/:id` `POST /guides/:id/like` `POST /guides/:id/favorite` `DELETE /guides/:id` |
| activities | JWT | `GET /activities` `GET /activities/:id` `POST /activities` `POST /activities/:id/signup` `DELETE /activities/:id` |
| comments | JWT | `POST /comments` `GET /comments` `DELETE /comments/:id` `POST /comments/like/:id` `GET /comments/replies/:id` |
| personal | JWT | `GET /personal` `PUT /personal` `GET /personal/member/:id` `GET /personal/follow` `GET /personal/fans` `POST /personal/follow/:followeeId` `GET /personal/notes` `GET /personal/strategys` `GET /personal/liked-notes` `GET /personal/liked-strategys` `GET /personal/favorite-note` `GET /personal/favorite-strategy` `GET /personal/publish-activity` `GET /personal/join-activity` |
| home | 公开 | `GET /slides` `GET /feed` |
| search | 公开/JWT | `GET /search` `GET /hot-searches` `GET /search/history`(JWT) `DELETE /search/history/:id`(JWT) `DELETE /search/history`(JWT) |
| tags | 公开 | `GET /tags` |
| members | JWT | `GET /members/info` |
| messages | JWT | `GET /messages/interactions` `GET /messages/system` `GET /messages/unread-count` `GET /messages/stream` `PUT /messages/:id/read` `PUT /messages/read-all` `DELETE /messages/:id` `POST /messages/interactions/:id/reply` |
| upload | JWT | `POST /upload/images` `POST /upload/video` |

### 后台 `/api/v1/admin`

| 模块 | 认证 | 接口 |
|------|------|------|
| auth | 公开/JWT | `GET /admin/captcha` `POST /admin/login` `POST /admin/logout` `GET /admin/auth/me`(JWT) `PUT /admin/auth/password`(JWT) |
| members | JWT | `GET /admin/members` `GET /admin/members/:id` `POST /admin/members/:id/restore` `PUT /admin/members/:id` `DELETE /admin/members/:id` `POST /admin/members/batch-delete` `PUT /admin/members/batch-status` `GET /admin/members/export` |
| leaders | JWT | `GET /admin/leaders` `GET /admin/leaders/:id` `PUT /admin/leaders/:id` `DELETE /admin/leaders/:id` `POST /admin/leaders/batch-delete` `PUT /admin/leaders/batch-status` `GET /admin/leaders/export` |
| center | JWT | `PUT /admin/avatar` `PUT /admin/profile` `GET /admin/notifications` `GET /admin/notifications/unread` `GET /admin/notifications/stream` `PUT /admin/notifications/:id/read` `PUT /admin/notifications/read-all` `DELETE /admin/notifications/:id` `DELETE /admin/notifications` |
| notes | JWT | `GET /admin/notes` `GET /admin/notes/:id` `PUT /admin/notes/:id` `DELETE /admin/notes/:id` `POST /admin/notes/batch-delete` `POST /admin/notes/export` |
| guides | JWT | `GET /admin/guides` `GET /admin/guides/:id` `PUT /admin/guides/:id` `DELETE /admin/guides/:id` `POST /admin/guides/batch-delete` `POST /admin/guides/export` |
| activities | JWT | `GET /admin/activities` `GET /admin/activities/:id` `POST /admin/activities` `PUT /admin/activities/:id` `DELETE /admin/activities/:id` `POST /admin/activities/batch-delete` `POST /admin/activities/export` `GET /admin/leadernames` |
| comments | JWT | `GET /admin/comments` `GET /admin/comments/:id` `DELETE /admin/comments/:id` `POST /admin/comments/batch-delete` |
| tags | JWT | `GET /admin/tags` `POST /admin/tags` `PUT /admin/tags/:id` `DELETE /admin/tags/:id` `POST /admin/tags/batch-delete` |
| sensitive-words | JWT | `GET /admin/sensitive-words` `POST /admin/sensitive-words` `DELETE /admin/sensitive-words/:id` |
| permission/admins | JWT | `GET /admin/admins` `GET /admin/admins/role` `POST /admin/admins` `PUT /admin/admins/:id` `DELETE /admin/admins/:id` `PUT /admin/admins/:id/reset-password` |
| permission/menus | JWT | `GET /admin/menus` `GET /admin/menus/all` `POST /admin/menus` `PUT /admin/menus/:id` `DELETE /admin/menus/:id` |
| permission/roles | JWT | `GET /admin/roles` `POST /admin/roles` `PUT /admin/roles/:id` `DELETE /admin/roles/:id` `GET /admin/roles/:id/menus` `GET /admin/roles/:id/permissions` `PUT /admin/roles/:id/permissions` |
| permission/api-permissions | JWT | `GET /admin/api-permissions` `GET /admin/api-permissions/module` `POST /admin/api-permissions` `PUT /admin/api-permissions/:id` `DELETE /admin/api-permissions/:id` |
| workbench | JWT | `GET /admin/workbench` |
| messages | JWT | `GET /admin/interactions` `GET /admin/system` `POST /admin/system` `DELETE /admin/:id` |

## 响应格式

统一响应格式：

```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

- `code: 0` 表示成功
- `code: 4xx/5xx` 表示错误

## 开发

```bash
make build          # 编译到 bin/tuyou-api.exe
make test           # 运行全部测试
make test-coverage  # 生成 coverage.html
make fmt vet lint   # 格式化 / 静态检查 / lint
make help           # 查看全部命令
```

数据库相关的集成测试通过环境变量开关控制，未开启时自动跳过，不会拖累无数据库环境的机器。

接口契约：[docs/api/openapi.json](docs/api/openapi.json)（OpenAPI 3.0，全部 121 个接口）。

---

## CI 与贡献

仓库提供 GitLab CI 配置（build + vet + test，lint 暂为非阻断），见 [.gitlab-ci.yml](.gitlab-ci.yml)；尚未在 Runner 上实测。
分支模型、提交规范与署名要求见 [CONTRIBUTING.md](CONTRIBUTING.md)。

---

## 部署

### Docker Compose（推荐）

```bash
# 本地开发（端口仅绑定 127.0.0.1；先准备 .env 和 configs/config.docker.yaml）
docker compose --env-file .env up -d

# 生产编排草案（HTTP 暂只绑定本机；未补齐 HTTPS 前不可开放公网）
# 先按 docs/DEPLOY.md 准备 .env 和 configs/config.docker.yaml，
# 将配置中的 MySQL/Redis host 分别改为 mysql/redis。
docker compose -f docker-compose.prod.yml --env-file .env up -d
```

完整流程、上线前安全清单与尚未完成的回滚设计见 [docs/DEPLOY.md](docs/DEPLOY.md)。
仓库同时提供 `Dockerfile`（多阶段构建）。

---

## 安全约定

- **不要把任何真实凭据提交进仓库。** `configs/*.yaml` 已被 `.gitignore` 排除，
  新增环境时请沿用 `config.<环境>.yaml` 命名，不要破坏这条规则。
- `configs/config.example.yaml` 有测试守卫（`pkg/config/loader_test.go`）：
  一旦其中出现 `LTAI` 前缀的 AccessKey 或弱默认密码，测试会立刻失败。
- 生产环境请通过环境变量注入全部密钥，并确保 `jwt.secret` 为随机值。

---

## License

`LICENSE` 中的 MIT 文本目前是拟议许可；**项目涉及多位贡献者，须在有权人确认授权范围后才能正式对外声明 MIT 授权**。在确认前不要把仓库或文档标成已获 MIT 授权。
