package json

import (
	"encoding/json"
)

// ParseJSONArrayFromString 从 JSON 字符串解析为指定类型的切片
// 如果 jsonStr 为空，返回空切片
func ParseJSONArrayFromString[T any](jsonStr string) ([]T, error) {
	if jsonStr == "" {
		return []T{}, nil
	}

	var result []T
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, err
	}

	return result, nil
}

// ToJSONString 将任意类型转换为 JSON 字符串
// 如果数据为空（切片长度为0），返回 "[]"
func ToJSONString(v any) (string, error) {
	// 检查是否是空切片
	if arr, ok := v.([]interface{}); ok && len(arr) == 0 {
		return "[]", nil
	}

	data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// ParseInt64ArrayFromJSON 从 JSON 字符串解析 int64 数组
// 如果 jsonStr 为空，返回空切片
func ParseInt64ArrayFromJSON(jsonStr string) ([]int64, error) {
	if jsonStr == "" {
		return []int64{}, nil
	}

	var ids []int64
	if err := json.Unmarshal([]byte(jsonStr), &ids); err != nil {
		return nil, err
	}

	return ids, nil
}

// Int64ArrayToJSON 将 int64 数组转换为 JSON 字符串
// 如果数组为空，返回 "[]"
func Int64ArrayToJSON(ids []int64) (string, error) {
	if len(ids) == 0 {
		return "[]", nil
	}

	data, err := json.Marshal(ids)
	if err != nil {
		return "", err
	}

	return string(data), nil
}
