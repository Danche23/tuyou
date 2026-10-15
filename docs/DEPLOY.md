# 部署说明

本文档描述 tuyou-api 的可复现部署流程。

> ⚠️ 文档中的「上线前必做」小节列出了当前**尚未落地**的安全项。
> 在这些项全部完成前，不要把服务暴露到公网。

---

## 一、上线前必做（安全清单）

| # | 事项 | 为什么必须做 |
|---|---|---|
| 1 | **数据库端口不对宿主机开放** | 生产环境 MySQL/Redis 只需在容器网络内可达。端口映射到宿主机 = 数据库直接暴露 |
| 2 | **全部密码改为强随机值** | 弱密码 + 端口暴露等于没有密码 |
| 3 | **`jwt.secret` 改为随机值** | 默认值公开可查，任何人都能伪造管理员 token |
| 4 | **启用 HTTPS** | 否则 JWT、手机号、验证码在公网明文传输 |
| 5 | **CORS 改为前端域名白名单** | 不要用 `["*"]` 搭配 `allow_credentials: true` |
| 6 | **Nginx 配置 `client_max_body_size`** | Nginx 默认上限 1MB，而视频接口允许 300MB，不放开会直接 413 |
| 7 | **接口限流** | 尤其是短信验证码接口，不加限流存在被刷、产生真实费用的风险 |

生成随机密钥：

```bash
openssl rand -base64 48
```

---

## 二、环境准备

| 依赖 | 版本要求 |
|---|---|
| Go | 1.25+（源码构建时） |
| Docker | 20.10+ |
| Docker Compose | v2（使用 `docker compose` 子命令） |
| MySQL | 8.0+ |
| Redis | 7+ |
| Nginx | 任意近期版本（反向代理 + 静态资源） |

---

## 三、配置注入

**核心原则：真实凭据不落进仓库，由部署机受限的 `.env` 文件注入容器环境；容器环境与运行时配置仍可被有 Docker 管理权限的人读取，须限制权限。**

准备配置文件（不含真实凭据）：

- 本地直接运行二进制：复制为 `configs/config.yaml`。
- 使用本仓库的 Dockerfile / Docker Compose：复制为 `configs/config.docker.yaml`，因为容器启动命令固定读取这个文件。

```bash
cp configs/config.example.yaml configs/config.docker.yaml
```

**Docker 部署还必须修改这份文件中的非密钥地址**：将 `database.mysql.host` 从 `127.0.0.1` 改成 `mysql`，将 `database.redis.host` 从 `127.0.0.1` 改成 `redis`。容器内的 `127.0.0.1` 只指向 API 容器自身，连不到另外两个服务；保留端口 `3306`、`6379` 和 `app.port: 9300`。本地直接运行时保留 `127.0.0.1`，改为复制到 `configs/config.yaml` 即可。生产还要将样例中的 CORS 域名、短信与 OSS 等非密钥设置替换为实际配置，密钥则仍由下述环境变量注入。

准备环境变量文件（**不要提交**）：

```bash
cat > .env <<'EOF'
MYSQL_PASSWORD=<强随机密码>
REDIS_PASSWORD=<强随机密码>
JWT_SECRET=<openssl rand -base64 48 的输出>
SMS_ACCESS_KEY_ID=<阿里云短信 AccessKeyId>
SMS_ACCESS_KEY_SECRET=<阿里云短信 AccessKeySecret>
OSS_ACCESS_KEY_ID=<阿里云 OSS AccessKeyId>
OSS_ACCESS_KEY_SECRET=<阿里云 OSS AccessKeySecret>
EOF

chmod 600 .env
```

