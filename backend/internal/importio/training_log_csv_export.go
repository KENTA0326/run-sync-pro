package importio

import (
	"bufio"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
)

// TrainingLogCSVHeader はインポート互換の CSV 1 行目（エクスポートでも同じ列順）。
func TrainingLogCSVHeader() []string {
	return []string{"training_date", "distance", "duration", "pace", "kind", "shoe_id", "memo"}
}

func trainingLogCSVRecordToRow(rec TrainingLogCSVRecord) []string {
	return []string{
		rec.TrainingDate,
		strconv.FormatFloat(rec.Distance, 'f', -1, 64),
		strconv.Itoa(rec.Duration),
		rec.Pace,
		strconv.Itoa(rec.Kind),
		strconv.FormatUint(uint64(rec.ShoeID), 10),
		rec.Memo,
	}
}

// StreamWriteTrainingLogCSV は bufio + csv.Writer でレスポンス等へストリーム出力する。
// produce は emit を呼ぶたびに 1 行書き出す。大量件数でもバッチごとに Flush できる土台。
func StreamWriteTrainingLogCSV(
	ctx context.Context,
	w io.Writer,
	produce func(emit func(TrainingLogCSVRecord) error) error,
) (int, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	bw := bufio.NewWriter(w)
	cw := csv.NewWriter(bw)

	if err := ctx.Err(); err != nil {
		return 0, fmt.Errorf("importio: csv write cancelled: %w", err)
	}
	if err := cw.Write(TrainingLogCSVHeader()); err != nil {
		return 0, fmt.Errorf("importio: csv write header: %w", err)
	}

	count := 0
	emit := func(rec TrainingLogCSVRecord) error {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("importio: csv write cancelled: %w", err)
		}
		if err := cw.Write(trainingLogCSVRecordToRow(rec)); err != nil {
			return fmt.Errorf("importio: csv write row: %w", err)
		}
		count++
		// 一定件数ごとに下流へ流す（メモリだけでなくクライアントへの送信も進める）
		if count%200 == 0 {
			cw.Flush()
			if err := bw.Flush(); err != nil {
				return err
			}
		}
		return nil
	}

	if err := produce(emit); err != nil {
		return count, err
	}

	cw.Flush()
	if err := cw.Error(); err != nil {
		return count, fmt.Errorf("importio: csv writer: %w", err)
	}
	if err := bw.Flush(); err != nil {
		return count, err
	}
	return count, nil
}
