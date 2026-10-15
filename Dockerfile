# ============ 构建阶段 ============
FROM golang:1.25-alpine AS builder

# 国内镜像加速（海外服务器可删掉这两行）
ENV GOPROXY=https://goproxy.cn,direct
ENV GO111MODULE=on

WORKDIR /build

# 先复制依赖文件，利用 Docker 缓存层
COPY go.mod go.sum ./
RUN go mod download

# 复制源码并编译
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o tuyou-api ./cmd/server/main.go

# ============ 运行阶段 ============
FROM alpine:3.20

# 时区 + CA 证书（调阿里云 API 需要）
RUN apk add --no-cache tzdata ca-certificates

ENV TZ=Asia/Shanghai

WORKDIR /app

# 从构建阶段复制二进制
COPY --from=builder /build/tuyou-api .
# 复制配置文件
COPY --from=builder /build/configs ./configs

EXPOSE 9300

CMD ["./tuyou-api", "-c", "configs/config.docker.yaml"]
