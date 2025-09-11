package api

import (
	"image/color"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hd2yao/ecshop/common/utils/captcha"
)

// CaptchaHandler 验证码处理器
type CaptchaHandler struct{}

// NewCaptchaHandler 创建验证码处理器
func NewCaptchaHandler() *CaptchaHandler {
	return &CaptchaHandler{}
}

// GenerateCaptcha 生成验证码
// @Summary 生成验证码
// @Description 根据参数生成不同类型的验证码
// @Tags 验证码
// @Accept json
// @Produce json
// @Param request body captcha.CaptchaRequest true "验证码请求参数"
// @Success 200 {object} CaptchaResponse "生成成功"
// @Failure 400 {object} ErrorResponse "参数错误"
// @Failure 500 {object} ErrorResponse "生成失败"
// @Router /api/captcha/generate [post]
func (h *CaptchaHandler) GenerateCaptcha(c *gin.Context) {
	var req captcha.CaptchaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    400,
			Message: "参数格式错误",
			Error:   err.Error(),
		})
		return
	}

	// 设置默认值
	if req.Config.Width == 0 || req.Config.Height == 0 {
		req.Config = captcha.GetDefaultConfig()
	}
	if req.CaptchaType == "" {
		req.CaptchaType = "string"
	}

	// 生成验证码
	id, b64s, answer, err := captcha.Generate(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    500,
			Message: "验证码生成失败",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, CaptchaResponse{
		Code:      200,
		Message:   "生成成功",
		CaptchaId: id,
		ImageData: b64s,
		Answer:    answer, // 实际应用中不应该返回答案
	})
}

// VerifyCaptcha 验证验证码
// @Summary 验证验证码
// @Description 验证用户输入的验证码是否正确
// @Tags 验证码
// @Accept json
// @Produce json
// @Param request body VerifyRequest true "验证请求参数"
// @Success 200 {object} VerifyResponse "验证结果"
// @Failure 400 {object} ErrorResponse "参数错误"
// @Router /api/captcha/verify [post]
func (h *CaptchaHandler) VerifyCaptcha(c *gin.Context) {
	var req VerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    400,
			Message: "参数格式错误",
			Error:   err.Error(),
		})
		return
	}

	if req.CaptchaId == "" || req.Answer == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    400,
			Message: "验证码ID和答案不能为空",
		})
		return
	}

	// 验证验证码
	isValid := captcha.Verify(req.CaptchaId, req.Answer)

	c.JSON(http.StatusOK, VerifyResponse{
		Code:    200,
		Message: map[bool]string{true: "验证通过", false: "验证失败"}[isValid],
		Valid:   isValid,
	})
}

// GetCaptchaTypes 获取支持的验证码类型
// @Summary 获取验证码类型列表
// @Description 获取系统支持的所有验证码类型和配置选项
// @Tags 验证码
// @Produce json
// @Success 200 {object} TypesResponse "类型列表"
// @Router /api/captcha/types [get]
func (h *CaptchaHandler) GetCaptchaTypes(c *gin.Context) {
	types := []CaptchaTypeInfo{
		{
			Type:        "string",
			Name:        "字符串验证码",
			Description: "随机字符串，支持数字和字母",
			Examples:    []string{"A2B9X", "K8N4M", "Z7P3Q"},
		},
		{
			Type:        "digit",
			Name:        "数字验证码",
			Description: "纯数字验证码",
			Examples:    []string{"12345", "98762", "54321"},
		},
		{
			Type:        "math",
			Name:        "数学验证码",
			Description: "简单的数学计算题",
			Examples:    []string{"1+2=?", "5-3=?", "2*4=?"},
		},
		{
			Type:        "chinese",
			Name:        "中文验证码",
			Description: "中文字符验证码",
			Examples:    []string{"验证码", "测试用", "汉字符"},
		},
		{
			Type:        "audio",
			Name:        "音频验证码",
			Description: "语音播报数字验证码",
			Examples:    []string{"音频文件"},
		},
	}

	defaultConfig := captcha.GetDefaultConfig()
	defaultDraw := captcha.GetDefaultDrawOptions()

	c.JSON(http.StatusOK, TypesResponse{
		Code:          200,
		Message:       "获取成功",
		Types:         types,
		DefaultConfig: defaultConfig,
		DefaultDraw:   defaultDraw,
	})
}

