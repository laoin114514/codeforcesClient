// Package params 提供将 Go struct 编码为 Codeforces API 查询参数的工具。
package params

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// Encode 通过 JSON 往返将 struct 转换为 map[string]any。
// 可选参数 extra 会在编码后合并进去。
func Encode(v any, extra map[string]any) (map[string]any, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("marshal: %w", err)
	}
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}
	for k, val := range extra {
		result[k] = val
	}
	return result, nil
}

// ToOrderedString 将 map 转为按 key 字母序排列的 key=value 查询串。
// 用于签名时保证参数顺序一致。
func ToOrderedString(m map[string]any) string {
	if len(m) == 0 {
		return ""
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	for i, k := range keys {
		if i > 0 {
			b.WriteByte('&')
		}
		b.WriteString(k)
		b.WriteByte('=')
		b.WriteString(fmt.Sprintf("%v", m[k]))
	}
	return b.String()
}
