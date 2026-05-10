package apperrors

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// FromGORM は GORM エラーをセンチネルに写し替える。既存連鎖は Unwrap で辿れる。
func FromGORM(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("%w: %w", ErrNotFound, err)
	}
	return err
}

// FromGORMAnn は FromGORM のあと Annotate で包む。
func FromGORMAnn(operation string, err error) error {
	return Annotate(operation, FromGORM(err))
}
