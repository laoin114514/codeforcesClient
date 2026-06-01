package codeforcessdk

import (
	"encoding/json"
	"fmt"
	"sort"
)

type StructTransfer struct {
	v       any
	mapData map[string]any
}

func NewStructTransfer(v any) *StructTransfer {
	data, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	var result map[string]any
	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil
	}
	return &StructTransfer{
		v:       v,
		mapData: result,
	}
}

// 将结构体转换为map
func (s *StructTransfer) ToMap() (map[string]any, error) {
	return s.mapData, nil
}

// 将结构体转换为有序参数字符串
func (s *StructTransfer) ToOrderStr() (string, error) {
	keyArr := []string{}
	for k, _ := range s.mapData {
		keyArr = append(keyArr, k)
	}
	sort.Strings(keyArr)
	str := ""
	for _, key := range keyArr {
		str += fmt.Sprintf("%s=%v&", key, s.mapData[key])
	}
	if len(str) == 0 {
		return "", nil
	}
	str = str[:len(str)-1]
	return str, nil
}

// 添加参数
func (s *StructTransfer) AddParam(key string, value any) *StructTransfer {
	s.mapData[key] = value
	return s
}
func (s *StructTransfer) AddParamWithMap(params map[string]any) *StructTransfer {
	for k, v := range params {
		s.mapData[k] = v
	}
	return s
}
