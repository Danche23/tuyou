package sensitive

import (
	"strings"
	"sync"
	"unicode"

	"gorm.io/gorm"

	"tuyou/internal/model/entity"
)

// node trie 前缀树节点
type node struct {
	children map[rune]*node // 子节点
	isEnd    bool           // 是否是某个词的结尾
	word     string         // isEnd 时记录完整词，便于返回命中的词
}

// SensitiveMatcher 敏感词匹配器（启动构建，请求期只读；Reload 热更新时用锁保护）
type SensitiveMatcher struct {
	root *node
	mu   sync.RWMutex
}

// NewFromDB 从 sensitive_word 表加载启用词构建 trie（启动时调用一次）
func NewFromDB(db *gorm.DB) (*SensitiveMatcher, error) {
	var words []string
	err := db.Model(&entity.SensitiveWord{}).
		Where("status = ?", 1).
		Pluck("word", &words).Error
	if err != nil {
		return nil, err
	}
	return New(words), nil
}

// New 用给定词表构建（测试用，不碰 DB）
func New(words []string) *SensitiveMatcher {
	m := &SensitiveMatcher{root: &node{children: make(map[rune]*node)}}
	for _, w := range words {
		m.insert(w)
	}
	return m
}

// insert 插入单个词（私有）
func (m *SensitiveMatcher) insert(word string) {
	word = normalize(word)
	if word == "" {
		return
	}
	cur := m.root
	for _, ch := range word {
		if cur.children[ch] == nil {
			cur.children[ch] = &node{children: make(map[rune]*node)}
		}
		cur = cur.children[ch]
	}
	cur.isEnd = true
	cur.word = word
}

// Validate 校验内容：返回 (命中的词, 是否命中)；未命中 ("", false)
func (m *SensitiveMatcher) Validate(content string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	runes := []rune(normalize(content))
	for i := 0; i < len(runes); i++ {
		cur := m.root
		for j := i; j < len(runes); j++ {
			next, ok := cur.children[runes[j]]
			if !ok {
				break
			}
			cur = next
			if cur.isEnd {
				return cur.word, true
			}
		}
	}
	return "", false
}

// Reload 用新词表重建 trie（Phase 2 后台增删词后热更新）
func (m *SensitiveMatcher) Reload(words []string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.root = &node{children: make(map[rune]*node)}
	for _, w := range words {
		m.insert(w)
	}
}

// normalize 词/内容统一规范化：
//   - ToLower（英文大小写通吃）
//   - 去掉全部空白（含全角空格 U+3000），关闭「赌 博」式空格绕过
//   - 全角 ASCII 转半角（ｆｕｃｋ → fuck），关闭全角绕过
// 插入词与校验内容共用，保证二者落在同一规范化空间。
//
// 已知取舍（2026-08-17 评审拍板，硬拦截优先）：去空格会把无辜短语拼成命中词，
// 如「颜色 情调」→「颜色情调」含「色情」会被误拦。缓解：命中后重跑保留空格的匹配、
// 软拦截送人工审核 → 列为 Phase 2 待办。
func normalize(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		switch {
		case unicode.IsSpace(r):
			continue
		case r >= 0xFF01 && r <= 0xFF5E: // 全角 ASCII → 半角
			b.WriteRune(r - 0xFEE0)
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// normalizeWithIndex 规范化内容并记录每个规范化后 rune 对应原文 rune 的下标。
// 用 per-rune unicode.ToLower（而非整串 strings.ToLower），保证 rune 一一对应不串位。
func normalizeWithIndex(s string) (runes []rune, origIndex []int) {
	for i, r := range []rune(s) {
		r = unicode.ToLower(r)
		switch {
		case unicode.IsSpace(r):
			continue
		case r >= 0xFF01 && r <= 0xFF5E:
			r -= 0xFEE0
		}
		runes = append(runes, r)
		origIndex = append(origIndex, i)
	}
	return runes, origIndex
}

// findSpans 在规范化 rune 上找所有命中词的 [start,end) 区间（可重叠，如「赌博网」含「赌」「赌博」）
func (m *SensitiveMatcher) findSpans(runes []rune) [][2]int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var spans [][2]int
	for i := 0; i < len(runes); i++ {
		cur := m.root
		for j := i; j < len(runes); j++ {
			next, ok := cur.children[runes[j]]
			if !ok {
				break
			}
			cur = next
			if cur.isEnd {
				spans = append(spans, [2]int{i, j + 1})
			}
		}
	}
	return spans
}

// Mask 命中词等长打星：把原文中"命中词覆盖的连续区间"整体替换成等长 *。
// 区间含规范化时被剥掉的内部空白/全角字符，保证输出长度 = 原文长度（设计 §8.2）。
func (m *SensitiveMatcher) Mask(content string) string {
	if content == "" {
		return content
	}
	nr, origIndex := normalizeWithIndex(content)
	spans := m.findSpans(nr)
	if len(spans) == 0 {
		return content
	}
	orig := []rune(content)
	mask := make([]bool, len(orig))
	for _, sp := range spans {
		lo, hi := origIndex[sp[0]], origIndex[sp[1]-1]
		for k := lo; k <= hi; k++ {
			mask[k] = true
		}
	}
	var b strings.Builder
	for i, r := range orig {
		if mask[i] {
			b.WriteRune('*')
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// FindWords 返回内容命中的敏感词（normalize 后词形），按首次出现顺序去重。
func (m *SensitiveMatcher) FindWords(content string) []string {
	nr, _ := normalizeWithIndex(content)
	m.mu.RLock()
	defer m.mu.RUnlock()
	seen := make(map[string]bool)
	var words []string
	for i := 0; i < len(nr); i++ {
		cur := m.root
		for j := i; j < len(nr); j++ {
			next, ok := cur.children[nr[j]]
			if !ok {
				break
			}
			cur = next
			if cur.isEnd && !seen[cur.word] {
				seen[cur.word] = true
				words = append(words, cur.word)
			}
		}
	}
	return words
}

// IsSensitive 判断多个文本中任一命中敏感词（供发布接口打 is_sensitive 快照）
func (m *SensitiveMatcher) IsSensitive(texts ...string) bool {
	for _, t := range texts {
		if len(m.FindWords(t)) > 0 {
			return true
		}
	}
	return false
}
