# Emoji 和中文显示问题说明

## 问题描述

在生成的 PNG 图片中遇到以下问题：
1. ✅ **已解决**：中文显示为方块 → 现在使用微软雅黑字体，中文正常显示
2. ⚠️ **技术限制**：国旗 emoji 显示为文字图案而不是彩色国旗

## 技术原因

### 为什么国旗不能显示为彩色图标？

1. **gg 库限制**：
   - Go 的 `github.com/fogleman/gg` 图像库只支持加载单个字体
   - 不支持字体 fallback 机制
   - 不支持彩色 emoji (COLR/CPAL 表)

2. **Windows 字体限制**：
   - `seguiemj.ttf` (Segoe UI Emoji)：支持 emoji 但不支持中文
   - `msyh.ttc` (微软雅黑)：支持中文和单色 emoji，但国旗显示为字母或方块
   - Windows 没有同时完美支持中文和彩色 emoji 的单一字体

3. **国旗 emoji 特殊性**：
   - 国旗由两个 Unicode 字符组成（区域指示符）
   - 需要特殊的字体支持才能正确显示为国旗
   - 大多数单色字体将其显示为字母（如 HK、US）

## 当前解决方案

已实现智能字体选择机制，优先级如下：

1. **微软雅黑粗体** (`msyhbd.ttc`) - 推荐
   - ✅ 完美支持中文
   - ✅ 支持大部分 emoji（单色）
   - ⚠️ 国旗显示为字母组合（如 HK、US）

2. **微软雅黑** (`msyh.ttc`)
   - ✅ 支持中文
   - ⚠️ emoji 支持一般

3. **黑体** (`simhei.ttf`)
   - ✅ 支持中文
   - ❌ emoji 支持较差

4. **宋体** (`simsun.ttc`)
   - ✅ 支持中文
   - ❌ 不支持 emoji

## 显示效果对比

### 文本输出（终端）
```
+----------------+-----------+-------------+-------------+
| 🇭🇰香港1        | Hysteria2 | 8.45 MB/s   | 12.30 MB/s  |
| 🇺🇸美国2        | AnyTLS    | 15.67 MB/s  | 18.90 MB/s  |
+----------------+-----------+-------------+-------------+
```
终端输出的 emoji 显示效果取决于终端字体设置。

### PNG 图片输出
```
节点名称           | 类型      | 平均速度    | 最高速度
HK香港1           | Hysteria2 | 8.45 MB/s  | 12.30 MB/s
US美国2           | AnyTLS    | 15.67 MB/s | 18.90 MB/s
```
图片中国旗会显示为字母代码（如 HK、US），但中文正常显示。

## 替代方案

### 方案 1：保持现状（推荐）
- 中文正常显示
- emoji 显示为单色或字母代码
- 不需要额外配置

### 方案 2：安装 Noto 字体（高级用户）

1. 下载并安装 Google Noto 字体：
   - Noto Sans CJK SC（简体中文）: https://github.com/googlefonts/noto-cjk
   - Noto Color Emoji（彩色 emoji）: https://github.com/googlefonts/noto-emoji

2. 修改 `util/image.go` 中的字体路径：
   ```go
   fonts := []string{
       "C:\\Windows\\Fonts\\NotoSansCJKsc-Regular.otf",  // Noto 中文字体
       "C:\\Windows\\Fonts\\NotoColorEmoji.ttf",         // Noto emoji 字体
       // ... 其他 fallback 字体
   }
   ```

3. **注意**：即使使用 Noto 字体，`gg` 库仍无法同时渲染中文和彩色 emoji

### 方案 3：移除节点名称中的国旗 emoji（最简单）

如果您不需要国旗显示，可以在节点名称中移除国旗 emoji，只保留文字：
```
原节点名称: 🇭🇰香港1
修改后:     香港1
```

### 方案 4：使用更高级的图像库（复杂）

替换 `gg` 库为支持多字体的库（需要大量代码改动）：
- `freetype-go` + 手动字体管理
- 调用外部工具（如 ImageMagick）
- 使用 HTML to Image 转换

## 推荐设置

**对于大多数用户**，建议：
1. 保持当前设置（使用微软雅黑）
2. 接受国旗显示为字母代码（如 HK、US、JP）
3. 中文和数字完美显示
4. 图片清晰美观

**节点命名建议**：
```
✅ 推荐: 香港节点1, 美国节点2, 日本节点3
✅ 可接受: HK香港1, US美国2, JP日本3  (emoji显示为字母)
⚠️ 不推荐: 🇭🇰1, 🇺🇸2  (图片中只显示字母HK、US)
```

## 技术细节

### 字体加载逻辑
```go
func loadBestFont(dc *gg.Context, size float64) {
    fonts := []string{
        "C:\\Windows\\Fonts\\msyhbd.ttc",   // 微软雅黑粗体（推荐）
        "C:\\Windows\\Fonts\\msyh.ttc",     // 微软雅黑
        "C:\\Windows\\Fonts\\simhei.ttf",   // 黑体
        "C:\\Windows\\Fonts\\simsun.ttc",   // 宋体
    }
    
    for _, font := range fonts {
        if _, err := os.Stat(font); err == nil {
            if err := dc.LoadFontFace(font, size); err == nil {
                return
            }
        }
    }
}
```

### 测试命令

```batch
# 生成测试图片
clash-speedtest.exe --clash-addr "http://127.0.0.1:9097" ^
    --clash-secret "your-secret" ^
    -f png ^
    -i "香港" ^
    -t 1 -r 5 -y

# 查看生成的图片
explorer results\
```

## 总结

✅ **已解决**：
- 中文显示正常
- 数字和字母显示正常
- 表格布局美观

⚠️ **已知限制**：
- 国旗 emoji 显示为字母代码（如 HK、US）
- 这是技术限制，不是 bug

💡 **建议**：
- 在节点名称中使用"香港"、"美国"等文字，而不仅仅是国旗 emoji
- 或者接受图片中国旗显示为字母代码


