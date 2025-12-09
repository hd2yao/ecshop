package food

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/hd2yao/ecshop/common/errcode"
	"github.com/hd2yao/ecshop/common/middleware"
	"github.com/hd2yao/ecshop/food/api/internal/svc"
	"github.com/hd2yao/ecshop/food/api/internal/types"
	"github.com/hd2yao/ecshop/food/rpc/types/food"
)

type CreateOrUpdateFoodLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewCreateOrUpdateFoodLogic 新增/修改美食信息（food_id为0或未提供时为新增，否则为修改）
func NewCreateOrUpdateFoodLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateOrUpdateFoodLogic {
	return &CreateOrUpdateFoodLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateOrUpdateFoodLogic) CreateOrUpdateFood(req *types.CreateOrUpdateFoodRequest) (resp *types.CreateOrUpdateFoodResponse, err error) {
	// 1. 从 context 中获取用户 ID（JWT 中间件已验证并存入）
	userID, ok := middleware.GetUserIDFromContext(l.ctx)
	if !ok {
		// 正常情况下不会到这里，因为中间件已经验证过 token
		return &types.CreateOrUpdateFoodResponse{
			Code:    errcode.UserTokenInvalid.Code(),
			Message: errcode.UserTokenInvalid.Msg(),
		}, nil
	}

	action := "新增"
	if req.Id > 0 {
		action = "修改"
	}
	l.Infof("%s美食信息，用户ID: %d, 美食ID: %d", action, userID, req.Id)

	// 2. 转换 API types 为 proto 结构
	foodDetailProto := make([]*food.FoodStepDetail, 0, len(req.FoodDetail))
	for _, detail := range req.FoodDetail {
		foodDetailProto = append(foodDetailProto, &food.FoodStepDetail{
			Step:        detail.Step,
			StepImgUrl:  detail.StepImgUrl,
			StepContent: detail.StepContent,
		})
	}

	foodListProto := make([]*food.FoodListItem, 0, len(req.FoodList))
	for _, item := range req.FoodList {
		foodListProto = append(foodListProto, &food.FoodListItem{
			RecipesTag:   item.RecipesTag,
			RecipesName:  item.RecipesName,
			RecipesSpecs: item.RecipesSpecs,
		})
	}

	// 3. 调用 RPC 服务创建或更新美食信息
	rpcResp, err := l.svcCtx.FoodRpc.CreateOrUpdateFood(l.ctx, &food.CreateOrUpdateFoodReq{
		FoodId:         req.Id,
		UserId:         userID,
		FoodName:       req.FoodName,
		FoodDes:        req.FoodDes,
		FoodCateTag:    int32(req.FoodCateTag),
		FoodUrl:        req.FoodUrl,
		FoodVideoUrl:   req.FoodVideoUrl,
		FoodTime:       int32(req.FoodTime),
		FoodDifficulty: int32(req.FoodDifficulty),
		FoodDetail:     foodDetailProto,
		FoodList:       foodListProto,
		SkuIds:         req.SkuIds,
	})

	if err != nil {
		l.Errorf("调用 RPC %s美食信息失败: %v", action, err)
		return &types.CreateOrUpdateFoodResponse{
			Code:    errcode.CommonServerError.Code(),
			Message: errcode.CommonServerError.Msg(),
		}, nil
	}

	// 3. 转换响应
	resp = &types.CreateOrUpdateFoodResponse{
		Code:    int(rpcResp.Code),
		Message: rpcResp.Message,
		FoodId:  rpcResp.FoodId,
	}

	// 4. 如果成功，填充美食信息
	if rpcResp.Code == int32(errcode.Success.Code()) && rpcResp.FoodInfo != nil {
		// 转换 FoodDetail
		foodDetails := make([]types.FoodStepDetail, 0, len(rpcResp.FoodInfo.FoodDetail))
		for _, detail := range rpcResp.FoodInfo.FoodDetail {
			foodDetails = append(foodDetails, types.FoodStepDetail{
				Step:        detail.Step,
				StepImgUrl:  detail.StepImgUrl,
				StepContent: detail.StepContent,
			})
		}

		// 转换 FoodList
		foodListItems := make([]types.FoodListItem, 0, len(rpcResp.FoodInfo.FoodList))
		for _, item := range rpcResp.FoodInfo.FoodList {
			foodListItems = append(foodListItems, types.FoodListItem{
				RecipesTag:   item.RecipesTag,
				RecipesName:  item.RecipesName,
				RecipesSpecs: item.RecipesSpecs,
			})
		}

		resp.FoodInfo = types.FoodInfo{
			Id:             rpcResp.FoodInfo.Id,
			UserId:         rpcResp.FoodInfo.UserId,
			FoodName:       rpcResp.FoodInfo.FoodName,
			FoodDes:        rpcResp.FoodInfo.FoodDes,
			FoodCateTag:    int(rpcResp.FoodInfo.FoodCateTag),
			FoodUrl:        rpcResp.FoodInfo.FoodUrl,
			FoodVideoUrl:   rpcResp.FoodInfo.FoodVideoUrl,
			FoodTime:       int(rpcResp.FoodInfo.FoodTime),
			FoodDifficulty: int(rpcResp.FoodInfo.FoodDifficulty),
			FoodDetail:     foodDetails,
			FoodList:       foodListItems,
			FoodStatus:     int(rpcResp.FoodInfo.FoodStatus),
			SkuIds:         rpcResp.FoodInfo.SkuIds,
			FoodCreatetime: rpcResp.FoodInfo.FoodCreatetime,
			FoodUpdatetime: rpcResp.FoodInfo.FoodUpdatetime,
		}
	}

	return resp, nil
}
