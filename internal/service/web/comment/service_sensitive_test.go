package comment

import (
	"testing"

	"tuyou/internal/model/entity"
	"tuyou/pkg/sensitive"
)

func TestSensitiveSnapshot_Hit(t *testing.T) {
	m := sensitive.New([]string{"赌博", "毒品"})
	isSensitive, hitWords := sensitiveSnapshot(m, "这个赌博网站，还有贩卖毒品")
	if isSensitive != 1 {
		t.Fatalf("命中应 is_sensitive=1，got %d", isSensitive)
	}
	if hitWords != "赌博;毒品" {
		t.Fatalf("命中词应分号连接，got %q", hitWords)
	}
}

func TestSensitiveSnapshot_Miss(t *testing.T) {
	m := sensitive.New([]string{"赌博"})
	isSensitive, hitWords := sensitiveSnapshot(m, "今天去爬山")
	if isSensitive != 0 || hitWords != "" {
		t.Fatalf("未命中应 (0,\"\")，got (%d,%q)", isSensitive, hitWords)
	}
}

func TestSensitiveSnapshot_NilMatcher(t *testing.T) {
	isSensitive, hitWords := sensitiveSnapshot(nil, "任何内容")
	if isSensitive != 0 || hitWords != "" {
		t.Fatalf("nil 匹配器应 (0,\"\")，got (%d,%q)", isSensitive, hitWords)
	}
}

func TestMaskContent_NilSafe(t *testing.T) {
	s := &commentService{sensitive: nil}
	if got := s.maskContent("abc"); got != "abc" {
		t.Fatalf("nil 匹配器应原样返回，got %q", got)
	}
}

func TestCommentUser_MasksNickname(t *testing.T) {
	m := sensitive.New([]string{"赌博"})
	svc := &commentService{sensitive: m}
	u := svc.commentUser(&entity.Member{BaseEntity: entity.BaseEntity{ID: 1}, Nickname: "赌博哥"})
	if u.Nickname != "**哥" {
		t.Fatalf("昵称应打星，got %q", u.Nickname)
	}
	u2 := svc.commentUser(nil)
	if u2.Nickname != "" {
		t.Fatalf("nil 用户应返回空昵称，got %q", u2.Nickname)
	}
}

func TestReplyToNickname_Masks(t *testing.T) {
	m := sensitive.New([]string{"毒品"})
	svc := &commentService{sensitive: m}
	mid := uint(1)
	members := map[uint]*entity.Member{1: {BaseEntity: entity.BaseEntity{ID: 1}, Nickname: "贩卖毒品"}}
	if got := svc.replyToNickname(&mid, members); got != "贩卖**" {
		t.Fatalf("被回复昵称应打星，got %q", got)
	}
}
