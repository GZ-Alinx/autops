# 第一阶段：构建阶段
FROM golang:1.23-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制go.mod和go.sum文件
COPY go.mod go.sum ./
# 下载依赖
RUN go mod tidy
# 复制源代码
COPY . .
# 构建项目
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o autops main.go

# 第二阶段：运行阶段
FROM alpine:latest

# 设置工作目录
WORKDIR /app

# 从构建阶段复制可执行文件
COPY --from=builder /app/autops .

# 复制配置文件
COPY config.yaml .

# 暴露端口（根据应用实际监听端口调整）
EXPOSE 8080

# 运行应用
CMD ["./autops"]