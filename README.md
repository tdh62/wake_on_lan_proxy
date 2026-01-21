# Wake-on-LAN 服务

一个简单易用的局域网唤醒（Wake-on-LAN）HTTP服务，提供Web界面用于远程唤醒局域网内的设备。

## 功能特性

- 🌐 Web界面操作，简单直观
- 🚀 支持标准Wake-on-LAN魔术包
- 🔧 灵活的MAC地址格式支持（AA:BB:CC:DD:EE:FF 或 AA-BB-CC-DD-EE-FF）
- 📡 可配置广播地址
- 🐳 Docker支持

## 快速开始

### 本地运行

```bash
# 编译并运行
go run main.go

# 或者先编译
go build -o wol-service
./wol-service
```

访问 http://localhost:24000

### Docker运行

```bash
# 构建镜像
docker build -t wol-service .

# 运行容器
docker run -d -p 24000:24000 --name wol-service wol-service

# 使用host网络模式（推荐，用于局域网广播）
docker run -d --network host --name wol-service wol-service
```

访问 http://localhost:24000

### Docker Compose

```yaml
version: '3.8'
services:
  wol-service:
    image: wol-service
    container_name: wol-service
    network_mode: host
    restart: unless-stopped
```

## 使用说明

1. 在Web界面输入目标设备的MAC地址
2. （可选）输入广播地址，默认为 255.255.255.255
3. 点击"发送唤醒包"按钮
4. 系统会发送魔术包到指定的广播地址

## 前置条件

目标设备需要满足以下条件：

1. 主板BIOS中启用Wake-on-LAN功能
2. 网卡支持WOL功能
3. 设备连接电源（即使关机状态）
4. 网络连接正常

## 技术细节

### 魔术包格式

Wake-on-LAN魔术包由102字节组成：
- 前6个字节：`0xFF 0xFF 0xFF 0xFF 0xFF 0xFF`
- 后96个字节：目标MAC地址重复16次

### 网络协议

- 协议：UDP
- 端口：9（标准WOL端口）
- 广播地址：可配置，默认255.255.255.255

### 项目结构

```
.
├── main.go              # 主程序
├── main_test.go         # 单元测试
├── go.mod               # Go模块文件
├── Dockerfile           # Docker镜像构建文件
├── .dockerignore        # Docker忽略文件
└── .github/
    └── workflows/
        └── docker-build.yml  # GitHub Actions工作流
```

## 许可证

MIT License
