# 使用官方 Go 镜像作为构建阶段
FROM docker.xuanyuan.me/library/golang:1.24-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制 go.mod 和 go.sum 文件
COPY src/go.mod src/go.sum ./

ENV GOPROXY=https://goproxy.cn,direct

# 下载依赖
RUN go mod download

# 复制源代码
COPY src/ .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# 使用轻量级的 alpine 作为运行阶段
FROM docker.xuanyuan.me/library/alpine:latest

# 镜像信息
ENV TZ=Asia/Shanghai
LABEL name=unicomMonitor
LABEL url=https://github.com/zgcwkjOpenProject/GO_UnicomMonitor

# 更新镜像源并安装基础工具
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
    apk update && \
    apk add --no-cache \
    tzdata \
    ca-certificates && \
    rm -rf /var/cache/apk/*

# 设置工作目录
WORKDIR /root/

# 从构建阶段复制编译好的二进制文件和必要的资源
COPY --from=builder /app/main ./
COPY --from=builder /app/config.json ./
COPY --from=builder /app/static ./static/

# 暴露端口
EXPOSE 25678

# 运行应用
CMD ["./main"]
