package database

import (
	"context"

	"gorm.io/gorm"
)

// WithTx は db 上でトランザクションを開始し、コールバック fn をその tx で実行する。
//
// defer は LIFO（後から登録したものが先に実行される）。この関数では
//   1) パニック捕捉（recover）を先に登録 → 終了時には後から実行
//   2) 通常のエラー時 Rollback を後に登録 → 終了時には先に実行
// と並べている。パニック時は (2) が先に走るが、(2) は err のみ見るため
// パニック直後は多くの場合 err が未設定のままなので Rollback しない。(1) が recover し、Rollback してから再 panic する。
//
// 名前付き戻り値 err により、内側 defer が return 直前の err（Commit 失敗を含む）を参照できる。
//
// defer に渡すクロージャは登録時に評価され、本体は関数終了時に走る。キャプチャする err は
// ポインタではなく「名前付き戻り値のスロット」として参照されるため、終了時点の値が見える。
func WithTx(ctx context.Context, db *gorm.DB, fn func(tx *gorm.DB) error) (err error) {
	tx := db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// 1つ目に登録 → 関数終了時には 2 番目に実行（LIFO）
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p) // 呼び出し元へパニックを伝播（recover で握りつぶすならここを変える）
		}
	}()

	// 2つ目に登録 → 先に実行。パニック時は err が未設定のことが多く、この分岐では Rollback しない。
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
