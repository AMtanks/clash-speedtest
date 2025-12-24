# Clash Speedtest

一个简单易用的 Clash 代理节点测速工具，支持速度测试、延迟监控和可视化输出。

## ✨ 特性

- 🌐 **Web UI 界面** - 可视化操作，无需记忆命令
- 🚀 **多种测速方式** - 支持 Cloudflare、Speedtest.net、Fast.com
- ⚡ **并发测速** - 同时测试多个节点，最高提速 5 倍
- 📊 **多种输出格式** - 支持文本表格和图片输出
- 🏓 **延迟监控** - 实时 Ping 测试，显示最小/最大/平均延迟
- 🎯 **灵活筛选** - 支持节点名称包含/排除规则
- 🎨 **彩色输出** - 根据速度和延迟自动着色

## 📦 安装

### 下载预编译版本

从 [Releases](https://github.com/starudream/clash-speedtest/releases) 页面下载适合您系统的版本。

### 从源码编译

```bash
git clone https://github.com/starudream/clash-speedtest.git
cd clash-speedtest
go build -ldflags '-s -w' -o clash-speedtest ./cmd
```

## 🚀 快速开始

### 方式一：Web UI（推荐）

**Windows 用户**：
```bash
# 双击运行
start-web.bat

# 或命令行启动
clash-speedtest.exe web --port 8080
```

**其他系统**：
```bash
./clash-speedtest web --port 8080
```

然后在浏览器中访问 `http://localhost:8080`

### 方式二：命令行

```bash
# 测试所有节点
clash-speedtest -y

# 测试包含"香港"的节点
clash-speedtest -i "香港" -y

# 并发测试 2 个节点（更快）
clash-speedtest -i "香港" -n 2 -y

# 测试并输出为图片
clash-speedtest -i "美国" -f png -y

# 完整配置：并发 + Ping + 图片
clash-speedtest -i "日本" -n 2 -p -f png -y
```

## 📖 命令行参数

### 基础参数

```bash
--clash-addr string      Clash 外部控制器地址 (默认 "http://127.0.0.1:9090")
--clash-secret string    Clash 外部控制器密钥
--clash-proxy string     Clash 代理地址 (http 或 socks5)
-y, --confirm            自动确认开始测试
```

### 测速配置

```bash
-d, --download string    测速方式: cloudflare, speedtest, fast (默认 "cloudflare")
-t, --threads int        下载线程数 (默认 1)
-r, --size int           每线程下载大小 MB (默认 10)
-n, --concurrent int     并发测试节点数 1-5 (默认 1)
```

### 延迟测试

```bash
-p, --ping               启用周期性 Ping 测试
--ping-interval int      Ping 间隔秒数 (默认 60)
--ping-timeout int       Ping 超时毫秒 (默认 5000)
```

### 节点筛选

```bash
-i, --includes strings   包含的节点名称（支持多个）
-e, --excludes strings   排除的节点名称（支持多个）
```

### 输出配置

```bash
-f, --format string      输出格式: txt 或 png (默认 "txt")
-o, --output string      输出文件路径 (默认 "output")
```

## 📝 使用示例

### 基础测试

```bash
# 测试所有节点
clash-speedtest -y

# 测试香港节点
clash-speedtest -i "香港" -y

# 测试香港或美国节点
clash-speedtest -i "香港" -i "美国" -y
```

### 并发测速

```bash
# 同时测试 2 个节点（推荐）
clash-speedtest -i "香港" -n 2 -y

# 同时测试 3 个节点（更快）
clash-speedtest -i "美国" -n 3 -y
```

### 延迟监控

```bash
# 启用 Ping 测试
clash-speedtest -i "日本" -p -y

# 自定义 Ping 间隔
clash-speedtest -i "新加坡" -p --ping-interval 30 -y
```

### 图片输出

```bash
# 基础图片输出
clash-speedtest -i "英国" -f png -y

# 图片 + Ping
clash-speedtest -i "德国" -p -f png -y
```

### 完整配置

```bash
clash-speedtest \
  --clash-addr "http://127.0.0.1:9090" \
  --clash-secret "your-secret" \
  -i "日本" \
  -n 2 \
  -p \
  -f png \
  -y
```

## 📊 输出说明

### 文本输出

```
+----------+--------+---------------+------------+
|   NAME   |  TYPE  |      IP       |    DOWN    |
+----------+--------+---------------+------------+
| 🇭🇰香港01  | Vless  | 157.254.20.4  | 2.47 MB/s  |
| 🇺🇸美国01  | Trojan | 104.238.220.1 | 1.50 MB/s  |
+----------+--------+---------------+------------+
```

### 启用 Ping 后

```
+----------+--------+----------+----------+----------+------------+
|   NAME   |  TYPE  | PING-MIN | PING-MAX | PING-AVG |    DOWN    |
+----------+--------+----------+----------+----------+------------+
| 🇭🇰香港01  | Vless  |   45ms   |   52ms   |   48ms   | 2.47 MB/s  |
| 🇺🇸美国01  | Trojan |  156ms   |  178ms   |  165ms   | 1.50 MB/s  |
+----------+--------+----------+----------+----------+------------+
```

### 颜色说明

**延迟**：
- 🟢 绿色：< 100ms
- 🟡 黄色：100-300ms
- 🔴 红色：> 300ms

**速度**：
- 🟢 绿色：> 5 MB/s
- 🟡 黄色：1-5 MB/s
- 🔴 红色：< 1 MB/s

## 🔧 Clash 配置

确保 Clash 配置文件中启用了外部控制器：

```yaml
# External Controller
external-controller: 127.0.0.1:9090
secret: "your-secret-key"  # 可选

# 代理端口（任选其一）
port: 7890          # HTTP
socks-port: 7891    # SOCKS5
mixed-port: 7892    # 混合端口
```

## ❓ 常见问题

### 连接失败

- 检查 Clash 是否运行
- 确认 `--clash-addr` 地址正确
- 如果设置了 secret，使用 `--clash-secret` 参数

### 所有节点无法连接

- 确认 `--clash-proxy` 地址正确
- 检查 Clash 代理端口是否开启

### Ping 显示 N/A

- 确保使用了 `-p` 参数
- 测试时间需要足够长（建议至少 30 秒）

## 🛠️ 开发

### 项目结构

```
clash-speedtest/
├── api/          # API 客户端
├── cmd/          # 命令行入口
├── emoji/        # Emoji 处理
├── job/          # 核心任务逻辑
├── util/         # 工具函数
└── web/          # Web UI
```

### 编译

```bash
# Windows
go build -ldflags '-s -w' -o clash-speedtest.exe ./cmd

# Linux
GOOS=linux GOARCH=amd64 go build -ldflags '-s -w' -o clash-speedtest ./cmd

# macOS
GOOS=darwin GOARCH=amd64 go build -ldflags '-s -w' -o clash-speedtest ./cmd
```

## 📄 许可证

[MIT License](./LICENSE)

## 🙏 致谢

- [Cloudflare Speed Test](https://speed.cloudflare.com)
- [Speedtest.net](https://www.speedtest.net)
- [Fast.com](https://fast.com)
- [fogleman/gg](https://github.com/fogleman/gg)

---

**⭐ 如果这个项目对您有帮助，请给个 Star！**
