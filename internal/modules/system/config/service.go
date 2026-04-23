package system_config

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/lvjiaben/goweb-scaffold/internal/bootstrap"
)

type Service struct {
	repo *Repo
}

func NewService(runtime *bootstrap.Runtime) *Service {
	return &Service{repo: NewRepo(runtime)}
}

func (s *Service) List(params ListParams) (map[string]any, error) {
	filter := configListFilter{
		Keyword: bootstrap.LikeKeyword(params.Keyword),
		Key:     bootstrap.LikeKeyword(bootstrap.FilterString(params.Filters, "config_key", "key")),
		Name:    bootstrap.LikeKeyword(bootstrap.FilterString(params.Filters, "config_name", "name")),
	}
	total, err := s.repo.Count(filter)
	if err != nil {
		return nil, err
	}
	rows, err := s.repo.List(filter, params.Page, params.PageSize)
	if err != nil {
		return nil, err
	}
	items := make([]ItemResponse, 0, len(rows))
	for _, item := range rows {
		items = append(items, configResponse(item))
	}
	return bootstrap.PagedResult(items, total, params.Page, params.PageSize), nil
}

func (s *Service) Detail(id int64) (ItemResponse, error) {
	item, err := s.repo.FindByID(id)
	if err != nil {
		return ItemResponse{}, err
	}
	return configResponse(item), nil
}

func (s *Service) SaveConfig(req SaveRequest) (SaveResult, error) {
	req.ConfigKey = strings.TrimSpace(req.ConfigKey)
	req.ConfigName = strings.TrimSpace(req.ConfigName)
	req.Remark = strings.TrimSpace(req.Remark)
	if req.ConfigKey == "" {
		return SaveResult{}, validationError("config_key is required")
	}
	if req.ConfigName == "" {
		return SaveResult{}, validationError("config_name is required")
	}
	if len(req.ConfigValue) == 0 {
		req.ConfigValue = json.RawMessage(`{}`)
	}
	if err := s.ensureConfigKeyUnique(req.ID, req.ConfigKey); err != nil {
		return SaveResult{}, validationError(err.Error())
	}

	if req.ID == 0 {
		item := SystemConfig{
			ConfigKey:   req.ConfigKey,
			ConfigName:  req.ConfigName,
			ConfigValue: JSON(req.ConfigValue),
			Remark:      req.Remark,
		}
		if err := s.repo.Create(&item); err != nil {
			return SaveResult{}, err
		}
		return SaveResult{ID: item.ID}, nil
	}

	item, err := s.repo.FindByID(req.ID)
	if err != nil {
		return SaveResult{}, err
	}
	if err := s.repo.Update(&item, map[string]any{
		"config_key":   req.ConfigKey,
		"config_name":  req.ConfigName,
		"config_value": JSON(req.ConfigValue),
		"remark":       req.Remark,
		"updated_at":   time.Now(),
	}); err != nil {
		return SaveResult{}, err
	}
	return SaveResult{ID: item.ID}, nil
}

func (s *Service) DeleteConfigs(ids []int64) (DeleteResult, error) {
	ids = bootstrap.NormalizeIDs(ids...)
	if len(ids) == 0 {
		return DeleteResult{}, validationError("ids is required")
	}
	if err := s.repo.DeleteByIDs(ids); err != nil {
		return DeleteResult{}, err
	}
	return DeleteResult{Deleted: len(ids)}, nil
}

func (s *Service) ensureConfigKeyUnique(currentID int64, key string) error {
	count, err := s.repo.CountByKey(key, currentID)
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("config_key already exists")
	}
	return nil
}

func configResponse(item SystemConfig) ItemResponse {
	return ItemResponse{
		ID:          item.ID,
		Dir:         "system",
		ConfigKey:   item.ConfigKey,
		ConfigName:  item.ConfigName,
		ConfigValue: json.RawMessage(item.ConfigValue),
		Remark:      item.Remark,
		Type:        "input",
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
	}
}

type validationError string

func (e validationError) Error() string {
	return string(e)
}

func isValidationError(err error) bool {
	_, ok := err.(validationError)
	return ok
}
