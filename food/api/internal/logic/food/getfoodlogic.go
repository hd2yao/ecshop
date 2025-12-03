package food

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/hd2yao/ecshop/common/errcode"
	"github.com/hd2yao/ecshop/food/api/internal/svc"
	"github.com/hd2yao/ecshop/food/api/internal/types"
	"github.com/hd2yao/ecshop/food/rpc/types/food"
)

type GetFoodLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGetFoodLogic 查询单个美食基础信息
func NewGetFoodLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFoodLogic {
	return &GetFoodLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetFoodLogic) GetFood(req *types.GetFoodRequest) (resp *types.GetFoodResponse, err error) {
	l.Infof("查询美食信息，美食ID: %d", req.FoodId)

	// 调用 RPC 服务获取美食信息
	rpcResp, err := l.svcCtx.FoodRpc.GetFood(l.ctx, &food.GetFoodReq{
		FoodId: req.FoodId,
	})

	if err != nil {
		l.Errorf("调用 RPC 获取美食信息失败: %v", err)
		return &types.GetFoodResponse{
			Code:    errcode.CommonServerError.Code(),
			Message: errcode.CommonServerError.Msg(),
		}, nil
	}

	// 转换响应
	resp = &types.GetFoodResponse{
		Code:    int(rpcResp.Code),
		Message: rpcResp.Message,
	}

	// 如果成功，填充美食信息
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
