package captcha

import (
	"io/ioutil"
	"log"

	"github.com/golang/freetype/truetype"
)

var Fonts []*truetype.Font

func init() {
	// 字体路径可根据系统情况调整
	paths := []string{
		// macOS 中文字体
		"/System/Library/Fonts/STHeiti Light.ttc",
		"/System/Library/Fonts/NewYorkItalic.ttc",
		"/System/Library/Fonts/Apple Braille.ttc",
	}

	for _, path := range paths {
		data, err := ioutil.ReadFile(path)
		if err != nil {
			log.Printf("加载字体失败: %s (%v)", path, err)
			continue
		}
		font, err := truetype.Parse(data)
		if err != nil {
			log.Printf("解析字体失败: %s (%v)", path, err)
			continue
		}
		Fonts = append(Fonts, font)
	}

	if len(Fonts) == 0 {
		log.Println("未成功加载任何字体，drawText 可能无法正常绘制文字")
	} else {
		log.Printf("成功加载 %d 个字体", len(Fonts))
	}
}
