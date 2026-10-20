package main

import (
	"database/sql"
	"testing"
)

func TestEncodeValue(t *testing.T) {
	tests := []struct {
		name  string
		input sql.RawBytes
		want  string
	}{
		{"null", nil, "NULL"},
		{"empty", sql.RawBytes{}, "''"},
		{"text", sql.RawBytes("途友"), "CONVERT(0xe98094e58f8b USING utf8mb4)"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := encodeValue(test.input); got != test.want {
				t.Fatalf("encodeValue() = %q, want %q", got, test.want)
			}
		})
	}
}
