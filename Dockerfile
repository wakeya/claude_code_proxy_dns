# 构建阶段
FROM golang:1.26-alpine AS builder

WORKDIR /app

# 复制依赖文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o claude_code_proxy_dns ./cmd/server

# 运行阶段
FROM alpine:latest

# 安装 CA 证书和时区数据
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# 复制二进制文件
COPY --from=builder /app/claude_code_proxy_dns .

# 创建数据目录
RUN mkdir -p /app/data

# 暴露端口
EXPOSE 443 8442

# 设置环境变量
ENV ADMIN_PASSWORD=admin123

# 启动
ENTRYPOINT ["./claude_code_proxy_dns"]
CMD ["-data", "/app/data"]