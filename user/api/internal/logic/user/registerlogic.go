package user

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/hd2yao/ecshop/user/api/internal/svc"
	"github.com/hd2yao/ecshop/user/api/internal/types"
	"github.com/hd2yao/ecshop/user/rpc/types/user"
)

type RegisterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 用户注册
func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RegisterLogic) Register(req *types.RegisterRequest) (resp *types.RegisterResponse, err error) {
	// 1. 调用RPC服务处理用户注册
	rpcResp, err := l.svcCtx.UserRpc.Register(l.ctx, &user.RegisterReq{
		Email:     req.Email,
		EmailCode: req.EmailCode,
		Password:  req.Password,
		Name:      req.Name,
		Phone:     req.Phone,
		Sex:       int32(req.Sex),
	})
	if err != nil {
		l.Errorf("RPC调用失败: %v", err)
		return &types.RegisterResponse{
			Code:    500,
			Message: "系统错误",
		}, nil
	}

	// 2. 转换响应数据
	resp = &types.RegisterResponse{
		Code:    int(rpcResp.Code),
		Message: rpcResp.Message,
		UserId:  rpcResp.UserId,
		Token:   rpcResp.Token,
	}

	// 3. 转换用户信息
	if rpcResp.UserInfo != nil {
		resp.UserInfo = types.UserInfo{
			Id:         rpcResp.UserInfo.Id,
			Name:       rpcResp.UserInfo.Name,
			Avatar:     rpcResp.UserInfo.Avatar,
			Email:      rpcResp.UserInfo.Email,
			Phone:      rpcResp.UserInfo.Phone,
			Sex:        int(rpcResp.UserInfo.Sex),
			Points:     int(rpcResp.UserInfo.Points),
			CreateTime: rpcResp.UserInfo.CreateTime,
		}
	}

	return resp, nil
}
