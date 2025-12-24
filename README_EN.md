# Clash Speedtest

A simple and easy-to-use Clash proxy node speed testing tool with support for speed testing, latency monitoring, and visual output.

## ✨ Features

- 🌐 **Web UI** - Visual interface, no need to memorize commands
- 🚀 **Multiple Speed Test Methods** - Supports Cloudflare, Speedtest.net, Fast.com
- ⚡ **Concurrent Testing** - Test multiple nodes simultaneously, up to 5x faster
- 📊 **Multiple Output Formats** - Supports text table and image output
- 🏓 **Latency Monitoring** - Real-time Ping testing with min/max/avg statistics
- 🎯 **Flexible Filtering** - Support include/exclude rules for node names
- 🎨 **Colored Output** - Auto-colored based on speed and latency

## 📦 Installation

### Download Pre-built Binary

Download the appropriate version for your system from the [Releases](https://github.com/starudream/clash-speedtest/releases) page.

### Build from Source

```bash
git clone https://github.com/starudream/clash-speedtest.git
cd clash-speedtest
go build -ldflags '-s -w' -o clash-speedtest ./cmd
```

## 🚀 Quick Start

### Method 1: Web UI (Recommended)

**Windows**:
```bash
# Double-click to run
start-web.bat

# Or start from command line
clash-speedtest.exe web --port 8080
```

**Other Systems**:
```bash
./clash-speedtest web --port 8080
```

Then visit `http://localhost:8080` in your browser.

### Method 2: Command Line

```bash
# Test all nodes
clash-speedtest -y

# Test nodes containing "HK"
clash-speedtest -i "HK" -y

# Test 2 nodes concurrently (faster)
clash-speedtest -i "HK" -n 2 -y

# Test and output as image
clash-speedtest -i "US" -f png -y

# Full configuration: concurrent + ping + image
clash-speedtest -i "JP" -n 2 -p -f png -y
```

## 📖 Command Line Options

### Basic Options

```bash
--clash-addr string      Clash external controller address (default "http://127.0.0.1:9090")
--clash-secret string    Clash external controller secret
--clash-proxy string     Clash proxy address (http or socks5)
-y, --confirm            Auto confirm to start test
```

### Speed Test Configuration

```bash
-d, --download string    Speed test method: cloudflare, speedtest, fast (default "cloudflare")
-t, --threads int        Download threads (default 1)
-r, --size int           Download size per thread in MB (default 10)
-n, --concurrent int     Number of nodes to test concurrently 1-5 (default 1)
```

### Latency Test

```bash
-p, --ping               Enable periodic ping test
--ping-interval int      Ping interval in seconds (default 60)
--ping-timeout int       Ping timeout in milliseconds (default 5000)
```

### Node Filtering

```bash
-i, --includes strings   Include node names (supports multiple)
-e, --excludes strings   Exclude node names (supports multiple)
```

### Output Configuration

```bash
-f, --format string      Output format: txt or png (default "txt")
-o, --output string      Output file path (default "output")
```

## 📝 Usage Examples

### Basic Testing

```bash
# Test all nodes
clash-speedtest -y

# Test Hong Kong nodes
clash-speedtest -i "HK" -y

# Test Hong Kong or US nodes
clash-speedtest -i "HK" -i "US" -y
```

### Concurrent Testing

```bash
# Test 2 nodes simultaneously (recommended)
clash-speedtest -i "HK" -n 2 -y

# Test 3 nodes simultaneously (faster)
clash-speedtest -i "US" -n 3 -y
```

### Latency Monitoring

```bash
# Enable ping test
clash-speedtest -i "JP" -p -y

# Custom ping interval
clash-speedtest -i "SG" -p --ping-interval 30 -y
```

### Image Output

```bash
# Basic image output
clash-speedtest -i "UK" -f png -y

# Image + ping
clash-speedtest -i "DE" -p -f png -y
```

### Full Configuration

```bash
clash-speedtest \
  --clash-addr "http://127.0.0.1:9090" \
  --clash-secret "your-secret" \
  -i "JP" \
  -n 2 \
  -p \
  -f png \
  -y
```

## 📊 Output Description

### Text Output

```
+----------+--------+---------------+------------+
|   NAME   |  TYPE  |      IP       |    DOWN    |
+----------+--------+---------------+------------+
| 🇭🇰HK-01   | Vless  | 157.254.20.4  | 2.47 MB/s  |
| 🇺🇸US-01   | Trojan | 104.238.220.1 | 1.50 MB/s  |
+----------+--------+---------------+------------+
```

### With Ping Enabled

```
+----------+--------+----------+----------+----------+------------+
|   NAME   |  TYPE  | PING-MIN | PING-MAX | PING-AVG |    DOWN    |
+----------+--------+----------+----------+----------+------------+
| 🇭🇰HK-01   | Vless  |   45ms   |   52ms   |   48ms   | 2.47 MB/s  |
| 🇺🇸US-01   | Trojan |  156ms   |  178ms   |  165ms   | 1.50 MB/s  |
+----------+--------+----------+----------+----------+------------+
```

### Color Scheme

**Latency**:
- 🟢 Green: < 100ms
- 🟡 Yellow: 100-300ms
- 🔴 Red: > 300ms

**Speed**:
- 🟢 Green: > 5 MB/s
- 🟡 Yellow: 1-5 MB/s
- 🔴 Red: < 1 MB/s

## 🔧 Clash Configuration

Ensure external controller is enabled in your Clash config file:

```yaml
# External Controller
external-controller: 127.0.0.1:9090
secret: "your-secret-key"  # Optional

# Proxy ports (choose one)
port: 7890          # HTTP
socks-port: 7891    # SOCKS5
mixed-port: 7892    # Mixed port
```

## ❓ FAQ

### Connection Failed

- Check if Clash is running
- Verify `--clash-addr` is correct
- If secret is set, use `--clash-secret` parameter

### All Nodes Cannot Connect

- Verify `--clash-proxy` address is correct
- Check if Clash proxy port is enabled

### Ping Shows N/A

- Ensure `-p` parameter is used
- Test duration needs to be long enough (at least 30 seconds)

## 🛠️ Development

### Project Structure

```
clash-speedtest/
├── api/          # API clients
├── cmd/          # Command line entry
├── emoji/        # Emoji handling
├── job/          # Core task logic
├── util/         # Utility functions
└── web/          # Web UI
```

### Build

```bash
# Windows
go build -ldflags '-s -w' -o clash-speedtest.exe ./cmd

# Linux
GOOS=linux GOARCH=amd64 go build -ldflags '-s -w' -o clash-speedtest ./cmd

# macOS
GOOS=darwin GOARCH=amd64 go build -ldflags '-s -w' -o clash-speedtest ./cmd
```

## 📄 License

[MIT License](./LICENSE)

## 🙏 Acknowledgments

- [Cloudflare Speed Test](https://speed.cloudflare.com)
- [Speedtest.net](https://www.speedtest.net)
- [Fast.com](https://fast.com)
- [fogleman/gg](https://github.com/fogleman/gg)

---

**⭐ If this project helps you, please give it a Star!**
