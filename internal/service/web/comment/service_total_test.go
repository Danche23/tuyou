package comment

import (
	"testing"

	"tuyou/internal/model/entity"
)

// cmt 构造一个只有 ID 的评论（BaseEntity 是跨包嵌入，提升字段不能直接出现在字面量里，需显式嵌 BaseEntity）
func cmt(id uint) entity.Comment {
	return entity.Comment{BaseEntity: entity.BaseEntity{ID: id}}
}

func TestHotExcludeIDs_UnderOrEqualTwo(t *testing.T) {
	top := []entity.Comment{cmt(1), cmt(2)}
	if got := hotExcludeIDs(top, 2); got != nil {
		t.Fatalf("total=2 时应返回 nil（不排除热评），实际 %v", got)
	}
	if got := hotExcludeIDs(top, 1); got != nil {
		t.Fatalf("total=1 时应返回 nil，实际 %v", got)
	}
}

func TestHotExcludeIDs_OverTwo(t *testing.T) {
	top := []entity.Comment{cmt(1), cmt(2)}
	got := hotExcludeIDs(top, 3)
	if len(got) != 2 || got[0] != 1 || got[1] != 2 {
		t.Fatalf("total=3 时应排除热评 [1 2]，实际 %v", got)
	}
}

func TestDedupeComments_KeepsOrder(t *testing.T) {
	in := []entity.Comment{cmt(1), cmt(2), cmt(1), cmt(3)}
	got := dedupeComments(in)
	if len(got) != 3 || got[0].ID != 1 || got[1].ID != 2 || got[2].ID != 3 {
		t.Fatalf("去重应保序且去重：[1 2 3]，实际 %v", got)
	}
}
