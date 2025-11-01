# 🚀 快速开始指南

## 📦 准备工作

1. **确保 Clash 正在运行**
   - 检查 Clash 外部控制器地址（默认: `http://127.0.0.1:9090`）
   - 如果设置了密钥，记下密钥值

2. **查看 Clash 配置**
   ```yaml
   # 在你的 Clash 配置文件中找到这些信息
   external-controller: 127.0.0.1:9090  # 控制器地址
   secret: "your-secret-key"            # 密钥（如果有）
   ```

## 🎯 使用测试脚本

### Windows 批处理脚本 (speedtest.bat)

最简单的方式，双击即可运行：

```bash
# 直接运行（测试所有节点）
speedtest.bat

# 测试特定地区
speedtest.bat 香港
speedtest.bat 美国
speedtest.bat 日本
```

**修改配置**：
编辑 `speedtest.bat` 文件，修改以下变量：
```batch
set CLASH_ADDR=http://127.0.0.1:9090
set CLASH_SECRET=your-secret-key
set FORMAT=png
set ENABLE_PING=true
```

### PowerShell 脚本 (speedtest.ps1)

更强大的脚本，支持完整参数：

```powershell
# 查看帮助
.\speedtest.ps1 -Help

# 基础用法
.\speedtest.ps1                                    # 测试所有节点
.\speedtest.ps1 -Filter "香港"                     # 测试香港节点
.\speedtest.ps1 -Filter "美国" -Ping              # 测试美国节点 + Ping
.\speedtest.ps1 -Filter "日本" -Format txt        # 文本输出

# 高级用法
.\speedtest.ps1 -Filter "新加坡" -Ping -Threads 2 -Size 15

# 自定义 Clash 配置
.\speedtest.ps1 -ClashAddr "http://127.0.0.1:9097" -ClashSecret "your-secret"

# 完整配置
.\speedtest.ps1 `
  -Filter "英国" `
  -ClashAddr "http://127.0.0.1:9097" `
  -ClashSecret "your-secret" `
  -Ping `
  -Format png `
  -Threads 2 `
  -Size 20
```

## 🎨 直接使用命令行

```bash
# 基础测试
clash-speedtest.exe -y

# 测试特定地区
clash-speedtest.exe -i "香港" -y

# 图片输出 + Ping
clash-speedtest.exe -i "日本" -p -f png -y

# 完整配置
clash-speedtest.exe \
  --clash-addr "http://127.0.0.1:9097" \
  --clash-secret "your-secret" \
  -i "美国" \
  -p \
  -f png \
  -t 2 \
  -r 15 \
  -y
```

## 📊 输出结果

测试完成后，结果会保存在 `output` 目录：

- **文本格式**: `output/20251027-235959.txt`
- **图片格式**: `output/20251027-235959.png`

## 🎯 推荐配置

### 日常快速测试
```powershell
.\speedtest.ps1 -Filter "香港" -Format png
```

### 完整测试（带 Ping）
```powershell
.\speedtest.ps1 -Filter "日本" -Ping -Format png
```

### 高流量测试
```powershell
.\speedtest.ps1 -Filter "美国" -Threads 3 -Size 20 -Ping -Format png
```

## ❓ 常见问题

### 脚本无法运行（PowerShell）

如果遇到"无法加载脚本"错误，执行：
```powershell
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

### 连接失败

1. 检查 Clash 是否运行
2. 确认控制器地址和端口
3. 如果设置了密钥，确保提供了正确的密钥

### 所有节点无法连接

检查 `--clash-proxy` 参数，默认会自动从 Clash 获取代理端口

## 💡 提示

- 首次运行建议先测试少量节点（使用 `-i` 筛选）
- 启用 Ping 会增加测试时间，但能获得更全面的延迟信息
- 图片输出比文本输出更直观美观
- 默认使用 Cloudflare 测速，速度快且准确

---

更多详细信息请查看 [README.md](README.md)

