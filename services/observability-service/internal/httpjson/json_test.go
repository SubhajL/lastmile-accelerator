package httpjson

import (
	"bytes"
	"net/http/httptest"
	"strings"
	"testing"
)

type sample struct{ A string `json:"a"` }

func TestStrictDecode_UnknownField_ReturnsInvalid(t *testing.T) {
	r := httptest.NewRequest("POST", "/", bytes.NewBufferString(`{"a":"x","extra":1}`))
	rec := httptest.NewRecorder()
	var s sample
	err := StrictDecode(rec, r, &s, 1024)
	if err == nil {
		t.Fatal("expected error for unknown field")
	}
}

func TestStrictDecode_TooLarge_ReturnsTooLarge(t *testing.T) {
	big := strings.Repeat("a", 2048)
	body := "{\"a\":\"" + big + "\"}"
	r := httptest.NewRequest("POST", "/", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()
	var s sample
	err := StrictDecode(rec, r, &s, 10) // very small limit
	if err == nil || err != ErrTooLarge {
		t.Fatalf("expected ErrTooLarge, got %v", err)
	}
}
