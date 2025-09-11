package captcha

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"log"
	"math"
	"math/rand"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"github.com/mojocn/base64Captcha"
	"golang.org/x/image/font"
)

// ItemChar 验证码绘制器
type ItemChar struct {
	bgColor color.Color
	width   int
	height  int
	nrgba   *image.NRGBA
}

// NewItemChar 创建画布
func NewItemChar(w int, h int, bgColor color.RGBA) *ItemChar {
	d := ItemChar{width: w, height: h}
	m := image.NewNRGBA(image.Rect(0, 0, w, h))
	draw.Draw(m, m.Bounds(), &image.Uniform{bgColor}, image.ZP, draw.Src)
	d.nrgba = m
	return &d
}

// --- 绘制方法 ---
func (item *ItemChar) drawHollowLine() *ItemChar {
	first := item.width / 20
	end := first * 19

	lineColor := base64Captcha.RandLightColor()

	x1 := float64(rand.Intn(first))
	//y1 := float64(rand.Intn(y)+y);

	x2 := float64(rand.Intn(first) + end)

	multiple := float64(rand.Intn(5)+3) / float64(5)
	if int(multiple*10)%3 == 0 {
		multiple = multiple * -1.0
	}

	w := item.height / 20

	for ; x1 < x2; x1++ {

		y := math.Sin(x1*math.Pi*multiple/float64(item.width)) * float64(item.height/3)

		if multiple < 0 {
			y = y + float64(item.height/2)
		}
		item.nrgba.Set(int(x1), int(y), lineColor)

		for i := 0; i <= w; i++ {
			item.nrgba.Set(int(x1), int(y)+i, lineColor)
		}
	}

	return item
}

func random(min int64, max int64) float64 {
	return float64(min) + rand.Float64()*float64(max-min)
}

func (item *ItemChar) drawSineLine() *ItemChar {
	var py float64

	//振幅
	a := rand.Intn(item.height / 2)

	//Y轴方向偏移量
	b := random(int64(-item.height/4), int64(item.height/4))

	//X轴方向偏移量
	f := random(int64(-item.height/4), int64(item.height/4))
	// 周期
	var t float64
	if item.height > item.width/2 {
		t = random(int64(item.width/2), int64(item.height))
	} else if item.height == item.width/2 {
		t = float64(item.height)
	} else {
		t = random(int64(item.height), int64(item.width/2))
	}
	w := float64((2 * math.Pi) / t)

	// 曲线横坐标起始位置
	px1 := 0
	px2 := int(random(int64(float64(item.width)*0.8), int64(item.width)))

	c := base64Captcha.RandDeepColor()

	for px := px1; px < px2; px++ {
		if w != 0 {
			py = float64(a)*math.Sin(w*float64(px)+f) + b + (float64(item.width) / float64(5))
			i := item.height / 5
			for i > 0 {
				item.nrgba.Set(px+i, int(py), c)
				//fmt.Println(px + i,int(py) )
				i--
			}
		}
	}

	return item
}

func (item *ItemChar) drawSlimLine(num int) *ItemChar {
	first := item.width / 10
	end := first * 9

	y := item.height / 3

	for i := 0; i < num; i++ {

		point1 := point{X: rand.Intn(first), Y: rand.Intn(y)}
		point2 := point{X: rand.Intn(first) + end, Y: rand.Intn(y)}

		if i%2 == 0 {
			point1.Y = rand.Intn(y) + y*2
			point2.Y = rand.Intn(y)
		} else {
			point1.Y = rand.Intn(y) + y*(i%2)
			point2.Y = rand.Intn(y) + y*2
		}

		item.drawBeeline(point1, point2, base64Captcha.RandDeepColor())

	}
	return item
}

func (item *ItemChar) drawBeeline(point1 point, point2 point, lineColor color.RGBA) {
	dx := math.Abs(float64(point1.X - point2.X))

	dy := math.Abs(float64(point2.Y - point1.Y))
	sx, sy := 1, 1
	if point1.X >= point2.X {
		sx = -1
	}
	if point1.Y >= point2.Y {
		sy = -1
	}
	err := dx - dy
	for {
		item.nrgba.Set(point1.X, point1.Y, lineColor)
		item.nrgba.Set(point1.X+1, point1.Y, lineColor)
		item.nrgba.Set(point1.X-1, point1.Y, lineColor)
		item.nrgba.Set(point1.X+2, point1.Y, lineColor)
		item.nrgba.Set(point1.X-2, point1.Y, lineColor)
		if point1.X == point2.X && point1.Y == point2.Y {
			return
		}
		e2 := err * 2
		if e2 > -dy {
			err -= dy
			point1.X += sx
		}
		if e2 < dx {
			err += dx
			point1.Y += sy
		}
	}
}

func (item *ItemChar) drawNoise(noiseText string, fonts []*truetype.Font) error {
	if len(fonts) == 0 {
		return errors.New("no fonts available for noise")
	}

	// 噪点字符
	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetClip(item.nrgba.Bounds())
	c.SetDst(item.nrgba)
	c.SetHinting(font.HintingFull)
	rawFontSize := float64(item.height) / (1 + float64(rand.Intn(7))/float64(10))

	for _, char := range noiseText {
		rw := rand.Intn(item.width)
		rh := rand.Intn(item.height)
		fontSize := rawFontSize/2 + float64(rand.Intn(5))
		c.SetSrc(image.NewUniform(color.RGBA{uint8(rand.Intn(200)), uint8(rand.Intn(200)), uint8(rand.Intn(200)), 255}))
		c.SetFontSize(fontSize)
		
		// 随机选择字体
		fontIndex := rand.Intn(len(fonts))
		c.SetFont(fonts[fontIndex])
		
		pt := freetype.Pt(rw, rh)
		if _, err := c.DrawString(string(char), pt); err != nil {
			log.Printf("绘制噪点字符失败: %v", err)
		}
	}
	return nil
}

func (item *ItemChar) drawText(text string, fonts []*truetype.Font) error {
	if len(text) == 0 {
		return errors.New("text must not be empty")
	}
	if len(fonts) == 0 {
		return errors.New("no fonts available")
	}

	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetClip(item.nrgba.Bounds())
	c.SetDst(item.nrgba)
	c.SetHinting(font.HintingFull)

	fontWidth := item.width / len(text)
	for i, s := range text {
		fontSize := item.height * (rand.Intn(7) + 7) / 16
		c.SetSrc(image.NewUniform(color.RGBA{uint8(rand.Intn(150)), uint8(rand.Intn(150)), uint8(rand.Intn(150)), 255}))
		c.SetFontSize(float64(fontSize))
		
		// 随机选择字体
		fontIndex := rand.Intn(len(fonts))
		c.SetFont(fonts[fontIndex])
		
		x := fontWidth*i + fontWidth/fontSize
		y := item.height/2 + fontSize/2 - rand.Intn(item.height/16*3)
		pt := freetype.Pt(x, y)
		if _, err := c.DrawString(string(s), pt); err != nil {
			return err
		}
	}
	return nil
}

func (item *ItemChar) BinaryEncoding() []byte {
	var buf bytes.Buffer
	if err := png.Encode(&buf, item.nrgba); err != nil {
		panic(err.Error())
	}
	return buf.Bytes()
}
func (item *ItemChar) EncodeB64string() string {
	return fmt.Sprintf("data:%s;base64,%s", "image/png", base64.StdEncoding.EncodeToString(item.BinaryEncoding()))
}
func (item *ItemChar) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write(item.BinaryEncoding())
	return int64(n), err
}

type point struct {
	X int
	Y int
}
