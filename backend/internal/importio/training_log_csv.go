package importio

import (
	"bufio"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/KENTA0326/run-sync-pro/internal/apperrors"
	"github.com/KENTA0326/run-sync-pro/internal/timeutil"
)

// TrainingLogCSVRecord は走行ログ CSV の 1 行分（DB 投入前）。
type TrainingLogCSVRecord struct {
	TrainingDate string
	Distance     float64
	Duration     int
	Pace         string
	Kind         int
	ShoeID       uint
	Memo         string
}

// インポート時に必須の列（memo は任意列として別途許可）。
var trainingLogCSVRequiredColumns = []string{
	"training_date", "distance", "duration", "pace", "kind", "shoe_id",
}

// StreamTrainingLogCSV は一時ファイル上の CSV を bufio で読み、1 行ずつ fn に渡す。
// 戻り値は (処理件数, error)。1 行目はヘッダー行（列名）必須。memo 列は任意。
// ctx がキャンセルされたら読み取りループを中断する（リクエストタイムアウト・クライアント切断対策）。
func StreamTrainingLogCSV(ctx context.Context, path string, maxItems int, fn func(TrainingLogCSVRecord) error) (int, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if maxItems <= 0 {
		return 0, fmt.Errorf("importio: maxItems must be positive")
	}

	f, err := os.Open(path)
	if err != nil {
		return 0, fmt.Errorf("importio: open csv: %w", err)
	}
	defer f.Close()

	// 読み取り専用で共有ロック（同一ファイルへの並行インポートのサンプル）。
	// defer は LIFO: 先に funlock → 次に f.Close（閉じる前にアンロックする正しい順序）。
	if err := flockShared(f); err != nil {
		return 0, err
	}
	defer func() { _ = funlock(f) }()

	br := bufio.NewReader(f)
	cr := csv.NewReader(br)
	cr.FieldsPerRecord = -1
	// ReuseRecord: 行バッファを使い回してメモリ負荷を抑える。
	// fn には parse 済みの TrainingLogCSVRecord（値コピー）だけ渡す。row スライスや goroutine への生渡しは禁止。
	cr.ReuseRecord = true

	if err := ctx.Err(); err != nil {
		return 0, fmt.Errorf("importio: csv read cancelled: %w", err)
	}

	header, err := cr.Read()
	if err != nil {
		return 0, apperrors.BadRequest("CSV ヘッダーを読み取れませんでした", err)
	}
	colIndex, err := mapTrainingLogCSVHeader(header)
	if err != nil {
		return 0, err
	}

	count := 0
	for {
		if err := ctx.Err(); err != nil {
			return count, fmt.Errorf("importio: csv read cancelled: %w", err)
		}

		row, err := cr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return count, apperrors.BadRequest(fmt.Sprintf("CSV %d 行目の読み取りに失敗しました", count+2), err)
		}
		if isEmptyCSVRow(row) {
			continue
		}
		if count >= maxItems {
			return count, apperrors.BadRequest(fmt.Sprintf("1回の取り込み上限（%d件）を超えています", maxItems))
		}

		rec, err := parseTrainingLogCSVRow(row, colIndex)
		if err != nil {
			return count, apperrors.BadRequest(fmt.Sprintf("CSV %d 行目: %s", count+2, err.Error()), err)
		}
		if err := fn(rec); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

func mapTrainingLogCSVHeader(header []string) (map[string]int, error) {
	idx := make(map[string]int, len(header))
	for i, raw := range header {
		name := strings.TrimSpace(strings.ToLower(raw))
		if name == "" {
			continue
		}
		if _, dup := idx[name]; dup {
			return nil, apperrors.BadRequest(fmt.Sprintf("CSV ヘッダーが重複しています: %s", name))
		}
		idx[name] = i
	}
	for _, required := range trainingLogCSVRequiredColumns {
		if _, ok := idx[required]; !ok {
			return nil, apperrors.BadRequest(fmt.Sprintf("CSV ヘッダーに %s 列が必要です", required))
		}
	}
	return idx, nil
}

func parseTrainingLogCSVRow(row []string, col map[string]int) (TrainingLogCSVRecord, error) {
	get := func(name string) string {
		i := col[name]
		if i >= len(row) {
			return ""
		}
		return strings.TrimSpace(row[i])
	}

	dateStr := get("training_date")
	if _, err := timeutil.ParseCalendarDate(dateStr); err != nil {
		return TrainingLogCSVRecord{}, fmt.Errorf("training_date は YYYY-MM-DD 形式で指定してください")
	}

	dist, err := strconv.ParseFloat(get("distance"), 64)
	if err != nil || dist <= 0 {
		return TrainingLogCSVRecord{}, fmt.Errorf("distance は正の数値で指定してください")
	}

	dur, err := strconv.Atoi(get("duration"))
	if err != nil || dur <= 0 {
		return TrainingLogCSVRecord{}, fmt.Errorf("duration は正の整数で指定してください")
	}

	pace := get("pace")
	if pace == "" {
		return TrainingLogCSVRecord{}, fmt.Errorf("pace は必須です")
	}

	kind, err := strconv.Atoi(get("kind"))
	if err != nil || kind < 0 || kind > 4 {
		return TrainingLogCSVRecord{}, fmt.Errorf("kind は 0-4 の範囲で指定してください")
	}

	shoeID, err := strconv.ParseUint(get("shoe_id"), 10, 64)
	if err != nil || shoeID == 0 {
		return TrainingLogCSVRecord{}, fmt.Errorf("shoe_id は正の整数で指定してください")
	}

	memo := ""
	if i, ok := col["memo"]; ok && i < len(row) {
		memo = strings.TrimSpace(row[i])
	}

	return TrainingLogCSVRecord{
		TrainingDate: dateStr,
		Distance:     dist,
		Duration:     dur,
		Pace:         pace,
		Kind:         kind,
		ShoeID:       uint(shoeID),
		Memo:         memo,
	}, nil
}

func isEmptyCSVRow(row []string) bool {
	for _, c := range row {
		if strings.TrimSpace(c) != "" {
			return false
		}
	}
	return true
}
