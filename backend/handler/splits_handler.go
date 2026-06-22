package handler

import (
	"net/http"

	"github.com/KENTA0326/run-sync-pro/internal/apperrors"
	"github.com/KENTA0326/run-sync-pro/model"
	"github.com/gin-gonic/gin"
)

type FullMarathonSplitsInput struct {
	PaceSecPerKm float64 `json:"pace_sec_per_km" binding:"required,gt=0"`
	Page         int     `json:"page" binding:"omitempty,gt=0"`
}

type FullMarathonSplitsResponse struct {
	Page       int                `json:"page"`
	TotalPages int                `json:"total_pages"`
	Rows       []model.SplitRow `json:"rows"`
}

// FullMarathonSplits は「フルマラソン(42.195km)の1kmごとの通過タイム」を返す（10km単位ページング）。
// POST /api/v1/splits/full-marathon （レガシー: POST /splits/fullmarathon）
func (h *Handlers) FullMarathonSplits(c *gin.Context) {
	var input FullMarathonSplitsInput
	if err := c.ShouldBindJSON(&input); err != nil {
		respondHTTPError(c, apperrors.BadRequest("pace_sec_per_km を正しく指定してください", err))
		return
	}
	page := input.Page
	if page <= 0 {
		page = 1
	}

	all := h.splits.GenerateFullMarathonSplits(input.PaceSecPerKm)
	if len(all) == 0 {
		respondHTTPError(c, apperrors.BadRequest("スプリットを生成できません"))
		return
	}

	// 10km単位ページング（Finish行は最後のページに入る）
	perPage := 10
	totalRows := len(all)
	totalPages := (totalRows + perPage - 1) / perPage
	if page > totalPages {
		page = totalPages
	}
	start := (page - 1) * perPage
	end := start + perPage
	if end > totalRows {
		end = totalRows
	}

	c.JSON(http.StatusOK, FullMarathonSplitsResponse{
		Page:       page,
		TotalPages: totalPages,
		Rows:       all[start:end],
	})
}
