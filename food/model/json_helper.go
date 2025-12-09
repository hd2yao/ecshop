package model

import (
	"github.com/hd2yao/ecshop/common/utils/json"
)

// FoodStepDetailEntity 美食步骤详情实体（用于 JSON 序列化/反序列化）
type FoodStepDetailEntity struct {
	Step        string `json:"step"`        // 步骤名称
	StepImgUrl  string `json:"stepImgUrl"`  // 步骤图片URL（兼容 img）
	StepContent string `json:"stepContent"` // 步骤内容（兼容 content）
	// 兼容字段
	Img     string `json:"img,omitempty"`     // 兼容旧格式
	Content string `json:"content,omitempty"` // 兼容旧格式
}

// FoodListItemEntity 美食清单项实体（用于 JSON 序列化/反序列化）
type FoodListItemEntity struct {
	RecipesTag   string `json:"recipesTag"`   // 食材分类
	RecipesName  string `json:"recipesName"`  // 食材名称
	RecipesSpecs string `json:"recipesSpecs"` // 食材规格
}

// ParseFoodDetailFromJSON 从 JSON 字符串解析美食详情
func ParseFoodDetailFromJSON(jsonStr string) ([]FoodStepDetailEntity, error) {
	details, err := json.ParseJSONArrayFromString[FoodStepDetailEntity](jsonStr)
	if err != nil {
		return nil, err
	}

	// 处理兼容字段：如果 stepImgUrl 为空但 img 有值，则使用 img
	// 如果 stepContent 为空但 content 有值，则使用 content
	for i := range details {
		if details[i].StepImgUrl == "" && details[i].Img != "" {
			details[i].StepImgUrl = details[i].Img
		}
		if details[i].StepContent == "" && details[i].Content != "" {
			details[i].StepContent = details[i].Content
		}
	}

	return details, nil
}

// ParseFoodListFromJSON 从 JSON 字符串解析美食清单
func ParseFoodListFromJSON(jsonStr string) ([]FoodListItemEntity, error) {
	return json.ParseJSONArrayFromString[FoodListItemEntity](jsonStr)
}

// ParseSkuIdsFromJSON 从 JSON 字符串解析商品ID列表
func ParseSkuIdsFromJSON(jsonStr string) ([]int64, error) {
	return json.ParseInt64ArrayFromJSON(jsonStr)
}

// FoodDetailToJSON 将美食详情转换为 JSON 字符串
func FoodDetailToJSON(details []FoodStepDetailEntity) (string, error) {
	return json.ToJSONString(details)
}

// FoodListToJSON 将美食清单转换为 JSON 字符串
func FoodListToJSON(items []FoodListItemEntity) (string, error) {
	return json.ToJSONString(items)
}

// SkuIdsToJSON 将商品ID列表转换为 JSON 字符串
func SkuIdsToJSON(ids []int64) (string, error) {
	return json.Int64ArrayToJSON(ids)
}
