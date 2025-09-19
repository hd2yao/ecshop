package middleware

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/zeromicro/go-zero/rest"
)

// GinStyleLoggerConfig 定义日志配置
type GinStyleLoggerConfig struct {
	// SkipPaths 跳过日志记录的路径
	SkipPaths []string
	// TimeFormat 时间格式
	TimeFormat string
	// EnableColors 是否启用颜色输出
	EnableColors bool
}

// DefaultGinStyleLoggerConfig 默认配置
func DefaultGinStyleLoggerConfig() GinStyleLoggerConfig {
	return GinStyleLoggerConfig{
		SkipPaths:    []string{},
		TimeFormat:   "2006/01/02 - 15:04:05",
		EnableColors: true,
	}
}

// StatusCodeColor 返回状态码对应的颜色
func StatusCodeColor(code int, enableColors bool) string {
	if !enableColors {
		return ""
	}
	
	switch {
	case code >= http.StatusOK && code < http.StatusMultipleChoices:
		return "\033[97;42m" // white fg, green bg
	case code >= http.StatusMultipleChoices && code < http.StatusBadRequest:
		return "\033[90;47m" // black fg, white bg
	case code >= http.StatusBadRequest && code < http.StatusInternalServerError:
		return "\033[90;43m" // black fg, yellow bg
	default:
		return "\033[97;41m" // white fg, red bg
	}
}

// MethodColor 返回HTTP方法对应的颜色
func MethodColor(method string, enableColors bool) string {
	if !enableColors {
		return ""
	}
	
	switch method {
	case http.MethodGet:
		return "\033[97;44m" // white fg, blue bg
	case http.MethodPost:
		return "\033[97;42m" // white fg, green bg
	case http.MethodPut:
		return "\033[97;43m" // white fg, yellow bg
	case http.MethodDelete:
		return "\033[97;41m" // white fg, red bg
	case http.MethodPatch:
		return "\033[97;42m" // white fg, green bg
	case http.MethodHead:
		return "\033[97;45m" // white fg, magenta bg
	case http.MethodOptions:
		return "\033[90;47m" // black fg, white bg
	default:
		return "\033[0m" // reset
	}
}

// ResetColor 重置颜色
func ResetColor(enableColors bool) string {
	if !enableColors {
		return ""
	}
	return "\033[0m"
}

// GinStyleLogger 创建Gin风格的日志中间件
func GinStyleLogger(config ...GinStyleLoggerConfig) rest.Middleware {
	conf := DefaultGinStyleLoggerConfig()
	if len(config) > 0 {
		conf = config[0]
	}

	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// 检查是否需要跳过日志记录
			path := r.URL.Path
			for _, skipPath := range conf.SkipPaths {
				if path == skipPath {
					next(w, r)
					return
				}
			}

			start := time.Now()
			
			// 使用自定义的ResponseWriter来捕获状态码和响应大小
			recorder := &ResponseRecorder{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
				size:          0,
			}

			// 执行下一个处理器
			next(recorder, r)

			// 计算处理时间
			latency := time.Since(start)
			
			// 获取客户端IP
			clientIP := getClientIP(r)
			
			// 构建日志消息
			statusCode := recorder.statusCode
			bodySize := recorder.size
			
			// 颜色设置
			statusColor := StatusCodeColor(statusCode, conf.EnableColors)
			methodColor := MethodColor(r.Method, conf.EnableColors)
			resetColor := ResetColor(conf.EnableColors)
			
			// 格式化日志消息 (类似Gin的格式)
			logMessage := fmt.Sprintf("%s[GIN]%s %s | %s%3d%s | %13v | %15s | %s%-7s%s %s",
				"\033[90m", resetColor, // [GIN] 灰色
				start.Format(conf.TimeFormat),
				statusColor, statusCode, resetColor,
				latency,
				clientIP,
				methodColor, r.Method, resetColor,
				path,
			)
			
			// 如果有查询参数，添加到日志中
			if r.URL.RawQuery != "" {
				logMessage += "?" + r.URL.RawQuery
			}
			
			// 添加响应大小信息
			if bodySize > 0 {
				logMessage += fmt.Sprintf(" (%d bytes)", bodySize)
			}

			// 直接输出到标准输出，避免被go-zero的JSON格式包装
			fmt.Fprintln(os.Stdout, logMessage)
		}
	}
}

// ResponseRecorder 用于记录响应信息
type ResponseRecorder struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func (rec *ResponseRecorder) WriteHeader(code int) {
	rec.statusCode = code
	rec.ResponseWriter.WriteHeader(code)
}

func (rec *ResponseRecorder) Write(b []byte) (int, error) {
	size, err := rec.ResponseWriter.Write(b)
	rec.size += size
	return size, err
}

// getClientIP 获取客户端真实IP
func getClientIP(r *http.Request) string {
	// 检查 X-Forwarded-For 头
	if xForwardedFor := r.Header.Get("X-Forwarded-For"); xForwardedFor != "" {
		// X-Forwarded-For 可能包含多个IP，取第一个
		if idx := len(xForwardedFor); idx > 0 {
			if commaIdx := 0; commaIdx < idx {
				for i, char := range xForwardedFor {
					if char == ',' {
						commaIdx = i
						break
					}
				}
				if commaIdx > 0 {
					return xForwardedFor[:commaIdx]
				}
			}
			return xForwardedFor
		}
	}
	
	// 检查 X-Real-IP 头
	if xRealIP := r.Header.Get("X-Real-IP"); xRealIP != "" {
		return xRealIP
	}
	
	// 检查 X-Forwarded-Host 头
	if xForwardedHost := r.Header.Get("X-Forwarded-Host"); xForwardedHost != "" {
		return xForwardedHost
	}
	
	// 最后使用 RemoteAddr
	return r.RemoteAddr
}
