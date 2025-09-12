package captcha

import (
	"io/ioutil"
	"log"
	"runtime"

	"github.com/golang/freetype/truetype"
)

var Fonts []*truetype.Font

func init() {
	loadSystemFonts()
}

// loadSystemFonts 根据操作系统加载系统字体
func loadSystemFonts() {
	var fontPaths []string

	// 根据操作系统选择字体路径
	switch runtime.GOOS {
	case "darwin": // macOS
		fontPaths = []string{
			"/System/Library/Fonts/STHeiti Light.ttc",
			"/System/Library/Fonts/NewYorkItalic.ttf",
			"/System/Library/Fonts/Apple Braille.ttf",
		}
	case "windows": // Windows
		fontPaths = []string{
			"C:/Windows/Fonts/simhei.ttf", // 黑体
			"C:/Windows/Fonts/msyh.ttc",   // 微软雅黑
			"C:/Windows/Fonts/arial.ttf",  // Arial
		}
	case "linux": // Linux
		fontPaths = []string{
			"/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",
			"/usr/share/fonts/truetype/liberation/LiberationSans-Regular.ttf",
			"/usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc",
		}
	default:
		log.Printf("不支持的操作系统: %s，将使用默认字体", runtime.GOOS)
		return
	}

	// 尝试加载字体
	loadedCount := 0
	for _, path := range fontPaths {
		if font := loadFontFromPath(path); font != nil {
			Fonts = append(Fonts, font)
			loadedCount++
		}
	}

	if loadedCount == 0 {
		log.Println("警告: 未能加载任何系统字体，文字绘制可能出现问题")
		log.Println("建议检查系统字体路径或手动指定字体文件")
	} else {
		log.Printf("成功加载 %d 个系统字体", loadedCount)
	}
}

// loadFontFromPath 从路径加载字体
func loadFontFromPath(path string) *truetype.Font {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Printf("无法读取字体文件: %s (%v)", path, err)
		return nil
	}

	font, err := truetype.Parse(data)
	if err != nil {
		log.Printf("解析字体文件失败: %s (%v)", path, err)
		return nil
	}

	log.Printf("成功加载字体: %s", path)
	return font
}

// AddCustomFont 添加自定义字体
func AddCustomFont(fontData []byte) error {
	font, err := truetype.Parse(fontData)
	if err != nil {
		return err
	}
	Fonts = append(Fonts, font)
	return nil
}

// GetFontCount 获取已加载的字体数量
func GetFontCount() int {
	return len(Fonts)
}

// HasFonts 检查是否有可用字体
func HasFonts() bool {
	return len(Fonts) > 0
}
