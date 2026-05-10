package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"net/mail"
	"strings"
)

// ErrInvalidEmail は NewEmail が受理しない入力に使う。
var ErrInvalidEmail = errors.New("invalid email")

// Email は検証済みメールアドレス（値は外部から直接書き換え不可）。
type Email struct {
	value string
}

// NewEmail はトリムと RFC5322 準拠の単一アドレス解析で検証し、不正なら ErrInvalidEmail を返す。
func NewEmail(raw string) (Email, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return Email{}, ErrInvalidEmail
	}
	addr, err := mail.ParseAddress(s)
	if err != nil {
		return Email{}, ErrInvalidEmail
	}
	normalized := strings.TrimSpace(addr.Address)
	if normalized == "" {
		return Email{}, ErrInvalidEmail
	}
	return Email{value: normalized}, nil
}

// String は永続化・クエリ用のメール文字列を返す（値レシーバー: ミューテーションなし）。
func (e Email) String() string {
	return e.value
}

// Value は GORM / database/sql からカラムへ書き込む値を返す。
func (e Email) Value() (driver.Value, error) {
	if e.value == "" {
		return "", nil
	}
	return e.value, nil
}

// Scan は DB から読み込んだ値を Email に載せる（既存行が不正な場合はエラー）。
func (e *Email) Scan(src interface{}) error {
	switch v := src.(type) {
	case nil:
		*e = Email{}
		return nil
	case []byte:
		return e.scanString(string(v))
	case string:
		return e.scanString(v)
	default:
		return fmt.Errorf("email: unsupported Scan type %T", src)
	}
}

func (e *Email) scanString(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		*e = Email{}
		return nil
	}
	nv, err := NewEmail(s)
	if err != nil {
		return fmt.Errorf("%w: %q", err, s)
	}
	*e = nv
	return nil
}

// MarshalJSON は API では文字列として出力する（非公開フィールドをそのまま JSON に載せない）。
func (e Email) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.value)
}

// UnmarshalJSON は JSON の文字列から検証済み Email を組み立てる。
func (e *Email) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	nv, err := NewEmail(s)
	if err != nil {
		return fmt.Errorf("email json: %w", err)
	}
	*e = nv
	return nil
}
