# 多节点并发测速功能更新

## 更新内容

已成功实现多节点并发测速功能，现在可以同时测试多个节点，大幅提升测速效率！

## 主要改动

### 1. 代码修改
- **job/task.go**: 添加了工作池（worker pool）并发模式
  - 新增 `ConcurrentNodes` 字段控制并发数
  - 使用 goroutine 和 channel 实现并发测速
  - 限制最大并发数为 5，避免 Clash 压力过大

- **cmd/root.go**: 添加了 `--concurrent` / `-c` 参数
  - 默认值：1（串行测速，与之前行为一致）
  - 取值范围：1-5

- **speedtest.bat**: 添加了 `CONCURRENT_NODES` 配置
  - 默认值：2（同时测试2个节点）

### 2. 新增参数

```bash
-c, --concurrent int   number of nodes to test concurrently (1-5) (default 1)
```

## 使用方法

### 1. 重新编译

由于 Go 环境未在当前 PowerShell 中配置，请按以下步骤重新编译：

```bash
# 方法1: 使用 Go 直接编译
go build -ldflags "-s -w" -o clash-speedtest.exe ./cmd

# 方法2: 使用 Makefile（如果有 make）
make build
```

### 2. 命令行使用

#### 串行测速（默认，与之前一致）
```bash
clash-speedtest.exe -i "HK" -y
```

#### 并发测速 - 同时测试 2 个节点
```bash
clash-speedtest.exe -i "HK" -c 2 -y
```

#### 并发测速 - 同时测试 3 个节点
```bash
clash-speedtest.exe -i "US" -c 3 -y
```

#### 完整示例（带 ping 和图片输出）
```bash
clash-speedtest.exe -i "JP" -c 2 -p -f png -y
```

### 3. 使用 speedtest.bat

修改 `speedtest.bat` 中的配置：

```batch
set CONCURRENT_NODES=2  REM 同时测试2个节点（推荐）
set CONCURRENT_NODES=3  REM 同时测试3个节点（较快）
set CONCURRENT_NODES=1  REM 串行测试（最慢，最稳定）
```

然后运行：
```bash
speedtest.bat HK
```

## 性能对比

假设有 10 个节点，每个节点测速需要 30 秒：

| 并发数 | 预计耗时 | 效率提升 |
|--------|----------|----------|
| 1（串行） | 10 × 30s = 5分钟 | 基准 |
| 2（并发） | 5 × 30s = 2.5分钟 | **2倍** |
| 3（并发） | 4 × 30s = 2分钟 | **2.5倍** |
| 5（并发） | 2 × 30s = 1分钟 | **5倍** |

## 推荐配置

- **一般用户**: `CONCURRENT_NODES=2`
  - 平衡速度和稳定性
  - 适合大多数场景

- **追求速度**: `CONCURRENT_NODES=3-4`
  - 测速更快
  - 需要 Clash 性能较好

- **保守模式**: `CONCURRENT_NODES=1`
  - 最稳定
  - 与之前版本行为完全一致

## 注意事项

1. **并发数限制**: 最大为 5，超过会自动限制到 5
2. **Clash 压力**: 并发数越大，Clash 压力越大，建议不要超过 3
3. **网络带宽**: 多节点并发会占用更多带宽
4. **日志输出**: 日志中会显示 `worker X` 标识不同的工作线程

## 日志示例

串行模式（concurrent=1）：
```
INF start proxy test: 🇭🇰香港1 index=1/3
INF proxy test done: 🇭🇰香港1 index=1/3
INF start proxy test: 🇭🇰香港2 index=2/3
```

并发模式（concurrent=2）：
```
INF worker 1 start proxy test: 🇭🇰香港1
INF worker 2 start proxy test: 🇭🇰香港2  # 同时进行
INF worker 1 proxy test done: 🇭🇰香港1
INF worker 1 start proxy test: 🇭🇰香港3
```

## 故障排查

如果遇到问题，可以尝试：
1. 降低并发数到 1 或 2
2. 检查 Clash 是否正常运行
3. 查看日志中的错误信息

## 完成状态

✅ 代码修改完成
✅ 参数添加完成
✅ 批处理脚本更新完成
⏳ 需要重新编译后测试

---

**下一步**: 请重新编译项目，然后测试并发功能！

