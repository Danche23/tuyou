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
