package sensitive

import "testing"

// newTest 用词表构建匹配器
func newTest(words []string) *SensitiveMatcher { return New(words) }

func TestValidate_Hit(t *testing.T) {
	m := newTest([]string{"赌博", "毒品"})
	for _, content := range []string{"在线赌博平台", "贩卖毒品违法", "今天心情不好，赌博害人"} {
		word, hit := m.Validate(content)
		if !hit {
			t.Fatalf("内容 %q 应命中敏感词", content)
		}
		if word != "赌博" && word != "毒品" {
			t.Fatalf("返回的命中词 %q 不在词库中", word)
		}
	}
}

func TestValidate_Miss(t *testing.T) {
	m := newTest([]string{"赌博", "毒品"})
	for _, content := range []string{"今天去爬山了", "买了条裙子", ""} {
		if _, hit := m.Validate(content); hit {
			t.Fatalf("内容 %q 不应命中", content)
		}
	}
}

func TestValidate_PrefixBoundary(t *testing.T) {
	// 词库只有"法轮"（短词），"法拉利"只是前缀相同，不应命中
	m := newTest([]string{"法轮"})
	if _, hit := m.Validate("买了一台法拉利"); hit {
		t.Fatal("前缀词不应误判命中")
	}
	if word, hit := m.Validate("法轮功是邪教"); !hit || word != "法轮" {
		t.Fatal("应命中完整词 法轮")
	}
}

func TestValidate_CaseInsensitive(t *testing.T) {
	m := newTest([]string{"fuck"})
	if _, hit := m.Validate("Fuck you"); !hit {
		t.Fatal("英文应大小写通吃")
	}
}

func TestValidate_Trim(t *testing.T) {
	m := newTest([]string{"赌博"})
	if _, hit := m.Validate("  赌博  "); !hit {
		t.Fatal("首尾空格应 trim 后命中")
	}
}

func TestValidate_SpaceBypassClosed(t *testing.T) {
	// 2026-08-17 normalize 修复：去全部空白（含全角空格），「赌 博」式绕过应被拦
	m := newTest([]string{"赌博"})
	for _, content := range []string{"赌 博", "赌  博", "赌\t博", "赌\n博", "赌　博"} {
		if word, hit := m.Validate(content); !hit || word != "赌博" {
			t.Fatalf("去空格后 %q 应命中 赌博（got word=%q hit=%v）", content, word, hit)
		}
	}
}

func TestValidate_FullWidthASCII(t *testing.T) {
	// 2026-08-17 normalize 修复：全角 ASCII 转半角，全角绕过应被拦
	m := newTest([]string{"fuck"})
	for _, content := range []string{"ｆｕｃｋ", "ｆＵＣｋ", "ＦｕＣＫ"} {
		if word, hit := m.Validate(content); !hit || word != "fuck" {
			t.Fatalf("全角转半角后 %q 应命中 fuck（got word=%q hit=%v）", content, word, hit)
		}
	}
}

func TestValidate_WhitespaceOnlyNoHit(t *testing.T) {
	m := newTest([]string{"赌博", "fuck"})
	for _, content := range []string{"", " ", "  \t  ", "　"} {
		if _, hit := m.Validate(content); hit {
			t.Fatalf("纯空白内容 %q 不应命中", content)
		}
	}
}

func TestValidate_SpaceJoinTradeoff(t *testing.T) {
	// 已知取舍（2026-08-17 评审拍板）：去空格会拼出词边界，属接受误伤。
	// 「颜色 情调」规范化后含「色情」→ 会命中；硬拦截优先，软拦截送审为 Phase 2。
	m := newTest([]string{"色情"})
	if word, hit := m.Validate("颜色 情调"); !hit || word != "色情" {
		t.Fatalf("已接受的取舍：去空格后「颜色 情调」应命中 色情（got word=%q hit=%v）", word, hit)
	}
}

func TestValidate_EmptyWordIgnored(t *testing.T) {
	m := newTest([]string{"", "   ", "毒品"})
	if _, hit := m.Validate("  毒品 "); !hit {
		t.Fatal("空词应被忽略，正常词仍生效")
	}
}

func TestReload(t *testing.T) {
	m := newTest([]string{"赌博"})
	m.Reload([]string{"诈骗"})
	if _, hit := m.Validate("赌博网站"); hit {
		t.Fatal("Reload 后旧词应失效")
	}
	if _, hit := m.Validate("兼职诈骗"); !hit {
		t.Fatal("Reload 后新词应生效")
	}
}

func TestMask_NoHitUnchanged(t *testing.T) {
	m := newTest([]string{"赌博"})
	for _, c := range []string{"今天去爬山", "123", ""} {
		if got := m.Mask(c); got != c {
			t.Fatalf("未命中应原样返回，got %q want %q", got, c)
		}
	}
}

func TestMask_EqualLength(t *testing.T) {
	m := newTest([]string{"赌博", "毒品"})
	cases := []struct{ in, want string }{
		{"这个赌博网站真恶心", "这个**网站真恶心"},
		{"贩卖毒品违法", "贩卖**违法"},
		{"在线赌博赌博", "在线****"}, // 两处命中各自等长
	}
	for _, tc := range cases {
		if got := m.Mask(tc.in); got != tc.want {
			t.Fatalf("Mask(%q)=%q want %q", tc.in, got, tc.want)
		}
	}
}

func TestMask_SpaceInsideWordCovered(t *testing.T) {
	// 「赌 博」normalize 后命中"赌博"，原文跨度含空格 → 空格也打星，保持等长
	m := newTest([]string{"赌博"})
	if got, want := m.Mask("赌 博网站"), "***网站"; got != want {
		t.Fatalf("Mask= %q want %q", got, want)
	}
}

func TestMask_FullWidth(t *testing.T) {
	// 全角 ASCII 命中后，原文全角字符被等长打星
	m := newTest([]string{"fuck"})
	if got, want := m.Mask("ｆｕｃｋ you"), "**** you"; got != want {
		t.Fatalf("Mask= %q want %q", got, want)
	}
}

func TestFindWords_UniqueOrdered(t *testing.T) {
	m := newTest([]string{"赌博", "毒品"})
	got := m.FindWords("贩卖毒品和赌博，毒品害人")
	want := []string{"毒品", "赌博"} // 首次出现顺序，去重
	if len(got) != len(want) {
		t.Fatalf("FindWords= %v want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("FindWords= %v want %v", got, want)
		}
	}
	if got := m.FindWords("正常内容"); len(got) != 0 {
		t.Fatalf("未命中应返回空，got %v", got)
	}
}

func TestIsSensitive(t *testing.T) {
	m := newTest([]string{"赌博"})
	if !m.IsSensitive("在线赌博", "正常") {
		t.Fatal("任一字段命中即 true")
	}
	if m.IsSensitive("爬山", "吃火锅") {
		t.Fatal("全不命中应 false")
	}
	if m.IsSensitive() {
		t.Fatal("空参应 false")
	}
}
