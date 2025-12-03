package logic

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/hd2yao/ecshop/common/errcode"
	"github.com/hd2yao/ecshop/food/model"
	"github.com/hd2yao/ecshop/food/rpc/internal/svc"
	"github.com/hd2yao/ecshop/food/rpc/types/food"
)

// convertFoodDetailToJSON 将 proto FoodStepDetail 列表转换为 JSON 字符串
func convertFoodDetailToJSON(details []*food.FoodStepDetail) (string, error) {
	if len(details) == 0 {
		return "[]", nil
	}

	entities := make([]model.FoodStepDetailEntity, 0, len(details))
	for _, detail := range details {
		entities = append(entities, model.FoodStepDetailEntity{
			Step:        detail.Step,
			StepImgUrl:  detail.StepImgUrl,
			StepContent: detail.StepContent,
		})
	}

	return model.FoodDetailToJSON(entities)
}

// convertFoodListToJSON 将 proto FoodListItem 列表转换为 JSON 字符串
func convertFoodListToJSON(items []*food.FoodListItem) (string, error) {
	if len(items) == 0 {
		return "[]", nil
	}

	entities := make([]model.FoodListItemEntity, 0, len(items))
	for _, item := range items {
		entities = append(entities, model.FoodListItemEntity{
			RecipesTag:   item.RecipesTag,
			RecipesName:  item.RecipesName,
			RecipesSpecs: item.RecipesSpecs,
		})
	}

	return model.FoodListToJSON(entities)
}

type CreateOrUpdateFoodLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateOrUpdateFoodLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateOrUpdateFoodLogic {
	return &CreateOrUpdateFoodLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// CreateOrUpdateFood 新增/修改美食信息
func (l *CreateOrUpdateFoodLogic) CreateOrUpdateFood(in *food.CreateOrUpdateFoodReq) (*food.CreateOrUpdateFoodResp, error) {
	// 1. 参数验证
	if in.UserId <= 0 {
		return &food.CreateOrUpdateFoodResp{
			Code:    int32(errcode.CommonParamError.Code()),
			Message: "用户ID无效",
		}, nil
	}

	if in.FoodName == "" {
		return &food.CreateOrUpdateFoodResp{
			Code:    int32(errcode.CommonParamError.Code()),
			Message: "美食名称不能为空",
		}, nil
	}

	if in.FoodCateTag <= 0 {
		return &food.CreateOrUpdateFoodResp{
			Code:    int32(errcode.CommonParamError.Code()),
			Message: "美食分类无效",
		}, nil
	}

	now := time.Now()

	// 将 proto 结构化数据转换为 JSON 字符串
	foodDetailJSON, err := convertFoodDetailToJSON(in.FoodDetail)
	if err != nil {
		l.Errorf("转换 food_detail 为 JSON 失败: %v", err)
		return &food.CreateOrUpdateFoodResp{
			Code:    int32(errcode.CommonParamError.Code()),
			Message: "美食详情格式错误",
		}, nil
	}

	foodListJSON, err := convertFoodListToJSON(in.FoodList)
	if err != nil {
		l.Errorf("转换 food_list 为 JSON 失败: %v", err)
		return &food.CreateOrUpdateFoodResp{
			Code:    int32(errcode.CommonParamError.Code()),
			Message: "美食清单格式错误",
		}, nil
	}

	foodSkuIds := sql.NullString{Valid: false}
	if len(in.SkuIds) > 0 {
		skuIdsJSON, err := model.SkuIdsToJSON(in.SkuIds)
		if err != nil {
			l.Errorf("转换 skuIds 为 JSON 失败: %v", err)
			return &food.CreateOrUpdateFoodResp{
				Code:    int32(errcode.CommonParamError.Code()),
				Message: "商品ID列表格式错误",
			}, nil
		}
		foodSkuIds = sql.NullString{String: skuIdsJSON, Valid: true}
	}

	// 2. 判断是新增还是修改
	if in.FoodId > 0 {
		// 修改美食信息
		l.Infof("修改美食信息，美食ID: %d, 用户ID: %d", in.FoodId, in.UserId)

		// 2.1 查询原美食信息
		existingFood, err := l.svcCtx.FoodModel.FindOne(l.ctx, in.FoodId)
		if err != nil {
			if errors.Is(err, model.ErrNotFound) {
				return &food.CreateOrUpdateFoodResp{
					Code:    int32(errcode.CommonParamError.Code()),
					Message: "美食不存在",
				}, nil
			}
			l.Errorf("查询美食信息失败: %v", err)
			return &food.CreateOrUpdateFoodResp{
				Code:    int32(errcode.CommonServerError.Code()),
				Message: errcode.CommonServerError.Msg(),
			}, nil
		}

		// 2.2 验证权限（只能修改自己的美食）
		if existingFood.UserId != in.UserId {
			l.Errorf("用户尝试修改他人的美食，用户ID: %d, 美食ID: %d, 美食所属用户ID: %d", in.UserId, in.FoodId, existingFood.UserId)
			return &food.CreateOrUpdateFoodResp{
				Code:    int32(errcode.CommonParamError.Code()),
				Message: "无权限修改该美食",
			}, nil
		}

		// 2.3 更新美食信息
		existingFood.FoodName = in.FoodName
		existingFood.FoodDes = in.FoodDes
		existingFood.FoodCateTag = int64(in.FoodCateTag)
		existingFood.FoodUrl = in.FoodUrl
		existingFood.FoodVideoUrl = in.FoodVideoUrl
		existingFood.FoodTime = int64(in.FoodTime)
		existingFood.FoodDifficulty = int64(in.FoodDifficulty)
		existingFood.FoodDetail = foodDetailJSON
		existingFood.FoodList = foodListJSON
		existingFood.FoodSkuIds = foodSkuIds
		existingFood.FoodUpdatetime = now

		err = l.svcCtx.FoodModel.Update(l.ctx, existingFood)
		if err != nil {
			l.Errorf("更新美食信息失败: %v", err)
			return &food.CreateOrUpdateFoodResp{
				Code:    int32(errcode.CommonServerError.Code()),
				Message: errcode.CommonServerError.Msg(),
			}, nil
		}

		// 2.4 重新查询更新后的美食信息（确保数据一致性）
		updatedFood, err := l.svcCtx.FoodModel.FindOne(l.ctx, existingFood.Id)
		if err != nil {
			l.Errorf("查询更新后的美食信息失败: %v", err)
			return &food.CreateOrUpdateFoodResp{
				Code:    int32(errcode.CommonServerError.Code()),
				Message: errcode.CommonServerError.Msg(),
			}, nil
		}

		// 2.5 解析 JSON 并转换为 proto 结构
		foodDetails, err := model.ParseFoodDetailFromJSON(updatedFood.FoodDetail)
		if err != nil {
			l.Errorf("解析 food_detail JSON 失败: %v", err)
			foodDetails = []model.FoodStepDetailEntity{}
		}

		foodListItems, err := model.ParseFoodListFromJSON(updatedFood.FoodList)
		if err != nil {
			l.Errorf("解析 food_list JSON 失败: %v", err)
			foodListItems = []model.FoodListItemEntity{}
		}

		var skuIds []int64
		if updatedFood.FoodSkuIds.Valid && updatedFood.FoodSkuIds.String != "" {
			skuIds, err = model.ParseSkuIdsFromJSON(updatedFood.FoodSkuIds.String)
			if err != nil {
				l.Errorf("解析 skuIds JSON 失败: %v", err)
				skuIds = []int64{}
			}
		}

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
		if !updatedFood.FoodCreatetime.IsZero() {
			createTime = updatedFood.FoodCreatetime.Format("2006-01-02 15:04:05")
		}

		updateTime := ""
		if !updatedFood.FoodUpdatetime.IsZero() {
			updateTime = updatedFood.FoodUpdatetime.Format("2006-01-02 15:04:05")
		}

		return &food.CreateOrUpdateFoodResp{
			Code:    int32(errcode.Success.Code()),
			Message: errcode.Success.Msg(),
			FoodId:  updatedFood.Id,
			FoodInfo: &food.FoodInfo{
				Id:             updatedFood.Id,
				UserId:         updatedFood.UserId,
				FoodName:       updatedFood.FoodName,
				FoodDes:        updatedFood.FoodDes,
				FoodCateTag:    int32(updatedFood.FoodCateTag),
				FoodUrl:        updatedFood.FoodUrl,
				FoodVideoUrl:   updatedFood.FoodVideoUrl,
				FoodTime:       int32(updatedFood.FoodTime),
				FoodDifficulty: int32(updatedFood.FoodDifficulty),
				FoodDetail:     foodDetailProto,
				FoodList:       foodListProto,
				FoodStatus:     int32(updatedFood.FoodStatus),
				SkuIds:         skuIds,
				FoodCreatetime: createTime,
				FoodUpdatetime: updateTime,
			},
		}, nil
	} else {
		// 新增美食信息
		l.Infof("新增美食信息，用户ID: %d", in.UserId)

		// 3.1 创建新美食
		newFood := &model.Food{
			UserId:         in.UserId,
			FoodName:       in.FoodName,
			FoodDes:        in.FoodDes,
			FoodCateTag:    int64(in.FoodCateTag),
			FoodUrl:        in.FoodUrl,
			FoodVideoUrl:   in.FoodVideoUrl,
			FoodTime:       int64(in.FoodTime),
			FoodDifficulty: int64(in.FoodDifficulty),
			FoodDetail:     foodDetailJSON,
			FoodList:       foodListJSON,
			FoodStatus:     0, // 默认有效
			FoodSkuIds:     foodSkuIds,
			FoodCreatetime: now,
			FoodUpdatetime: now,
		}

		result, err := l.svcCtx.FoodModel.Insert(l.ctx, newFood)
		if err != nil {
			l.Errorf("新增美食信息失败: %v", err)
			return &food.CreateOrUpdateFoodResp{
				Code:    int32(errcode.CommonServerError.Code()),
				Message: errcode.CommonServerError.Msg(),
			}, nil
		}

		// 3.2 获取新插入的美食ID
		foodId, err := result.LastInsertId()
		if err != nil {
			l.Errorf("获取新美食ID失败: %v", err)
			return &food.CreateOrUpdateFoodResp{
				Code:    int32(errcode.CommonServerError.Code()),
				Message: errcode.CommonServerError.Msg(),
			}, nil
		}

		// 3.3 查询新创建的美食信息
		createdFood, err := l.svcCtx.FoodModel.FindOne(l.ctx, foodId)
		if err != nil {
			l.Errorf("查询新创建的美食信息失败: %v", err)
			return &food.CreateOrUpdateFoodResp{
				Code:    int32(errcode.CommonServerError.Code()),
				Message: errcode.CommonServerError.Msg(),
			}, nil
		}

		// 3.4 解析 JSON 并转换为 proto 结构
		foodDetails, err := model.ParseFoodDetailFromJSON(createdFood.FoodDetail)
		if err != nil {
			l.Errorf("解析 food_detail JSON 失败: %v", err)
			foodDetails = []model.FoodStepDetailEntity{}
		}

		foodListItems, err := model.ParseFoodListFromJSON(createdFood.FoodList)
		if err != nil {
			l.Errorf("解析 food_list JSON 失败: %v", err)
			foodListItems = []model.FoodListItemEntity{}
		}

		var skuIds []int64
		if createdFood.FoodSkuIds.Valid && createdFood.FoodSkuIds.String != "" {
			skuIds, err = model.ParseSkuIdsFromJSON(createdFood.FoodSkuIds.String)
			if err != nil {
				l.Errorf("解析 skuIds JSON 失败: %v", err)
				skuIds = []int64{}
			}
		}

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
		if !createdFood.FoodCreatetime.IsZero() {
			createTime = createdFood.FoodCreatetime.Format("2006-01-02 15:04:05")
		}

		updateTime := ""
		if !createdFood.FoodUpdatetime.IsZero() {
			updateTime = createdFood.FoodUpdatetime.Format("2006-01-02 15:04:05")
		}

		return &food.CreateOrUpdateFoodResp{
			Code:    int32(errcode.Success.Code()),
			Message: errcode.Success.Msg(),
			FoodId:  foodId,
			FoodInfo: &food.FoodInfo{
				Id:             createdFood.Id,
				UserId:         createdFood.UserId,
				FoodName:       createdFood.FoodName,
				FoodDes:        createdFood.FoodDes,
				FoodCateTag:    int32(createdFood.FoodCateTag),
				FoodUrl:        createdFood.FoodUrl,
				FoodVideoUrl:   createdFood.FoodVideoUrl,
				FoodTime:       int32(createdFood.FoodTime),
				FoodDifficulty: int32(createdFood.FoodDifficulty),
				FoodDetail:     foodDetailProto,
				FoodList:       foodListProto,
				FoodStatus:     int32(createdFood.FoodStatus),
				SkuIds:         skuIds,
				FoodCreatetime: createTime,
				FoodUpdatetime: updateTime,
			},
		}, nil
	}
}
