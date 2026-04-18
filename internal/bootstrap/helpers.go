package bootstrap

import (
	"encoding/json"
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
	if value, err := c.QueryInt64("page_size"); err == nil && value > 0 {
		pageSize = int(value)
	}
	if pageSize > 200 {
		pageSize = 200
	}
	return page, pageSize
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
