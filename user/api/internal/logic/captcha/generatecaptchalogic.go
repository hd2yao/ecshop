package captcha

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/hd2yao/ecshop/user/api/internal/svc"
	"github.com/hd2yao/ecshop/user/api/internal/types"
	"github.com/hd2yao/ecshop/user/rpc/types/user"
)

type GenerateCaptchaLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGenerateCaptchaLogic 生成验证码
func NewGenerateCaptchaLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GenerateCaptchaLogic {
	return &GenerateCaptchaLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GenerateCaptchaLogic) GenerateCaptcha(req *types.CaptchaRequest) (resp *types.CaptchaResponse, err error) {
	// 调用user-rpc服务生成验证码
	rpcReq := &user.GenerateCaptchaReq{
		CaptchaType: req.CaptchaType,
		Config: &user.CaptchaConfig{
			Width:   int32(req.Config.Width),
			Height:  int32(req.Config.Height),
			Length:  int32(req.Config.Length),
			BgColor: req.Config.BgColor,
		},
		DrawOpts: &user.DrawOptions{
			UseCustomDraw: req.DrawOpts.UseCustomDraw,
			DrawText:      req.DrawOpts.DrawText,
			DrawHollow:    req.DrawOpts.DrawHollow,
			DrawSine:      req.DrawOpts.DrawSine,
			DrawSlimLine:  int32(req.DrawOpts.DrawSlimLine),
			DrawNoiseText: req.DrawOpts.DrawNoiseText,
		},
	}

	rpcResp, err := l.svcCtx.UserRpc.GenerateCaptcha(l.ctx, rpcReq)
	if err != nil {
		l.Errorf("调用RPC生成验证码失败: %v", err)
		return &types.CaptchaResponse{
			Code:    500,
			Message: "验证码生成失败",
		}, nil
	}

	return &types.CaptchaResponse{
		Code:      int(rpcResp.Code),
		Message:   rpcResp.Message,
		CaptchaId: rpcResp.CaptchaId,
		ImageData: rpcResp.ImageData,
		Answer:    rpcResp.Answer,
	}, nil
}

