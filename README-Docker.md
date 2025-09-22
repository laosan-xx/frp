# FRP GoReleaser Docker 环境

这个 Docker 环境包含了运行 FRP 项目 GoReleaser 工作流所需的所有依赖。

## 包含的依赖

- **Ubuntu 22.04 LTS** 基础镜像
- **Go 1.23.4** 开发环境
- **Git** 版本控制
- **Make** 构建工具
- **交叉编译工具** (gcc-multilib, g++-multilib)
- **打包工具** (zip, tar)

## 文件说明

- `Dockerfile.goreleaser` - Docker 镜像定义文件
- `docker-compose.goreleaser.yml` - Docker Compose 配置文件
- `build-docker.sh` - 构建脚本

## 使用方法

### 方法一：直接使用 Docker

1. 构建镜像：
```bash
docker build -f Dockerfile.goreleaser -t frp-goreleaser:latest .
```

2. 运行容器：
```bash
docker run -it --rm -v $(pwd):/workspace frp-goreleaser:latest
```

### 方法二：使用 Docker Compose

1. 启动服务：
```bash
docker-compose -f docker-compose.goreleaser.yml up -d
```

2. 进入容器：
```bash
docker-compose -f docker-compose.goreleaser.yml exec frp-builder bash
```

3. 停止服务：
```bash
docker-compose -f docker-compose.goreleaser.yml down
```

### 方法三：使用构建脚本

```bash
chmod +x build-docker.sh
./build-docker.sh
```

## 在容器中执行构建

进入容器后，可以执行以下命令：

```bash
# 构建项目
./package.sh

# 或者分步执行
make
make -f ./Makefile.cross-compiles

# 运行 GoReleaser（需要设置 GITHUB_TOKEN）
export GITHUB_TOKEN=your_token_here
goreleaser release --clean --release-notes=./Release.md
```

## 注意事项

1. 确保 Docker 已安装并运行
2. 如果需要访问网络（下载依赖），容器会使用 host 网络模式
3. Go 模块缓存和构建缓存会持久化存储
4. 项目文件会挂载到容器的 `/workspace` 目录

## 环境变量

- `GO_VERSION=1.23.4` - Go 版本
- `GOPATH=/go` - Go 工作目录
- `GO111MODULE=on` - 启用 Go 模块
- `CGO_ENABLED=0` - 禁用 CGO（用于静态编译）