`configs/*.yaml` 与 `.env` 均已被 `.gitignore` 排除。
环境变量与配置项的对应关系见 [README 的环境变量表](../README.md#2-环境变量推荐的生产做法)。

---

## 四、构建

### 方式 A：Docker 构建（推荐）

```bash
docker build -t tuyou-api:latest .
```

镜像不会内置真实配置；单独运行镜像时必须挂载 `configs/config.docker.yaml` 并注入必要环境变量。生产编排已负责挂载，不能把“镜像构建成功”视为“服务可独立启动”。

### 方式 B：本地构建

```bash
make build
# 产物：bin/tuyou-api.exe
```

---

## 五、启动

### 方式 A：Docker Compose（推荐）

仓库提供两份编排文件：

| 文件 | 用途 | MySQL / Redis 端口 |
|---|---|---|
| `docker-compose.yml` | 本地开发 | 仅映射到 `127.0.0.1:3308` / `127.0.0.1:6379`，仅本机连库 |
| `docker-compose.prod.yml` | 待 HTTPS 补齐的生产编排草案 | **不映射**，仅 Docker 网络内可达；HTTP 暂只监听 `127.0.0.1:80` |

在隔离环境试运行生产编排草案（密码由环境变量注入，**未设置会直接拒绝启动**；当前 HTTP 只绑定 `127.0.0.1:80`，不能作为公网入口）：

```bash
docker compose -f docker-compose.prod.yml --env-file .env up -d
```

本地开发启动：先按上文准备 `.env` 与 `configs/config.docker.yaml`（容器内数据库 host 为 `mysql` / `redis`），然后执行：

```bash
docker compose --env-file .env up -d
```

开发编排的 MySQL、Redis、API、Nginx 端口均只绑定到 `127.0.0.1`，不能当作公网部署方案。

生产编排里 `tuyou-api` 已声明好环境变量映射（`MYSQL_PASSWORD` / `REDIS_PASSWORD` /
`JWT_SECRET` / 短信与 OSS 四项），因此 `config.docker.yaml` 中**不必再写真实凭据**。

### 方式 B：直接运行二进制

```bash
set -a && source .env && set +a
./bin/tuyou-api -c configs/config.yaml
```

---

## 六、Nginx 反向代理

前端为 SPA（history 路由）时，关键配置：

```nginx
server {
    listen 80;
    server_name your-domain.example;
    client_max_body_size 500m;          # 必配：否则视频上传被 413 拒绝

    location / {
        root /usr/share/nginx/html;
        index index.html;
        try_files $uri $uri/ /index.html;
    }

    location /api/ {
        proxy_pass http://tuyou-api:9300;
        proxy_set_header Host              $host;
        proxy_set_header X-Real-IP         $remote_addr;
        proxy_set_header X-Forwarded-For   $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_read_timeout 300s;

        # SSE 长连接：必须关闭缓冲，否则事件会被 Nginx 攒着不发
        proxy_buffering off;
        proxy_cache off;
    }
}
```

> 服务端已在 SSE 响应中设置 `X-Accel-Buffering: no`，与上面的 `proxy_buffering off` 双重保险。

启用 HTTPS 后，把 `listen 80` 改为跳转，另起 `listen 443 ssl`（证书路径按实际填写），并补充安全响应头
（`X-Content-Type-Options`、`X-Frame-Options`、`Strict-Transport-Security`）。

---

## 七、部署后验证

```bash
# 1. 生产环境通过 Nginx 入口健康检查（API 的 9300 没有映射到宿主机）
curl -f http://127.0.0.1/api/v1/health
# 期望：{"status":"ok","message":"TuYou API is running"}；公网发布前必须启用 HTTPS

# 2. 容器状态
docker compose -f docker-compose.prod.yml --env-file .env ps

# 3. 应用日志（确认无 Fatal，尤其是敏感词库加载与存储初始化；空词库未必报错，需单独检查表）
docker compose -f docker-compose.prod.yml --env-file .env logs tuyou-api

# 4. 确认数据库未暴露到宿主机
ss -lntp | grep -E '3306|3308|6379'
# 期望：无输出（生产环境下不应监听在这些端口）
```

> 启动日志中若出现「敏感词库加载失败」或「存储初始化失败」，进程会直接退出——
> 这是刻意设计（`logger.Fatal`），表示依赖不满足时**不允许带病启动**。

---

## 八、回滚（待实际演练）

当前生产编排使用 `build:` 直接从工作区构建应用镜像，**没有固定镜像版本或经过验证的回滚脚本**。不要把重新运行 `docker compose up -d` 当作回滚——它会使用当前工作区代码。上线前需先定义固定版本的镜像和上一个版本的保留策略，在服务器上实际演练应用回退；数据库变更及备份恢复还需单独设计和演练。

---

## 九、待补项（当前缺口）

以下能力目前尚未落地，上线前需补齐：

| 项目 | 说明 |
|---|---|
| 数据库定时备份 | 目前数据卷已持久化，但没有定时 `mysqldump` 与异地保留 |
| 监控与告警 | 无指标采集（`/metrics`）、无告警渠道 |
| 版本化迁移 | 表结构依赖启动时 `AutoMigrate`，缺少可回滚的迁移版本文件 |
| CI/CD | 已有待提交的 GitLab CI 检查文件，但尚未在 Runner 上验证；部署仍为手工，无自动回滚 |
| 固定版本与回滚演练 | 当前 compose 使用 `build:`；需固定镜像版本并实际演练应用/数据回退 |
| 公网前置条件 | 域名与 HTTPS 证书、接口限流、CORS 白名单、真实容器首次启动验证仍未完成 |
| 前端产物 | Nginx 挂载的 `frontend/dist` 需由前端项目构建产出 |
| 敏感词种子 | 启动不会自动插入敏感词，需对全新库按 README 手动导入，并检查启用词数量；`sql/seed_data.sql` 会清空业务表，严禁用于已有数据的数据库 |
