package app

import (
	"github.com/gin-gonic/gin"

	"github.com/hd2yao/ecshop/common/errcode"
)

type response struct {
	ctx  *gin.Context
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

func NewResponse(ctx *gin.Context) *response {
	return &response{ctx: ctx}
}

// Success 带数据的成功响应
func (r *response) Success(data interface{}) {
	r.Code = errcode.Success.Code()
	r.Msg = errcode.Success.Msg()
	r.Data = data
	r.ctx.JSON(errcode.Success.HttpStatusCode(), r)
}

// SuccessOk 不带数据的成功响应
// 针对只需要知道成功状态的接口响应，简化接口程序的调用
func (r *response) SuccessOk() {
	r.Success("")
}

// Error 带错误信息的响应
func (r *response) Error(err *errcode.AppError) {
	r.Code = err.Code()
	r.Msg = err.Msg()
	r.ctx.JSON(err.HttpStatusCode(), r)
}
