package sensitive

import (
	"strings"
	"sync"

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

// normalize 词/内容统一规范化：TrimSpace + ToLower（英文大小写通吃）
func normalize(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}
