package codegen

import (
	"strings"

	"github.com/lvjiaben/goweb-scaffold/internal/bootstrap"
	genservice "github.com/lvjiaben/goweb-scaffold/internal/gen/service"
	"gorm.io/gorm"
)

type Repo struct {
	db *gorm.DB
}

func NewRepo(runtime *bootstrap.Runtime) *Repo {
	if runtime == nil {
		return &Repo{}
	}
	return &Repo{db: runtime.DB}
}

func (r *Repo) ListHistory() ([]CodegenHistory, error) {
	var rows []CodegenHistory
	if r.db == nil {
		return rows, nil
	}
	err := r.db.Order("id DESC").Find(&rows).Error
	return rows, err
}

func (r *Repo) CreateHistory(record *CodegenHistory) error {
	if r.db == nil {
		return nil
	}
	return r.db.Create(record).Error
}

func (r *Repo) SaveHistory(record *CodegenHistory) error {
	if r.db == nil {
		return nil
	}
	return r.db.Save(record).Error
}

func (r *Repo) DeleteHistory(ids []int64) error {
	if r.db == nil {
		return nil
	}
	return r.db.Where("id IN ?", ids).Delete(&CodegenHistory{}).Error
}

func (r *Repo) HistoryByID(id int64) (CodegenHistory, error) {
	var row CodegenHistory
	if r.db == nil {
		return row, gorm.ErrRecordNotFound
	}
	err := r.db.First(&row, id).Error
	return row, err
}

func (r *Repo) LatestHistoryByModule(moduleName string) (CodegenHistory, error) {
	var row CodegenHistory
	if r.db == nil {
		return row, gorm.ErrRecordNotFound
	}
	err := r.db.Where("module_name = ?", strings.TrimSpace(moduleName)).Order("id DESC").First(&row).Error
	return row, err
}

func (r *Repo) BusinessTables() ([]BusinessTable, error) {
	var rows []tableInfo
	if r.db == nil {
		return []BusinessTable{}, nil
	}
	if err := r.db.Raw(`
SELECT
  t.table_name,
  COALESCE(obj_description(to_regclass(format('%I.%I', t.table_schema, t.table_name)), 'pg_class'), '') AS table_comment
FROM information_schema.tables t
WHERE t.table_schema = 'public'
  AND table_type = 'BASE TABLE'
ORDER BY t.table_name ASC`).Scan(&rows).Error; err != nil {
		return nil, err
	}

	items := make([]BusinessTable, 0, len(rows))
	for _, row := range rows {
		if isExcludedTable(row.TableName) {
			continue
		}
		displayName := strings.TrimSpace(row.TableComment)
		if displayName == "" {
			displayName = row.TableName
		}
		items = append(items, BusinessTable{
			TableName:    row.TableName,
			DisplayName:  displayName,
			TableComment: strings.TrimSpace(row.TableComment),
		})
	}
	return items, nil
}

func (r *Repo) TableColumns(tableName string) ([]genservice.ColumnInfo, error) {
	if isExcludedTable(tableName) || r.db == nil {
		return []genservice.ColumnInfo{}, nil
	}

	var rows []genservice.ColumnInfo
	err := r.db.Raw(`
SELECT
  c.column_name,
  c.data_type,
  (c.is_nullable = 'YES') AS is_nullable,
  COALESCE(c.column_default, '') AS column_default,
  c.ordinal_position,
  COALESCE(pk.is_primary_key, FALSE) AS is_primary_key,
  COALESCE(col_description(to_regclass(format('%I.%I', c.table_schema, c.table_name)), c.ordinal_position), '') AS column_comment,
  COALESCE(obj_description(to_regclass(format('%I.%I', c.table_schema, c.table_name)), 'pg_class'), '') AS table_comment
FROM information_schema.columns c
LEFT JOIN (
  SELECT
    kcu.table_name,
    kcu.column_name,
    TRUE AS is_primary_key
  FROM information_schema.table_constraints tc
  JOIN information_schema.key_column_usage kcu
    ON tc.constraint_name = kcu.constraint_name
   AND tc.table_schema = kcu.table_schema
  WHERE tc.constraint_type = 'PRIMARY KEY'
    AND tc.table_schema = 'public'
) pk
  ON pk.table_name = c.table_name
 AND pk.column_name = c.column_name
WHERE c.table_schema = 'public'
  AND c.table_name = ?
ORDER BY c.ordinal_position ASC`, tableName).Scan(&rows).Error
	return rows, err
}
