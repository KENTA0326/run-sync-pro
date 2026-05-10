package testutil

import (
	"encoding/json"
	"net/http/httptest"
	"reflect"
	"testing"
)

// AssertResponseJSON は recorder のステータスと JSON ボディ全体を want と突き合わせる。
func AssertResponseJSON(t *testing.T, rec *httptest.ResponseRecorder, wantCode int, wantBody any) {
	t.Helper()
	if rec.Code != wantCode {
		t.Fatalf("HTTP status: got %d want %d; body=%s", rec.Code, wantCode, rec.Body.String())
	}
	var got any
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("response JSON decode: %v; raw=%q", err, rec.Body.String())
	}
	if !reflect.DeepEqual(got, wantBody) {
		t.Fatalf("response JSON mismatch\ngot:  %#v\nwant: %#v\nraw: %s", got, wantBody, rec.Body.String())
	}
}
