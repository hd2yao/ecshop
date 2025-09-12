# 指定多架构支持，保证在 Mac M 系列 (arm64) 和 Linux (amd64) 都能正确构建
FROM --platform=$BUILDPLATFORM golang:1.24

LABEL maintainer="Dysania <dysania@hd2yao>"

# 设置 Go 模块代理和 PATH
ENV GOPROXY=https://goproxy.cn,direct
ENV PATH=$GOPATH/bin:$PATH

# 切换 root 安装依赖
USER root

# 使用国内镜像源 + 安装常用工具 + protobuf-compiler (protoc)
RUN apt-get update && apt-get install -y --no-install-recommends \
    curl unzip git vim screen protobuf-compiler \
 && apt-get clean \
 && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*


# 安装 goctl 和 protoc 插件（固定版本更稳定）
RUN go install github.com/zeromicro/go-zero/tools/goctl@v1.6.4 && \
    go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.34.2 && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1

# 设置工作目录（和 docker-compose 中挂载路径保持一致）
WORKDIR /usr/src/code

# 暴露常用端口（仅声明，不做映射）
EXPOSE 8000 8001 8002 8003 9000 9001 9002 9003
