package user

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/hd2yao/ecshop/common/errcode"
	"github.com/hd2yao/ecshop/user/api/internal/svc"
	"github.com/hd2yao/ecshop/user/api/internal/types"
	"github.com/hd2yao/ecshop/user/rpc/types/user"
)

type RegisterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// Register 用户注册
func (l *RegisterLogic) Register(req *types.RegisterRequest) (resp *types.RegisterResponse, err error) {
	// 1. 调用RPC服务处理用户注册
	rpcResp, err := l.svcCtx.UserRpc.Register(l.ctx, &user.RegisterReq{
		Email:     req.Email,
		EmailCode: req.EmailCode,
		Password:  req.Password,
		Name:      req.Name,
		Phone:     req.Phone,
		Sex:       int32(req.Sex),
		Avatar:    req.Avatar,
	})
	if err != nil {
		l.Errorf("RPC调用失败: %v", err)
		return nil, errcode.CommonServerError
	}

	// 2. 检查RPC响应的错误码
	if rpcResp.Code != int32(errcode.Success.Code()) {
		// 根据RPC返回的错误码找到对应的AppError
		if appErr := l.getAppErrorByCode(int(rpcResp.Code)); appErr != nil {
			return nil, appErr
		}
		// 如果找不到对应的错误码，返回通用服务错误
		return nil, errcode.CommonServerError
	}

	// 3. 构建成功响应
	resp = &types.RegisterResponse{
		Code:    int(rpcResp.Code),
		Message: rpcResp.Message,
		UserId:  rpcResp.UserId,
		Token:   rpcResp.Token,
	}

	// 4. 转换用户信息
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

// getAppErrorByCode 根据错误码获取对应的AppError
func (l *RegisterLogic) getAppErrorByCode(code int) *errcode.AppError {
	switch code {
	case errcode.CommonParamError.Code():
		return errcode.CommonParamError
	case errcode.CommonServerError.Code():
		return errcode.CommonServerError
	case errcode.UserCodeEmailError.Code():
		return errcode.UserCodeEmailError
	case errcode.UserAccountExist.Code():
		return errcode.UserAccountExist
	default:
		return nil
	}
}
