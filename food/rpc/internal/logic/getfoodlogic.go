package logic

import (
	"context"
	"errors"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/hd2yao/ecshop/common/errcode"
	"github.com/hd2yao/ecshop/food/model"
	"github.com/hd2yao/ecshop/food/rpc/internal/svc"
	"github.com/hd2yao/ecshop/food/rpc/types/food"
)

type GetFoodLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetFoodLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFoodLogic {
	return &GetFoodLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetFood 查询单个美食基础信息
func (l *GetFoodLogic) GetFood(in *food.GetFoodReq) (*food.GetFoodResp, error) {
	// 1. 参数验证
	if in.FoodId <= 0 {
		return &food.GetFoodResp{
			Code:    int32(errcode.CommonParamError.Code()),
			Message: "美食ID无效",
		}, nil
	}

	l.Infof("查询美食信息，美食ID: %d", in.FoodId)

	// 2. 查询美食信息
	foodData, err := l.svcCtx.FoodModel.FindOne(l.ctx, in.FoodId)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			l.Infof("美食不存在，美食ID: %d", in.FoodId)
			return &food.GetFoodResp{
				Code:    int32(errcode.CommonParamError.Code()),
				Message: "美食不存在",
			}, nil
		}
		// 数据库错误
		l.Errorf("查询美食信息失败: %v", err)
		return &food.GetFoodResp{
			Code:    int32(errcode.CommonServerError.Code()),
			Message: errcode.CommonServerError.Msg(),
		}, nil
	}

	// 3. 检查美食状态
	if foodData.FoodStatus != 0 {
		l.Infof("美食已删除，美食ID: %d", in.FoodId)
		return &food.GetFoodResp{
			Code:    int32(errcode.CommonParamError.Code()),
			Message: "美食不存在或已删除",
		}, nil
	}

	// 4. 解析 JSON 字段为结构化数据
	foodDetails, err := model.ParseFoodDetailFromJSON(foodData.FoodDetail)
	if err != nil {
		l.Errorf("解析 food_detail JSON 失败: %v", err)
		return &food.GetFoodResp{
			Code:    int32(errcode.CommonServerError.Code()),
			Message: "解析美食详情失败",
		}, nil
	}

	foodListItems, err := model.ParseFoodListFromJSON(foodData.FoodList)
	if err != nil {
		l.Errorf("解析 food_list JSON 失败: %v", err)
		return &food.GetFoodResp{
			Code:    int32(errcode.CommonServerError.Code()),
			Message: "解析美食清单失败",
		}, nil
	}

	var skuIds []int64
	if foodData.FoodSkuIds.Valid && foodData.FoodSkuIds.String != "" {
		skuIds, err = model.ParseSkuIdsFromJSON(foodData.FoodSkuIds.String)
		if err != nil {
			l.Errorf("解析 skuIds JSON 失败: %v", err)
			// 不阻断流程，继续处理
		}
	}

	// 5. 转换为 proto 结构
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

	// 6. 返回美食信息
	return &food.GetFoodResp{
		Code:    int32(errcode.Success.Code()),
		Message: errcode.Success.Msg(),
		FoodInfo: &food.FoodInfo{
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
		},
	}, nil
}
