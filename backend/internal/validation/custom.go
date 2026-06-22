// Package validation は Gin / go-playground/validator にカスタムルールを登録する。
// タグだけでは表しにくいドメイン検証を、再利用可能なルールとして切り出す。
package validation

import (
	"time"

	"github.com/KENTA0326/run-sync-pro/internal/timeutil"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// RegisterGinBindingValidators は process 全体で一度呼ぶ（main とテストの TestMain など）。
func RegisterGinBindingValidators() {
	v, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		return
	}
	_ = registerCustom(v)
}

func registerCustom(v *validator.Validate) error {
	if err := v.RegisterValidation("caldate", validateCalDate); err != nil {
		return err
	}
	return v.RegisterValidation("caldate_not_future", validateCalDateNotFutureJST)
}

// validateCalDate は "YYYY-MM-DD" を JST カレンダー日として解釈できるか（timeutil.ParseCalendarDate と同義）。
func validateCalDate(fl validator.FieldLevel) bool {
	s, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}
	if s == "" {
		return true
	}
	_, err := timeutil.ParseCalendarDate(s)
	return err == nil
}

// validateCalDateNotFutureJST はカレンダー日が「日本の今日」より後でないこと。
// caldate と併用すること（単体では不正な文字列は false）。
func validateCalDateNotFutureJST(fl validator.FieldLevel) bool {
	s, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}
	if s == "" {
		return true
	}
	t, err := timeutil.ParseCalendarDate(s)
	if err != nil {
		return false
	}
	jst := timeutil.JST
	tDay := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, jst)
	now := timeutil.NowJST()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, jst)
	return !tDay.After(today)
}
