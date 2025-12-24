# Clash Speedtest

![golang](https://img.shields.io/github/actions/workflow/status/starudream/clash-speedtest/golang.yml?style=for-the-badge&logo=github&label=golang)
![release](https://img.shields.io/github/v/release/starudream/clash-speedtest?style=for-the-badge)
![license](https://img.shields.io/github/license/starudream/clash-speedtest?style=for-the-badge)

一个功能强大的 Clash 代理节点测速工具，支持速度测试、延迟监控和可视化输出。

## ✨ 特性

- 🌐 **Web UI 界面**：全新的 Web 界面，支持所有配置项的可视化操作
- 🚀 **多种测速方式**：支持 Cloudflare、Speedtest.net、Fast.com
- ⚡ **并发测速**：支持多节点同时测试，大幅提升测速效率（最高 5 倍速度）
- 📊 **多种输出格式**：支持文本表格和精美图片输出
- 🏓 **Ping 延迟监控**：支持周期性 ping 测试，显示最小值、最大值和平均值
- 🎯 **灵活的节点筛选**：支持包含/排除规则
- 🌈 **多协议支持**：支持 Shadowsocks、VMess、Trojan、VLESS、Hysteria2、AnyTLS 等
- 🎨 **彩色输出**：根据速度/延迟自动着色，一目了然
- 📱 **实时监控**：Web UI 实时显示测速进度和日志

## 📦 安装

### 下载预编译版本

从 [Releases](https://github.com/starudream/clash-speedtest/releases) 页面下载适合您系统的版本。

### 从源码编译

```bash
git clone https://github.com/starudream/clash-speedtest.git
cd clash-speedtest
go build -ldflags '-s -w' -o clash-speedtest.exe ./cmd
```

## 🚀 快速开始

### 🌐 Web UI 模式（推荐）

**Windows 用户**：
```bash
# 方法 1: 双击运行 start-web.bat（推荐）
start-web.bat

# 方法 2: 命令行启动
clash-speedtest.exe web --port 8080
```

**其他系统**：
```bash
# 启动 Web 服务器
./clash-speedtest web --port 8080

# 在浏览器中访问
http://localhost:8080
```

**Web UI 特性**：
- ✅ 可视化配置所有参数
- ✅ 实时显示测速进度
- ✅ 彩色结果展示
- ✅ 运行日志实时输出
- ✅ 无需记忆命令行参数

详细使用说明请查看 [Web UI 使用指南](./WEB_UI_README.md)

---

### 💻 CLI 模式
```

## 🚀 快速开始

### 🌐 Web UI 模式（推荐）

**Windows 用户**：
```bash
# 方法 1: 双击运行 start-web.bat（推荐）
start-web.bat

# 方法 2: 命令行启动
clash-speedtest.exe web --port 8080
```

**其他系统**：
```bash
# 启动 Web 服务器
./clash-speedtest web --port 8080

# 在浏览器中访问
http://localhost:8080
```

**Web UI 特性**：
- ✅ 可视化配置所有参数
- ✅ 实时显示测速进度
- ✅ 彩色结果展示
- ✅ 运行日志实时输出
- ✅ 无需记忆命令行参数

详细使用说明请查看 [Web UI 使用指南](./WEB_UI_README.md)

---

### 💻 CLI 模式

#### 基础用法

```bash
# 测试所有节点
clash-speedtest -y

# 测试包含"香港"的节点
clash-speedtest -i "香港" -y

# 并发测试 2 个节点（推荐，更快）
clash-speedtest -i "香港" -n 2 -y

# 测试并输出为图片
clash-speedtest -i "美国" -f png -y

# 测试并启用 ping 监控
clash-speedtest -i "日本" -p -f png -y

# 完整配置：并发 + Ping + 图片
clash-speedtest -i "日本" -n 2 -p -f png -y
```

## 📖 详细用法

```bash
> clash-speedtest -h
Usage:
  clash-speedtest [flags]

Flags:
      --clash-addr string     clash external controller address (default "http://127.0.0.1:9090")
      --clash-proxy string    clash proxy url, http or socks5
      --clash-secret string   clash external controller secret
  -y, --confirm               confirm to start speedtest
  -d, --download string       download type: cloudflare, speedtest, fast (default "cloudflare")
  -e, --excludes strings      exclude proxy names, after filter by include
  -f, --format string         output format: txt or png (default "txt")
  -h, --help                  help for clash-speedtest
  -i, --includes strings      include proxy names
  -o, --output string         output file path (default "output")
  -n, --concurrent int        number of nodes to test concurrently (1-5) (default 1)
  -p, --ping                  enable periodic ping test during speedtest
      --ping-interval int     ping interval in seconds (default 60)
      --ping-timeout int      ping timeout in milliseconds (default 5000)
  -r, --size uint16           download size for each thread (default 10)
  -t, --threads uint16        download threads for each type, each thread will took 2MB traffic (default 1)
  -v, --version               version for clash-speedtest
```

### 参数说明

#### 连接配置

- **`--clash-addr`**
  - Clash 外部控制器地址
  - 默认：`http://127.0.0.1:9090`
  - 示例：`--clash-addr "http://127.0.0.1:9097"`

- **`--clash-secret`**
  - Clash 外部控制器密钥
  - 在 Clash 配置文件中的 `external-controller` 和 `secret` 字段
  - 示例：`--clash-secret "your-secret-key"`

- **`--clash-proxy`**
  - Clash 代理地址，用于下载测速
  - 通常为 `http://127.0.0.1:7890`（HTTP）或 `socks5://127.0.0.1:7891`（SOCKS5）
  - 可通过 Clash API `/configs` 获取 `port`（HTTP）、`socks-port`（SOCKS5）或 `mixed-port`
  - 如果配置了认证，格式：`http://user:pass@127.0.0.1:7890`

#### 测速配置

- **`-d, --download`**
  - 下载测速方式
  - 可选值：`cloudflare`、`speedtest`、`fast`
  - 默认：`cloudflare`
  - 说明：
    - `cloudflare` - 使用 Cloudflare Speed Test（推荐）
    - `speedtest` - 使用 Speedtest.net
    - `fast` - 使用 Fast.com

- **`-r, --size`**
  - 每个线程的下载大小
  - 单位：MB（cloudflare 和 fast）
  - 默认：10 MB
  - 示例：`-r 15` 表示每个线程下载 15MB

- **`-t, --threads`**
  - 下载线程数（单个节点内的并发）
  - 默认：1
  - 每个线程约消耗 2MB 流量
  - 示例：`-t 3` 表示使用 3 个线程同时下载

- **`-n, --concurrent`**
  - 并发测速的节点数（多节点同时测速）
  - 默认：1（串行测速）
  - 取值范围：1-5
  - 推荐：2-3（平衡速度和稳定性）
  - 示例：`-n 2` 表示同时测试 2 个节点
  - ⚠️ 注意：
    - 并发数越大，Clash 压力越大
    - 会占用更多网络带宽
    - 建议不要超过 3
    - 并发模式使用日志输出代替进度条

#### Ping 配置

- **`-p, --ping`**
  - 启用周期性 ping 测试
  - 会在后台每隔一定时间 ping 所有节点
  - 最终结果显示最小值、最大值和平均值

- **`--ping-interval`**
  - Ping 间隔时间（秒）
  - 默认：60 秒
  - 示例：`--ping-interval 30` 表示每 30 秒 ping 一次

- **`--ping-timeout`**
  - Ping 超时时间（毫秒）
  - 默认：5000 毫秒（5 秒）
  - 示例：`--ping-timeout 3000` 表示 3 秒超时

#### 节点筛选

- **`-i, --includes`**
  - 包含的代理名称（模糊匹配）
  - 支持多个条件，满足其一即包含
  - 示例：
    ```bash
    # 单个条件
    -i "香港"
    
    # 多个条件（使用多个 -i）
    -i "香港" -i "美国"
    
    # 多个条件（使用逗号分隔）
    -i "香港,美国"
    
    # 使用引号包裹特殊字符
    -i "香港|美国"
    ```

- **`-e, --excludes`**
  - 排除的代理名称（模糊匹配）
  - 在 includes 筛选之后执行
  - 用法同 includes
  - 示例：`-i "美国" -e "Premium"` 表示包含"美国"但排除"Premium"

#### 输出配置

- **`-f, --format`**
  - 输出格式
  - 可选值：`txt`（文本表格）、`png`（图片）
  - 默认：`txt`
  - 示例：`-f png` 输出为精美图片

- **`-o, --output`**
  - 输出文件路径
  - 默认：当前目录下的 `output` 文件夹
  - 文件名格式：`20060102-150405.txt` 或 `20060102-150405.png`
  - 示例：`-o "results"` 输出到 `results` 文件夹

- **`-y, --confirm`**
  - 自动确认开始测试
  - 不加此参数需要手动确认

## 📝 使用示例

### 示例 1：快速测试所有节点

```bash
clash-speedtest -y
```

### 示例 2：测试特定地区节点

```bash
# 测试所有香港节点
clash-speedtest -i "香港" -y

# 测试香港或美国节点
clash-speedtest -i "香港" -i "美国" -y

# 测试美国节点但排除 Premium
clash-speedtest -i "美国" -e "Premium" -y
```

### 示例 3：使用 Ping 监控

```bash
# 启用 ping，文本输出
clash-speedtest -i "日本" -p -y

# 启用 ping，图片输出
clash-speedtest -i "新加坡" -p -f png -y

# 自定义 ping 间隔为 30 秒
clash-speedtest -i "台湾" -p --ping-interval 30 -y
```

### 示例 4：图片输出

```bash
# 基础图片输出
clash-speedtest -i "英国" -f png -y

# 图片输出 + Ping 监控
clash-speedtest -i "德国" -p -f png -y

# 完整配置
clash-speedtest \
  --clash-addr "http://127.0.0.1:9097" \
  --clash-secret "your-secret" \
  -d cloudflare \
  -t 1 \
  -r 15 \
  -p \
  --ping-interval 30 \
  -f png \
  -i "日本" \
  -o "results" \
  -y
```

### 示例 5：并发测速（多节点同时测试）

```bash
# 同时测试 2 个节点（推荐）
clash-speedtest -i "香港" -n 2 -y

# 同时测试 3 个节点（更快）
clash-speedtest -i "美国" -n 3 -y

# 并发 + Ping + 图片输出
clash-speedtest -i "日本" -n 2 -p -f png -y

# 性能对比：
# 串行（n=1）: 10个节点 × 30秒 = 5分钟
# 并发（n=2）: 5轮 × 30秒 = 2.5分钟（快2倍）
# 并发（n=3）: 4轮 × 30秒 = 2分钟（快2.5倍）
```

### 示例 6：高流量测试

```bash
# 使用 3 个线程，每线程 20MB
clash-speedtest -i "香港" -t 3 -r 20 -y
```

### 示例 7：使用不同测速服务

```bash
# 使用 Cloudflare（推荐）
clash-speedtest -i "美国" -d cloudflare -y

# 使用 Speedtest.net
clash-speedtest -i "美国" -d speedtest -y

# 使用 Fast.com
clash-speedtest -i "美国" -d fast -y
```

## 📊 输出说明

### 文本输出（不启用 ping）

```
+----------+--------+---------------+------------+
|   NAME   |  TYPE  |      IP       |    DOWN    |
+----------+--------+---------------+------------+
| 🇭🇰香港01  | Vless  | 157.254.20.4  | 2.47 MB/s  |
| 🇺🇸美国01  | Trojan | 104.238.220.1 | 1.50 MB/s  |
+----------+--------+---------------+------------+
```

### 文本输出（启用 ping）

```
+----------+--------+---------------+----------+----------+----------+------------+
|   NAME   |  TYPE  |      IP       | PING-MIN | PING-MAX | PING-AVG |    DOWN    |
+----------+--------+---------------+----------+----------+----------+------------+
| 🇭🇰香港01  | Vless  | 157.254.20.4  |   45ms   |   52ms   |   48ms   | 2.47 MB/s  |
| 🇺🇸美国01  | Trojan | 104.238.220.1 |  156ms   |  178ms   |  165ms   | 1.50 MB/s  |
+----------+--------+---------------+----------+----------+----------+------------+
```

### 图片输出

图片输出会生成一张精美的测速结果图，包含：
- 节点名称、类型、IP 地址
- Ping 延迟统计（如果启用）：最小值、最大值、平均值
- 下载速度
- 根据性能自动着色（绿色=优秀、黄色=一般、红色=较差）

### 列说明

- **NAME**：节点名称（通常包含国旗 emoji）
- **TYPE**：代理协议类型（Shadowsocks、VMess、Trojan、VLESS、Hysteria2、AnyTLS 等）
- **IP**：节点 IP 地址
- **PING-MIN**：最小延迟（启用 ping 时显示）
- **PING-MAX**：最大延迟（启用 ping 时显示）
- **PING-AVG**：平均延迟（启用 ping 时显示）
- **DOWN**：下载速度（总下载量 / 响应时间）
  - 如果显示"无法连接"，表示该节点无法正常连接

## 🎨 颜色说明

### Ping 延迟
- 🟢 绿色：< 100ms（优秀）
- 🟡 黄色：100-300ms（一般）
- 🔴 红色：> 300ms（较差）

### 下载速度
- 🟢 绿色：> 5 MB/s（优秀）
- 🟡 黄色：1-5 MB/s（一般）
- 🔴 红色：< 1 MB/s（较慢）

## 🔧 Clash 配置要求

确保您的 Clash 配置文件（`config.yaml`）中启用了外部控制器：

```yaml
# External Controller API
external-controller: 127.0.0.1:9090
secret: "your-secret-key"  # 可选，建议设置

# 代理端口（任选其一）
port: 7890                  # HTTP 代理端口
socks-port: 7891           # SOCKS5 代理端口
mixed-port: 7892           # 混合端口（HTTP + SOCKS5）
```

## 📌 常见问题

### 1. 连接 Clash 失败

**问题**：`connection refused` 或 `unauthorized`

**解决**：
- 检查 Clash 是否正在运行
- 确认 `--clash-addr` 地址和端口正确
- 如果设置了 secret，使用 `--clash-secret` 参数

### 2. 所有节点显示"无法连接"

**问题**：下载速度全部为 0 或"无法连接"

**解决**：
- 确认 `--clash-proxy` 地址正确
- 检查 Clash 的代理端口是否开启
- 尝试在浏览器中配置代理访问网站测试

### 3. Ping 显示 N/A

**问题**：启用 `-p` 后 ping 值全是 N/A

**解决**：
- 确保使用了 `-p` 参数
- 测试时间太短（建议至少 30 秒）
- 节点可能无法连接

### 4. 图片输出失败

**问题**：使用 `-f png` 但没有生成图片

**解决**：
- 检查 output 目录是否有写入权限
- 查看控制台错误信息
- 会自动降级到文本输出

## 🛠️ 开发

### 项目结构

```
clash-speedtest/
├── api/              # API 客户端
│   ├── clash/        # Clash API
│   ├── cloudflare/   # Cloudflare Speed Test
│   ├── fast/         # Fast.com
│   └── speedtest/    # Speedtest.net
├── cmd/              # 命令行入口
├── emoji/            # Emoji 处理
├── job/              # 核心任务逻辑
├── util/             # 工具函数
└── README.md
```

### 编译

```bash
# 标准编译
go build -o clash-speedtest ./cmd

# 压缩编译
go build -ldflags '-s -w' -o clash-speedtest ./cmd

# 交叉编译（Windows）
GOOS=windows GOARCH=amd64 go build -ldflags '-s -w' -o clash-speedtest.exe ./cmd

# 交叉编译（Linux）
GOOS=linux GOARCH=amd64 go build -ldflags '-s -w' -o clash-speedtest ./cmd

# 交叉编译（macOS）
GOOS=darwin GOARCH=amd64 go build -ldflags '-s -w' -o clash-speedtest ./cmd
```

### 运行测试

```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./api/clash/
```

## 🙏 致谢

- [Cloudflare Speed Test](https://speed.cloudflare.com)
- [Speedtest.net](https://www.speedtest.net)
- [Fast.com](https://fast.com)
- [go-resty trace info](https://vearne.cc/archives/39953)
- [fogleman/gg](https://github.com/fogleman/gg) - 图片渲染

## 📄 许可证

[MIT License](./LICENSE)

---

**⭐ 如果这个项目对您有帮助，请给个 Star！**
