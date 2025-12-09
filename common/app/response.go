package app

import (
	"context"
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/hd2yao/ecshop/common/errcode"
)

// Response 响应结构
type Response struct {
	ctx        context.Context
	w          http.ResponseWriter
	Code       int         `json:"code"`
	Msg        string      `json:"msg"`
	Data       interface{} `json:"data,omitempty"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

// NewResponse 创建响应实例
func NewResponse(ctx context.Context, w http.ResponseWriter) *Response {
	return &Response{
		ctx: ctx,
		w:   w,
	}
}

// SetPagination 设置分页信息
func (r *Response) SetPagination(pagination *Pagination) *Response {
	r.Pagination = pagination
	return r
}

// Success 带数据的成功响应
func (r *Response) Success(data interface{}) {
	r.Code = errcode.Success.Code()
	r.Msg = errcode.Success.Msg()
	r.Data = data
	httpx.WriteJsonCtx(r.ctx, r.w, errcode.Success.HttpStatusCode(), r)
}

// SuccessOk 不带数据的成功响应
func (r *Response) SuccessOk() {
	r.Success("")
}

// Error 带错误信息的响应
func (r *Response) Error(err *errcode.AppError) {
	r.Code = err.Code()
	r.Msg = err.Msg()
	httpx.WriteJsonCtx(r.ctx, r.w, err.HttpStatusCode(), r)
}
