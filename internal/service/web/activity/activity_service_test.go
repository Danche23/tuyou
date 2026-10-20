package activity

import "testing"

func TestMask_NilSafe(t *testing.T) {
	svc := &activityService{sensitive: nil}
	if got := svc.mask("abc"); got != "abc" {
		t.Fatalf("nil 匹配器应原样返回，got %q", got)
	}
}
