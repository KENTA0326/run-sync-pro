package database

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// WithTx は db 上でトランザクションを開始し、コールバック fn をその tx で実行する。
//
// defer は LIFO（後から登録されたものから実行される）。
//   - 後から登録された Rollback 用 defer が先に走る。その時点ではパニック直後は err が未設定なことが多く、Rollback しない。
//   - 続いて panic 捕捉用 defer が実行され、recover が Rollback し、panic を error に載せて返す（名前付き戻り値 err）。
//
// 名前付き戻り値 err により、内側 defer が return 直前の err（Commit 失敗を含む）を参照できる。
func WithTx(ctx context.Context, db *gorm.DB, fn func(tx *gorm.DB) error) (err error) {
	tx := db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// 1つ目に登録 → パニック時は 2 番目に実行
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			if pe, ok := p.(error); ok {
				err = fmt.Errorf("database WithTx: panic recovered: %w", pe)
			} else {
				err = fmt.Errorf("database WithTx: panic recovered: %v", p)
			}
		}
	}()

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if err = fn(tx); err != nil {
		return err
	}

	err = tx.Commit().Error
	return err
}

// WithGlobalTx はパッケージ変数 DB で WithTx を呼ぶ糖衣構文。
func WithGlobalTx(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return WithTx(ctx, DB, fn)
}
