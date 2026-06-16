package handler

const (
	defaultListLimit = 50
	maxListLimit     = 500
)

// listPaginationQuery は一覧 GET の limit / offset（cursor は将来拡張）。
type listPaginationQuery struct {
	Limit  int `form:"limit" binding:"omitempty,min=1,max=500"`
	Offset int `form:"offset" binding:"omitempty,min=0"`
}

func (q *listPaginationQuery) normalize() (limit, offset int) {
	limit = q.Limit
	if limit <= 0 {
		limit = defaultListLimit
	}
	if limit > maxListLimit {
		limit = maxListLimit
	}
	offset = q.Offset
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

// PaginationMeta はページネーション メタ情報。
type PaginationMeta struct {
	Limit  int   `json:"limit"`
	Offset int   `json:"offset"`
	Total  int64 `json:"total"`
}
