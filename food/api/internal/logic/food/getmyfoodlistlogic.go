package food

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/hd2yao/ecshop/common/app"
	"github.com/hd2yao/ecshop/common/errcode"
	"github.com/hd2yao/ecshop/common/middleware"
	"github.com/hd2yao/ecshop/food/api/internal/svc"
	"github.com/hd2yao/ecshop/food/api/internal/types"
	"github.com/hd2yao/ecshop/food/rpc/types/food"
)

type GetMyFoodListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGetMyFoodListLogic 查询我的美食列表
func NewGetMyFoodListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMyFoodListLogic {
	return &GetMyFoodListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetMyFoodListLogic) GetMyFoodList(req *types.GetMyFoodListRequest) (resp *types.GetMyFoodListResponse, err error) {
	// 1. 从 context 中获取用户 ID（JWT 中间件已验证并存入）
	userID, ok := middleware.GetUserIDFromContext(l.ctx)
	if !ok {
		// 正常情况下不会到这里，因为中间件已经验证过 token
		return &types.GetMyFoodListResponse{
			Code:    errcode.UserTokenInvalid.Code(),
			Message: errcode.UserTokenInvalid.Msg(),
		}, nil
	}

	// 2. 创建分页对象（统一处理分页参数验证和默认值）
	pagination := app.NewPagination(req.Page, req.PageSize)

	l.Infof("查询我的美食列表，用户ID: %d, 页码: %d, 每页: %d", userID, pagination.Page, pagination.PageSize)

	// 3. 调用 RPC 服务获取美食列表
	rpcResp, err := l.svcCtx.FoodRpc.GetMyFoodList(l.ctx, &food.GetMyFoodListReq{
		UserId:   userID,
		Page:     int32(pagination.Page),
		PageSize: int32(pagination.PageSize),
	})

	if err != nil {
		l.Errorf("调用 RPC 获取美食列表失败: %v", err)
		return &types.GetMyFoodListResponse{
			Code:    errcode.CommonServerError.Code(),
			Message: errcode.CommonServerError.Msg(),
		}, nil
	}

	// 4. 设置分页信息
	pagination.SetTotalRows(int(rpcResp.Total))

	// 5. 转换响应
	resp = &types.GetMyFoodListResponse{
		Code:     int(rpcResp.Code),
		Message:  rpcResp.Message,
		Total:    pagination.TotalRows,
		Page:     pagination.Page,
		PageSize: pagination.PageSize,
		FoodList: make([]types.FoodInfo, 0, len(rpcResp.FoodList)),
	}

	// 6. 转换美食列表
	for _, foodInfo := range rpcResp.FoodList {
		// 转换 FoodDetail
		foodDetails := make([]types.FoodStepDetail, 0, len(foodInfo.FoodDetail))
		for _, detail := range foodInfo.FoodDetail {
			foodDetails = append(foodDetails, types.FoodStepDetail{
				Step:        detail.Step,
				StepImgUrl:  detail.StepImgUrl,
				StepContent: detail.StepContent,
			})
		}

		// 转换 FoodList
		foodListItems := make([]types.FoodListItem, 0, len(foodInfo.FoodList))
		for _, item := range foodInfo.FoodList {
			foodListItems = append(foodListItems, types.FoodListItem{
				RecipesTag:   item.RecipesTag,
				RecipesName:  item.RecipesName,
				RecipesSpecs: item.RecipesSpecs,
			})
		}

		resp.FoodList = append(resp.FoodList, types.FoodInfo{
			Id:             foodInfo.Id,
			UserId:         foodInfo.UserId,
			FoodName:       foodInfo.FoodName,
			FoodDes:        foodInfo.FoodDes,
			FoodCateTag:    int(foodInfo.FoodCateTag),
			FoodUrl:        foodInfo.FoodUrl,
			FoodVideoUrl:   foodInfo.FoodVideoUrl,
			FoodTime:       int(foodInfo.FoodTime),
			FoodDifficulty: int(foodInfo.FoodDifficulty),
			FoodDetail:     foodDetails,
			FoodList:       foodListItems,
			FoodStatus:     int(foodInfo.FoodStatus),
			SkuIds:         foodInfo.SkuIds,
			FoodCreatetime: foodInfo.FoodCreatetime,
			FoodUpdatetime: foodInfo.FoodUpdatetime,
		})
	}

	return resp, nil
}
