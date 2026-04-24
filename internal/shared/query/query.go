package query

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"gorm.io/gorm"
)

var safeColumn = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

type Params struct {
	Page      int
	PageSize  int
	Search    string
	Filters   map[string]any
	SortBy    string
	SortOrder string
}

type Options struct {
	SearchFields []string
	LikeFields   []string
	ExactFields  []string
	RangeFields  []string
	AllowedSorts []string
	DefaultSorts []string
}

type Result struct {
	Query *gorm.DB
	Count *gorm.DB
}

func Apply(db *gorm.DB, params Params, options Options) Result {
	query := db
	count := db.Session(&gorm.Session{})

	if keyword := strings.TrimSpace(params.Search); keyword != "" && len(options.SearchFields) > 0 {
		likes := make([]string, 0, len(options.SearchFields))
		values := make([]any, 0, len(options.SearchFields))
		for _, field := range options.SearchFields {
			if !isAllowedColumn(field) {
				continue
			}
			likes = append(likes, field+" ILIKE ?")
			values = append(values, "%"+keyword+"%")
		}
		if len(likes) > 0 {
			expr := strings.Join(likes, " OR ")
			query = query.Where(expr, values...)
			count = count.Where(expr, values...)
		}
	}

	for _, field := range options.LikeFields {
		if value, ok := nonEmptyString(params.Filters[field]); ok && isAllowedColumn(field) {
			query = query.Where(field+" ILIKE ?", "%"+value+"%")
			count = count.Where(field+" ILIKE ?", "%"+value+"%")
		}
	}
	for _, field := range options.ExactFields {
		if value, ok := nonEmpty(params.Filters[field]); ok && isAllowedColumn(field) {
			query = query.Where(field+" = ?", value)
			count = count.Where(field+" = ?", value)
		}
	}
	for _, field := range options.RangeFields {
		if !isAllowedColumn(field) {
			continue
		}
		from, _ := nonEmptyString(params.Filters[field+"_from"])
		to, _ := nonEmptyString(params.Filters[field+"_to"])
		if from != "" {
			query = query.Where(field+" >= ?", from)
			count = count.Where(field+" >= ?", from)
		}
		if to != "" {
			query = query.Where(field+" <= ?", to)
			count = count.Where(field+" <= ?", to)
		}
	}

	query = applySort(query, params, options)

	if params.PageSize > 0 {
		if params.Page < 1 {
			params.Page = 1
		}
		query = query.Offset((params.Page - 1) * params.PageSize).Limit(params.PageSize)
	}
	return Result{Query: query, Count: count}
}

func DecodeFilters(raw string) map[string]any {
	result := map[string]any{}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return result
	}
	_ = json.Unmarshal([]byte(raw), &result)
	return result
}

func DefaultSorts(fields ...string) []string {
	available := map[string]bool{}
	for _, field := range fields {
		available[field] = true
	}
	switch {
	case available["sort"]:
		return []string{"sort DESC", "id DESC"}
	case available["weight"]:
		return []string{"weight DESC", "id DESC"}
	case available["weigh"]:
		return []string{"weigh DESC", "id DESC"}
	default:
		return []string{"id DESC"}
	}
}

func applySort(db *gorm.DB, params Params, options Options) *gorm.DB {
	sortBy := strings.TrimSpace(params.SortBy)
	sortOrder := strings.ToLower(strings.TrimSpace(params.SortOrder))
	if sortOrder != "asc" {
		sortOrder = "desc"
	}
	if sortBy != "" && contains(options.AllowedSorts, sortBy) && isAllowedColumn(sortBy) {
		return db.Order(fmt.Sprintf("%s %s", sortBy, strings.ToUpper(sortOrder)))
	}
	for _, order := range options.DefaultSorts {
		if strings.TrimSpace(order) != "" {
			db = db.Order(order)
		}
	}
	return db
}

func contains(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func isAllowedColumn(column string) bool {
	return safeColumn.MatchString(column)
}

func nonEmpty(value any) (any, bool) {
	switch v := value.(type) {
	case nil:
		return nil, false
	case string:
		return strings.TrimSpace(v), strings.TrimSpace(v) != ""
	case float64, int, int64, bool:
		return v, true
	default:
		return v, true
	}
}

func nonEmptyString(value any) (string, bool) {
	v, ok := nonEmpty(value)
	if !ok {
		return "", false
	}
	s := strings.TrimSpace(fmt.Sprint(v))
	return s, s != ""
}
