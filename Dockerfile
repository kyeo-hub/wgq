FROM golang:1.21-alpine AS builder

WORKDIR /app

# 安装必要的依赖
RUN apk add --no-cache git

# 复制 go.mod 和 go.sum
COPY go.mod go.sum* ./
RUN go mod download

# 复制源代码
COPY . .

# 编译
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o wgq ./cmd/main.go

# 最终镜像
FROM alpine:latest

RUN apk --no-cache add ca-certificates

# 安装 qwen (需要 Node.js)
RUN apk add --no-cache nodejs npm
RUN npm install -g @qwen-code/qwen-code@latest

WORKDIR /root/

# 从 builder 复制编译好的二进制文件
COPY --from=builder /app/wgq .

# 复制示例配置文件
COPY --from=builder /app/config.example.json .

EXPOSE 8080

CMD ["./wgq", "-config", "config.json"]
