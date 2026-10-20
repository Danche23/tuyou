package personal

import "testing"

func TestMask_NilSafe(t *testing.T) {
	svc := &personalService{sensitive: nil}
	if got := svc.mask("abc"); got != "abc" {
		t.Fatalf("nil 匹配器应原样返回，got %q", got)
	}
}
