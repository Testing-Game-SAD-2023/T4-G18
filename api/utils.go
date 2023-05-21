package api

import (
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PaginationParams struct {
	Page     int64
	PageSize int64
}

type IntervalParams struct {
	Start time.Time
	End   time.Time
}

type PaginatedResponse struct {
	Data     any                `json:"data"`
	Metadata PaginationMetadata `json:"metadata"`
}

type PaginationMetadata struct {
	HasNext  bool  `json:"hasNext"`
	Count    int64 `json:"count"`
	Page     int64 `json:"page"`
	PageSize int64 `json:"pageSize"`
}

func WithPagination(p *PaginationParams) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		offset := (p.Page - 1) * p.PageSize
		return db.Offset(int(offset)).Limit(int(p.PageSize))
	}
}

func WithInterval(i *IntervalParams) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("created_at between ? AND ?", i.Start, i.End)
	}
}

func WithOrder(column string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Order(clause.OrderBy{
			Columns: []clause.OrderByColumn{
				{
					Column: clause.Column{
						Name: column,
					},
				},
			},
		})
	}
}

func MakePaginatedResponse(v any, count int64, p *PaginationParams) *PaginatedResponse {
	return &PaginatedResponse{
		Data: v,
		Metadata: PaginationMetadata{
			Count:    count,
			HasNext:  (count - p.Page*p.PageSize) > 0,
			Page:     p.Page,
			PageSize: p.PageSize,
		},
	}
}

func byteCountIEC(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB",
		float64(b)/float64(div), "KMGTPE"[exp])
}
func Duplicated(v []string) bool {
	unique := make(map[string]struct{}, len(v))
	for _, item := range v {
		if _, seen := unique[item]; seen {
			return true
		}
		unique[item] = struct{}{}
	}
	return false
}
