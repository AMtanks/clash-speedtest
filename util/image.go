package util

import (
	"fmt"
	"image/color"
	"math"
	"os"

	"github.com/fogleman/gg"
)

// TableImageConfig 表格图片配置
type TableImageConfig struct {
	Title       string
	Headers     []string
	Rows        [][]string
	Colors      [][]color.Color // 每个单元格的颜色
	Width       int
	HeaderColor color.Color
	RowColor1   color.Color // 奇数行颜色
	RowColor2   color.Color // 偶数行颜色
	TextColor   color.Color
	BorderColor color.Color
	FontSize    float64
}

// DefaultTableImageConfig 返回默认配置
func DefaultTableImageConfig() *TableImageConfig {
	return &TableImageConfig{
		Width:       2560,                            // 2K 分辨率
		HeaderColor: color.RGBA{52, 73, 94, 255},     // 深蓝灰
		RowColor1:   color.RGBA{236, 240, 241, 255},  // 浅灰
		RowColor2:   color.RGBA{255, 255, 255, 255},  // 白色
		TextColor:   color.RGBA{44, 62, 80, 255},     // 深灰
		BorderColor: color.RGBA{189, 195, 199, 255},  // 边框灰
		FontSize:    18,                              // 增加字体大小以适应更大的分辨率
	}
}

// loadBestFont 尝试加载最佳字体（支持中文和emoji）
func loadBestFont(dc *gg.Context, size float64) {
	// 字体优先级列表（支持中文和emoji）
	fonts := []string{
		"C:\\Windows\\Fonts\\msyh.ttc",     // 微软雅黑（中文+emoji单色）
	}
	
	for _, font := range fonts {
		if _, err := os.Stat(font); err == nil {
			if err := dc.LoadFontFace(font, size); err == nil {
				return
			}
		}
	}
	
	// 如果都失败了，使用默认字体（可能不支持中文）
	dc.LoadFontFace("C:\\Windows\\Fonts\\arial.ttf", size)
}

// RenderTableImage 渲染表格为图片
func RenderTableImage(config *TableImageConfig) (*gg.Context, error) {
	if config == nil {
		config = DefaultTableImageConfig()
	}

	// 计算尺寸
	numCols := len(config.Headers)
	numRows := len(config.Rows)
	
	if numCols == 0 {
		return nil, fmt.Errorf("no headers provided")
	}

	// 计算列宽
	colWidth := float64(config.Width) / float64(numCols)
	rowHeight := config.FontSize * 2.5
	padding := config.FontSize * 0.5

	// 标题高度
	titleHeight := 0.0
	if config.Title != "" {
		titleHeight = config.FontSize * 3
	}

	// 计算总高度
	totalHeight := titleHeight + rowHeight + float64(numRows)*rowHeight + padding*2

	// 创建画布
	dc := gg.NewContext(config.Width, int(totalHeight))

	// 背景
	dc.SetColor(color.White)
	dc.Clear()

	currentY := padding

	// 绘制标题
	if config.Title != "" {
		dc.SetColor(config.TextColor)
		// 尝试加载支持中文和emoji的字体
		loadBestFont(dc, config.FontSize*1.5)
		dc.DrawStringAnchored(config.Title, float64(config.Width)/2, currentY+config.FontSize*1.5, 0.5, 0.5)
		currentY += titleHeight
	}

	// 加载最佳字体 - 同时支持中文和 emoji 显示
	loadBestFont(dc, config.FontSize)

	// 绘制表头
	dc.SetColor(config.HeaderColor)
	dc.DrawRectangle(0, currentY, float64(config.Width), rowHeight)
	dc.Fill()

	dc.SetColor(color.White)
	for i, header := range config.Headers {
		x := float64(i)*colWidth + colWidth/2
		y := currentY + rowHeight/2
		dc.DrawStringAnchored(header, x, y, 0.5, 0.5)
	}

	// 绘制边框
	dc.SetColor(config.BorderColor)
	dc.SetLineWidth(1)
	dc.DrawRectangle(0, currentY, float64(config.Width), rowHeight)
	dc.Stroke()

	currentY += rowHeight

	// 绘制数据行
	for rowIdx, row := range config.Rows {
		// 背景色（交替）
		if rowIdx%2 == 0 {
			dc.SetColor(config.RowColor1)
		} else {
			dc.SetColor(config.RowColor2)
		}
		dc.DrawRectangle(0, currentY, float64(config.Width), rowHeight)
		dc.Fill()

		// 绘制单元格
		for colIdx, cell := range row {
			x := float64(colIdx)*colWidth + colWidth/2
			y := currentY + rowHeight/2

			// 使用自定义颜色（如果有）
			if config.Colors != nil && rowIdx < len(config.Colors) && colIdx < len(config.Colors[rowIdx]) {
				dc.SetColor(config.Colors[rowIdx][colIdx])
			} else {
				dc.SetColor(config.TextColor)
			}

			dc.DrawStringAnchored(cell, x, y, 0.5, 0.5)
		}

		// 绘制行边框
		dc.SetColor(config.BorderColor)
		dc.SetLineWidth(0.5)
		dc.DrawRectangle(0, currentY, float64(config.Width), rowHeight)
		dc.Stroke()

		currentY += rowHeight
	}

	// 绘制列分隔线
	dc.SetColor(config.BorderColor)
	dc.SetLineWidth(0.5)
	for i := 1; i < numCols; i++ {
		x := float64(i) * colWidth
		dc.DrawLine(x, padding+titleHeight, x, currentY)
		dc.Stroke()
	}

	return dc, nil
}

// GetColorForValue 根据数值返回颜色
func GetColorForValue(value float64, minGood, maxBad float64) color.Color {
	if value < 0 || math.IsNaN(value) {
		return color.RGBA{149, 165, 166, 255} // 灰色
	}
	if value <= minGood {
		return color.RGBA{46, 204, 113, 255} // 绿色
	} else if value >= maxBad {
		return color.RGBA{231, 76, 60, 255} // 红色
	} else {
		return color.RGBA{241, 196, 15, 255} // 黄色
	}
}


