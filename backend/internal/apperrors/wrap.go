package apperrors

import "fmt"

// Annotate は操作コンテキストを付加する（ログ・調査向け）。
func Annotate(operation string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", operation, err)
}
