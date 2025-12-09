package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/hd2yao/ecshop/common/errcode"
	"github.com/hd2yao/ecshop/food/model"
	"github.com/hd2yao/ecshop/food/rpc/internal/svc"
	"github.com/hd2yao/ecshop/food/rpc/types/food"
)

type GetMyFoodListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetMyFoodListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMyFoodListLogic {
	return &GetMyFoodListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetMyFoodList 查询我的美食列表
func (l *GetMyFoodListLogic) GetMyFoodList(in *food.GetMyFoodListReq) (*food.GetMyFoodListResp, error) {
	// 1. 参数验证
	if in.UserId <= 0 {
		return &food.GetMyFoodListResp{
			Code:    int32(errcode.CommonParamError.Code()),
			Message: "用户ID无效",
		}, nil
	}

	page := in.Page
	pageSize := in.PageSize

	l.Infof("查询我的美食列表，用户ID: %d, 页码: %d, 每页: %d", in.UserId, page, pageSize)

	// 2. 查询美食列表（优先使用分页缓存，按需构建）
	foodDTOs, total, err := l.svcCtx.FoodModel.CacheService().GetMyFoodPage(l.ctx, in.UserId, page, pageSize)
	if err != nil {
		l.Errorf("查询美食列表失败: %v", err)
		return &food.GetMyFoodListResp{
			Code:    int32(errcode.CommonServerError.Code()),
			Message: errcode.CommonServerError.Msg(),
		}, nil
	}

	// 3. 转换数据格式
	foodList := make([]*food.FoodInfo, 0, len(foodDTOs))
	for _, foodData := range foodDTOs {
		// 解析 JSON 字段为结构化数据
		foodDetails, err := model.ParseFoodDetailFromJSON(foodData.FoodDetail)
		if err != nil {
			l.Errorf("解析 food_detail JSON 失败，美食ID: %d, 错误: %v", foodData.Id, err)
			foodDetails = []model.FoodStepDetailEntity{} // 使用空数组
		}

		foodListItems, err := model.ParseFoodListFromJSON(foodData.FoodList)
		if err != nil {
			l.Errorf("解析 food_list JSON 失败，美食ID: %d, 错误: %v", foodData.Id, err)
			foodListItems = []model.FoodListItemEntity{} // 使用空数组
		}

		var skuIds []int64
		if foodData.FoodSkuIds != "" {
			skuIds, err = model.ParseSkuIdsFromJSON(foodData.FoodSkuIds)
			if err != nil {
				l.Errorf("解析 skuIds JSON 失败，美食ID: %d, 错误: %v", foodData.Id, err)
				skuIds = []int64{} // 使用空数组
			}
		}

		// 转换为 proto 结构
		foodDetailProto := make([]*food.FoodStepDetail, 0, len(foodDetails))
		for _, detail := range foodDetails {
			foodDetailProto = append(foodDetailProto, &food.FoodStepDetail{
				Step:        detail.Step,
				StepImgUrl:  detail.StepImgUrl,
				StepContent: detail.StepContent,
			})
		}

		foodListProto := make([]*food.FoodListItem, 0, len(foodListItems))
		for _, item := range foodListItems {
			foodListProto = append(foodListProto, &food.FoodListItem{
				RecipesTag:   item.RecipesTag,
				RecipesName:  item.RecipesName,
				RecipesSpecs: item.RecipesSpecs,
			})
		}

		createTime := ""
		if !foodData.FoodCreatetime.IsZero() {
			createTime = foodData.FoodCreatetime.Format("2006-01-02 15:04:05")
		}

		updateTime := ""
		if !foodData.FoodUpdatetime.IsZero() {
			updateTime = foodData.FoodUpdatetime.Format("2006-01-02 15:04:05")
		}

		foodList = append(foodList, &food.FoodInfo{
			Id:             foodData.Id,
			UserId:         foodData.UserId,
			FoodName:       foodData.FoodName,
			FoodDes:        foodData.FoodDes,
			FoodCateTag:    int32(foodData.FoodCateTag),
			FoodUrl:        foodData.FoodUrl,
			FoodVideoUrl:   foodData.FoodVideoUrl,
			FoodTime:       int32(foodData.FoodTime),
			FoodDifficulty: int32(foodData.FoodDifficulty),
			FoodDetail:     foodDetailProto,
			FoodList:       foodListProto,
			FoodStatus:     int32(foodData.FoodStatus),
			SkuIds:         skuIds,
			FoodCreatetime: createTime,
			FoodUpdatetime: updateTime,
		})
	}

	// 4. 返回结果
	return &food.GetMyFoodListResp{
		Code:     int32(errcode.Success.Code()),
		Message:  errcode.Success.Msg(),
		Total:    int32(total),
		Page:     page,
		PageSize: pageSize,
		FoodList: foodList,
	}, nil
}
