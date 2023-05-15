package main

import (
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func withPagination(p *PaginationParams) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		offset := (p.page - 1) * p.pageSize
		return db.Offset(int(offset)).Limit(int(p.pageSize))
	}
}

func withInterval(i *IntervalParams) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("created_at between ? AND ?", i.startDate, i.endDate)
	}
}

func withOrder(column string) func(db *gorm.DB) *gorm.DB {
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

func makePaginatedResponse(v any, count int64, p *PaginationParams) *PaginatedResponse {
	return &PaginatedResponse{
		Data: v,
		Metadata: PaginationMetadata{
			Count:    count,
			HasNext:  (count - p.page*p.pageSize) > 0,
			Page:     p.page,
			PageSize: p.pageSize,
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

func duplicated(v []string) bool {
	unique := make(map[string]struct{}, len(v))
	for _, item := range v {
		if _, seen := unique[item]; seen {
			return true
		}
		unique[item] = struct{}{}
	}
	return false
}
