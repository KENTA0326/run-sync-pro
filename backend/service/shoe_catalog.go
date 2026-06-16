package service

import (
	"strings"

	"github.com/KENTA0326/run-sync-pro/model"
)

// DistinctShoeBrands は登録シューズからブランド名の重複を除いた一覧を返す（出現順を保持）。
// MapSlice で DTO 変換、Unique で map キー用の comparable 制約を満たす重複除去に使う。
func DistinctShoeBrands(shoes []model.Shoe) []string {
	brands := MapSlice(shoes, func(s model.Shoe) string { return strings.TrimSpace(s.Brand) })
	nonEmpty := make([]string, 0, len(brands))
	for _, b := range brands {
		if b != "" {
			nonEmpty = append(nonEmpty, b)
		}
	}
	return Unique(nonEmpty)
}
