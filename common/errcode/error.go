package errcode

import (
	"encoding/json"
	"fmt"
)

type AppError struct {
	code     int    `json:"code"`
	msg      string `json:"msg"`
	cause    error  `json:"cause"`    // 错误链
	occurred string `json:"occurred"` // 错误发生的位置（函数、文件、行号）
}

// 实现 error 接口
func (e *AppError) Error() string {
	if e == nil {
		return ""
	}
	formattedErr := struct {
		Code     int    `json:"code"`
		Msg      string `json:"msg"`
		Cause    string `json:"cause"`
		Occurred string `json:"occurred"`
	}{
		Code:     e.Code(),
		Msg:      e.Msg(),
		Occurred: e.occurred,
	}
	if e.cause != nil {
		formattedErr.Cause = e.cause.Error()
	}
	errByte, _ := json.Marshal(formattedErr)
	return string(errByte)
}

func (e *AppError) String() string {
	return e.Error()
}

func (e *AppError) Code() int {
	return e.code
}

func (e *AppError) Msg() string {
	return e.msg
}

// WithCause 设置错误原因，返回新的AppError实例
func (e *AppError) WithCause(cause error) *AppError {
	return &AppError{
		code:     e.code,
		msg:      e.msg,
		cause:    cause,
		occurred: e.occurred,
	}
}

func newError(code int, msg string) *AppError {
	if code > -1 {
		if _, duplicated := codes[code]; duplicated {
			panic(fmt.Sprintf("预定义错误码 %d 不能重复, 请检查后更换", code))
		}
		codes[code] = struct{}{}
	}

	return &AppError{code: code, msg: msg}
}
