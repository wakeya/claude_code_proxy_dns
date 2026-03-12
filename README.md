# Claude Code 透明代理

让 Claude Code 误以为在与官方 API 通信的透明代理服务。

## 功能特性

- 透明代理所有 Claude Code API 请求
- **多供应商管理**：支持配置多个 API 供应商，灵活切换
- **模型映射**：自动将请求模型名转换为供应商支持的模型（如 claude-sonnet-4 → glm-5）
- 自动生成 CA 证书（10年有效期）
- 前端配置页面管理供应商和模型映射
- 密码保护配置页面
- Docker 单容器部署
- 热更新配置无需重启

## 快速开始

### 1. 使用 Docker 部署

**方式一：使用 docker build（推荐）**

```bash
# 克隆项目
git clone <repo-url>
cd claude_code_proxy_dns

# 构建镜像
docker build -t claude_code_proxy_dns .

# 运行容器
docker run -d \
  --name claude_code_proxy_dns \
  -p 443:443 \
  -p 8442:8442 \
  -v ./data:/app/data \
  -e ADMIN_PASSWORD=admin123 \
  --cap-add NET_BIND_SERVICE \
  claude_code_proxy_dns

# 查看日志获取配置提示
docker logs claude_code_proxy_dns
```

**方式二：使用 docker-compose**

```bash
# 克隆项目
git clone <repo-url>
cd claude_code_proxy_dns

# 安装 docker-compose-plugin（如果未安装）
sudo apt-get install docker-compose-plugin

# 启动服务
docker compose up -d

# 查看日志获取配置提示
docker logs claude_code_proxy_dns
```

### 2. 测试构建和验证

```bash
# 1. 测试构建镜像
docker compose build

# 2. 查看构建的镜像
docker images | grep claude_code_proxy_dns

# 3. 启动服务
docker compose up -d

# 4. 查看服务日志
docker compose logs -f

# 5. 验证服务是否正常运行
curl -k https://localhost:8442

# 6. 查看容器状态
docker compose ps
```

### 3. 安装 CA 证书

代理使用自签名 CA 证书，需要在客户端机器上安装信任。

**方式一：设置环境变量（推荐）**

```bash
# 将 ca.crt 路径添加到环境变量
echo 'export NODE_EXTRA_CA_CERTS=/path/to/data/ca.crt' >> ~/.bashrc
source ~/.bashrc
```

**方式二：系统级别安装证书**

**macOS：**

```bash
# 安装证书到系统钥匙串
sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain ./data/ca.crt
```

**Linux (Ubuntu/Debian)：**

```bash
# 复制证书到系统证书目录
sudo cp ./data/ca.crt /usr/local/share/ca-certificates/

# 更新证书库
sudo update-ca-certificates

# 如果看到 "skipping duplicate certificate" 警告，说明证书已安装，无需担心
```

### 4. 配置系统

```bash
# 添加 hosts 映射（将 api.anthropic.com 指向代理服务器）
echo "127.0.0.1 api.anthropic.com" | sudo tee -a /etc/hosts
```

### 5. 访问配置页面

打开浏览器访问: `https://localhost:8442`

默认密码: `admin123`

## 配置说明

### 端口权限

代理服务需要绑定 443 端口（HTTPS 默认端口），这是一个特权端口（< 1024），需要特殊权限。

**Docker 部署：**

`docker-compose.yml` 已配置 `cap_add: NET_BIND_SERVICE`，允许容器内进程绑定特权端口：

```yaml
cap_add:
  - NET_BIND_SERVICE
```

**直接运行（非 Docker）：**

需要授予二进制文件绑定特权端口的能力：

```bash
# 方式一：使用 setcap（推荐）
sudo setcap 'cap_net_bind_service=+ep' ./claude_code_proxy_dns

# 方式二：使用 sudo 运行
sudo ./claude_code_proxy_dns -data ./data
```

**注意：** 不推荐修改为非特权端口（如 8443），因为：
- 客户端请求的是 `https://api.anthropic.com`，默认使用 443 端口
- 通过 hosts 文件将域名解析到 127.0.0.1，但无法改变端口号
- 如果代理监听非 443 端口，客户端请求会失败

如果必须使用非特权端口，需要：
```bash
# 使用 iptables 将 443 端口转发到 8443
sudo iptables -t nat -A PREROUTING -p tcp --dport 443 -j REDIRECT --to-port 8443
sudo iptables -t nat -A OUTPUT -p tcp --dport 443 -j REDIRECT --to-port 8443
```

### 环境变量

| 环境变量 | 说明 | 默认值 |
|---------|------|--------|
| ADMIN_PASSWORD | 管理密码 | admin123 |

### 供应商配置

在前端配置页面添加供应商：

| 字段 | 说明 | 示例 |
|------|------|------|
| 名称 | 供应商显示名称 | 阿里云 DashScope |
| API 地址 | 供应商的 API 端点 | https://dashscope.aliyuncs.com/api/v1/anthropic |
| API Token | 认证密钥 | sk-xxx |
| 模型映射 | 客户端模型 → 供应商模型 | claude-sonnet-4 → qwen-max |

### 客户端配置

客户端 `~/.claude/settings.json` 配置示例：

```json
{
  "env": {
    "ANTHROPIC_BASE_URL": "https://api.anthropic.com",
    "ANTHROPIC_MODEL": "claude-sonnet-4"
  }
}
```

代理会自动将 `claude-sonnet-4` 转换为供应商配置的模型（如 `glm-5`）。

## 端口说明

| 端口 | 用途 |
|------|------|
| 443 | 代理服务入口 |
| 8442 | 配置页面 |

## 项目结构

```
claude_code_proxy_dns/
├── cmd/server/          # 应用入口
├── internal/
│   ├── config/          # 配置管理
│   ├── cert/            # 证书管理
│   ├── proxy/           # 代理服务
│   ├── admin/           # 配置服务
│   └── frontend/        # 前端页面
├── data/                # 数据目录
├── Dockerfile
└── docker-compose.yml
```

## 开发

```bash
# 运行测试
make test

# 本地运行
make run

# 构建
make build
```

## 许可证

MIT License