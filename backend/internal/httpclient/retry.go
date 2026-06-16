package httpclient

import (
	"errors"
	"net"
	"net/http"
)

// isRetryableStatus は一時的なサーバーエラー（5xx）か。
func isRetryableStatus(code int) bool {
	return code == http.StatusInternalServerError ||
		code == http.StatusBadGateway ||
		code == http.StatusServiceUnavailable ||
		code == http.StatusGatewayTimeout
}

// isRetryableNetErr はネットワーク一時障害か。
func isRetryableNetErr(err error) bool {
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}
	return errors.Is(err, http.ErrHandlerTimeout)
}
