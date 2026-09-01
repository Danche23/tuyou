# TuYou - 户外旅行平台 API

一个基于 Go 语言开发的户外旅行平台后端 API，包含前台用户系统和后台管理系统。

## 功能特性

### 前台功能
- **用户认证**：手机号注册/登录、密码重置、短信验证码
- **笔记模块**：发布/查看/删除笔记、点赞、收藏
- **攻略模块**：发布/查看/删除攻略、点赞、收藏
- **活动模块**：创建活动（领队专属）、报名/取消、活动管理

### 后台管理
- **管理员认证**：图形验证码登录、个人信息管理
- **会员管理**：列表查询、详情查看、状态修改、批量操作、导出 Excel
- **领队管理**：同会员管理，角色固定为领队
- **笔记管理**：分页查询、详情查看、编辑、删除、批量操作

## 技术栈

| 组件 | 库 | 说明 |
|------|-----|------|
| HTTP 框架 | gin-gonic/gin v1.11.0 | |
| ORM | gorm.io/gorm v1.31.2 + driver/mysql | AutoMigrate 26 张表 |
| 缓存 | redis/go-redis/v9 | 可选，失败不阻塞启动 |
| JWT | golang-jwt/jwt/v5 | HS256，30 天过期 |
| 短信验证码 | alibabacloud dypnsapi v3 | 前台 send-code |
| 图形验证码 | mojocn/base64Captcha | 后台登录 |
| Excel 导出 | xuri/excelize/v2 | 会员/领队导出 |
| 配置 | spf13/viper | config.yaml → `TUYOU_*` 环境变量 |
| 日志 | go.uber.org/zap + lumberjack | 文件+控制台双写 |
| 验证 | go-playground/validator/v10 | |
| JSON | bytedance/sonic | |

## 项目结构

```
cmd/server/main.go                # 入口
cmd/gen_token/main.go             # 辅助生成 token
configs/config.yaml               # 主配置
internal/
  api/
    router.go                     # 全局中间件 + 路由注册
    v1/web/{auth,note,guide,activity}/
    v1/admin/{auth,member,center,note}/
  app/app.go                      # 生命周期 + 依赖注入
  middleware/                      # auth、cors、logger、recovery
  model/entity/                   # 26 张表 Entity
  model/dto/request|response/     # 按 admin/web 分模块
  repository/                     # 每模块 interface + 实现
  service/                        # 每模块 interface + 实现
  stream/consumer.go              # Redis Stream 异步浏览量消费者
pkg/
  config/  database/  errors/  jwt/  logger/  response/  utils/
```

## 快速开始

### 环境要求

- Go 1.25.0+
- MySQL 8.0+
- Redis 7.0+ (可选)

### 安装

```bash
# 克隆仓库
git clone https://github.com/qianqiang-del/tuyou.git
cd tuyou

# 安装依赖
make deps
```

### 配置

编辑 `configs/config.yaml`，配置数据库和 Redis 连接：

```yaml
database:
  host: 127.0.0.1
  port: 3306
  user: root
  password: your_password
  dbname: tuyou

redis:
  host: 127.0.0.1
  port: 6379
```

### 运行

```bash
# 启动服务
make run

# 编译
make build

# 运行测试
make test
```

服务默认运行在 `http://localhost:9300`

### 生成测试 Token

```bash
go run cmd/gen_token/main.go
```

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

## 开发规范

### 分层架构

```
DTO → Repository → Service → Controller → Router → App
```

| 层 | 职责 |
|----|------|
| DTO | 定义请求/响应结构体 + binding 校验 |
| Repository | 纯数据库操作 |
| Service | 业务逻辑 + 组装响应 |
| Controller | 取参数 → 调 Service → 返回响应 |

### 文件命名

- `repository/{xxx_interface.go, xxx_repository.go}`
- `service/{xxx_interface.go, xxx_service.go}`
- `api/v1/{xxx/controller.go, xxx_routes.go}`
- `dto/request/{xxx}.go`
- `dto/response/{xxx}.go`

## 部署

### 编译

```bash
make build
```

生成二进制文件 `bin/tuyou-api.exe`

### 运行

```bash
./bin/tuyou-api.exe
```

## 分支管理

- `master` - 主分支
- `develop` - 开发分支
- `feature/*` - 功能分支

## 许可证

MIT License
