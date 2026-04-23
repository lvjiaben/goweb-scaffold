package bootstrap

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/lvjiaben/goweb-core/httpx"
	"gorm.io/datatypes"
)

type IDsPayload struct {
	ID  int64   `json:"id"`
	IDs []int64 `json:"ids"`
}

func (p IDsPayload) Values() []int64 {
	return NormalizeIDs(append([]int64{p.ID}, p.IDs...)...)
}

func NormalizeIDs(values ...int64) []int64 {
	seen := make(map[int64]struct{}, len(values))
	out := make([]int64, 0, len(values))
	for _, value := range values {
		if value <= 0 {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func Pagination(c *httpx.Context) (page int, pageSize int) {
	page = 1
	pageSize = 20

	if value, err := c.QueryInt64("page"); err == nil && value > 0 {
		page = int(value)
	}
	for _, key := range []string{"page_size", "pageSize", "limit"} {
		if value, err := c.QueryInt64(key); err == nil && value > 0 {
			pageSize = int(value)
			break
		}
	}
	if pageSize > 200 {
		pageSize = 200
	}
	return page, pageSize
}

func QueryFirst(c *httpx.Context, keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(c.Query(key)); value != "" {
			return value
		}
	}
	return ""
}

func SearchKeyword(c *httpx.Context) string {
	return QueryFirst(c, "search", "keyword")
}

func ParseFilter(c *httpx.Context) map[string]any {
	raw := strings.TrimSpace(c.Query("filter"))
	if raw == "" {
		return map[string]any{}
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return map[string]any{}
	}
	return payload
}

func FilterString(filters map[string]any, keys ...string) string {
	for _, key := range keys {
		value, ok := filters[key]
		if !ok || value == nil {
			continue
		}
		switch typed := value.(type) {
		case string:
			if trimmed := strings.TrimSpace(typed); trimmed != "" {
				return trimmed
			}
		default:
			text := strings.TrimSpace(fmt.Sprint(typed))
			if text != "" && text != "<nil>" {
				return text
			}
		}
	}
	return ""
}

func FilterInt64(filters map[string]any, keys ...string) (int64, bool) {
	value := FilterString(filters, keys...)
	if value == "" {
		return 0, false
	}
	var result int64
	if _, err := fmt.Sscan(value, &result); err != nil {
		return 0, false
	}
	return result, true
}

func FilterRange(filters map[string]any, keys ...string) (string, string, bool) {
	for _, key := range keys {
		value, ok := filters[key]
		if !ok || value == nil {
			continue
		}
		switch typed := value.(type) {
		case []any:
			if len(typed) < 2 {
				continue
			}
			start := strings.TrimSpace(fmt.Sprint(typed[0]))
			end := strings.TrimSpace(fmt.Sprint(typed[1]))
			if start == "" && end == "" {
				continue
			}
			return start, end, true
		case []string:
			if len(typed) < 2 {
				continue
			}
			start := strings.TrimSpace(typed[0])
			end := strings.TrimSpace(typed[1])
			if start == "" && end == "" {
				continue
			}
			return start, end, true
		}
	}
	return "", "", false
}

func PagedResult(list any, total int64, page int, pageSize int) map[string]any {
	return map[string]any{
		"list":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
		"limit":     pageSize,
	}
}

func LikeKeyword(keyword string) string {
	trimmed := strings.TrimSpace(keyword)
	if trimmed == "" {
		return ""
	}
	return "%" + trimmed + "%"
}

func JSONValue(value any) datatypes.JSON {
	if value == nil {
		return datatypes.JSON([]byte("{}"))
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return datatypes.JSON([]byte("{}"))
	}
	return datatypes.JSON(raw)
}