// GetPresetConfigs 获取预设配置
// @Summary 获取预设验证码配置
// @Description 获取一些常用的验证码配置模板
// @Tags 验证码
// @Produce json
// @Success 200 {object} PresetsResponse "预设配置"
// @Router /api/captcha/presets [get]
func (h *CaptchaHandler) GetPresetConfigs(c *gin.Context) {
	presets := []PresetConfig{
		{
			Name:        "简单模式",
			Description: "基础验证码，适用于一般场景",
			Request: captcha.CaptchaRequest{
				CaptchaType: "string",
				Config: captcha.DriverConfig{
					Width:   200,
					Height:  60,
					Length:  4,
					Source:  "1234567890ABCDEFGHJKLMNPQRSTUVWXYZ",
					BgColor: color.RGBA{255, 255, 255, 255},
				},
				DrawOpts: captcha.DrawOptions{
					UseCustomDraw: false,
					DrawText:      true,
				},
			},
		},
		{
			Name:        "安全模式",
			Description: "增加干扰线和噪点，安全性更高",
			Request: captcha.CaptchaRequest{
				CaptchaType: "string",
				Config: captcha.DriverConfig{
					Width:   240,
					Height:  80,
					Length:  5,
					Source:  "23456789ABCDEFGHJKLMNPQRSTUVWXYZ",
					BgColor: color.RGBA{240, 240, 240, 255},
				},
				DrawOpts: captcha.DrawOptions{
					UseCustomDraw: true,
					DrawText:      true,
					DrawHollow:    true,
					DrawSine:      true,
					DrawSlimLine:  2,
					DrawNoiseText: "1234567890",
				},
			},
		},
		{
			Name:        "数学模式",
			Description: "数学运算验证码，用户友好",
			Request: captcha.CaptchaRequest{
				CaptchaType: "math",
				Config: captcha.DriverConfig{
					Width:   180,
					Height:  60,
					Length:  1,
					BgColor: color.RGBA{255, 255, 255, 255},
				},
				DrawOpts: captcha.DrawOptions{
					UseCustomDraw: true,
					DrawText:      true,
					DrawSlimLine:  1,
				},
			},
		},
		{
			Name:        "中文模式",
			Description: "中文字符验证码",
			Request: captcha.CaptchaRequest{
				CaptchaType: "chinese",
				Config: captcha.DriverConfig{
					Width:   200,
					Height:  80,
					Length:  4,
					Source:  "验证码测试中文字符池",
					BgColor: color.RGBA{248, 248, 248, 255},
				},
				DrawOpts: captcha.DrawOptions{
					UseCustomDraw: true,
					DrawText:      true,
					DrawHollow:    false,
					DrawSine:      true,
				},
			},
		},
	}

	c.JSON(http.StatusOK, PresetsResponse{
		Code:    200,
		Message: "获取成功",
		Presets: presets,
	})
}

// 响应结构体定义

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

type CaptchaResponse struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	CaptchaId string `json:"captcha_id"`
	ImageData string `json:"image_data"`
	Answer    string `json:"answer,omitempty"` // 仅用于测试，生产环境不返回
}

type VerifyRequest struct {
	CaptchaId string `json:"captcha_id" binding:"required"`
	Answer    string `json:"answer" binding:"required"`
}

type VerifyResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Valid   bool   `json:"valid"`
}

type CaptchaTypeInfo struct {
	Type        string   `json:"type"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Examples    []string `json:"examples"`
}

type TypesResponse struct {
	Code          int                  `json:"code"`
	Message       string               `json:"message"`
	Types         []CaptchaTypeInfo    `json:"types"`
	DefaultConfig captcha.DriverConfig `json:"default_config"`
	DefaultDraw   captcha.DrawOptions  `json:"default_draw"`
}

type PresetConfig struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Request     captcha.CaptchaRequest `json:"request"`
}

type PresetsResponse struct {
	Code    int            `json:"code"`
	Message string         `json:"message"`
	Presets []PresetConfig `json:"presets"`
}
