package home

import (
	"encoding/base64"
	"encoding/json"
	"time"
)

// FeedCursor Feed 游标分页的游标载体（keyset pagination）。
//
// 定位方式：游标记录"上一页最后一条"的完整排序键，下一页用
// 「排序键 < 游标」的谓词直接定位，不再依赖 offset——代价不随页码增长，
// 且翻页期间有新内容插入也不会重复/漏展示。
//
// 各字段都是必需的：排序键字段用于定位，Mode/Category 用于拦截跨模式/跨分类复用，
// IsNote 用于跨表并列时的最终 tiebreak（见 sortFeedEntries 的说明）。
type FeedCursor struct {
	Mode     string    `json:"m"`           // 归一化排序模式：推荐/最新/最热
	Category int       `json:"c"`           // 分类，防止换分类后复用旧游标
	Size     int       `json:"s,omitempty"` // 每页条数；翻页时不传 size 时沿用（0=旧版本游标，回退请求参数/默认值）
	Page     int       `json:"pg"`          // 递增页码，仅用于回填兼容字段 page
	IsNote   bool      `json:"n"`           // 边界元素来自笔记(true)还是攻略(false)
	ID       uint      `json:"i"`  // 边界元素 id（所有模式）
	Publish  time.Time `json:"p"`  // 边界元素发布时间（推荐/最新）
	Hot      int       `json:"g"`  // 边界元素热度 heat（最热）
	Like     int64     `json:"l"`  // 边界元素点赞数 like_count（最热）
	IsHot    int       `json:"o"`  // 边界元素热门标记 hot_flag（推荐：0=热门置顶段，1=普通段）
}

// FeedKeysetQuery 游标分页的仓储查询参数。
type FeedKeysetQuery struct {
	Mode     string      // 归一化排序模式：推荐/最新/最热
	Category int         // 分类：0=全部
	Limit    int         // 本表最多取多少条（调用方传 size+1）
	HotIDs   []uint      // 推荐模式：本表带热门标签的内容ID（为空表示本表无热门内容）
	Cursor   *FeedCursor // nil=第一页
}

// EncodeFeedCursor 游标 → URL 安全字符串（RawURLEncoding 避免 query 中出现 +//=）
func EncodeFeedCursor(c *FeedCursor) string {
	data, err := json.Marshal(c)
	if err != nil {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(data)
}

// DecodeFeedCursor 字符串 → 游标；解析失败返回 false。
// 调用方应返回参数错误而不是静默退回第一页——静默兜底会掩盖客户端的 bug，
// 表现成"翻页翻着翻着又回到开头"，比直接报错更难排查。
func DecodeFeedCursor(s string) (*FeedCursor, bool) {
	data, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return nil, false
	}
	var c FeedCursor
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, false
	}
	// id 自增从 1 开始，0 一定不是合法边界元素（能拦住空对象等构造出来的游标）
	if c.ID == 0 {
		return nil, false
	}
	// 按模式校验排序键：缺字段时 keyset 谓词会静默命中 0 行，表现成「已经翻到底」，
	// 比直接报参数错误更难排查。各字段的必需性见 FeedCursor 注释。
	switch c.Mode {
	case "最新":
		if c.Publish.IsZero() {
			return nil, false
		}
	case "推荐":
		if c.Publish.IsZero() || (c.IsHot != 0 && c.IsHot != 1) {
			return nil, false
		}
	case "最热":
		if c.Hot < 0 || c.Like < 0 {
			return nil, false
		}
	default:
		return nil, false
	}
	// Size=0 兼容本次改动前签发的旧游标；非 0 时必须落在接口允许的区间内
	if c.Size < 0 || c.Size > 100 {
		return nil, false
	}
	return &c, true
}

// feedKeysetPredicate 构造 keyset 谓词：返回"排在游标之后"的全部行。Cursor 为 nil 时无谓词（第一页）。
//
// isStrategy 决定跨表并列时的 tiebreak 方向。全局序最后一级是「笔记在攻略前」，
// 所以仅当游标元素是笔记、且当前查的是攻略表时，排序键与之完全相同的攻略行必须一并放行，
// 否则那条攻略会被漏掉（它在全局序中排在游标之后，但单表谓词判它"不大于游标"）。
// 反向不需要：游标是攻略时，同键的笔记排在前面，本就不该进入下一页。
func feedKeysetPredicate(q FeedKeysetQuery, isStrategy bool) (string, []interface{}) {
	c := q.Cursor
	if c == nil {
		return "", nil
	}
	tiebreak := isStrategy && c.IsNote

	switch q.Mode {
	case "最热":
		return buildKeyset([]string{"heat", "like_count", "id"},
			[]interface{}{c.Hot, c.Like, c.ID}, tiebreak)

	case "推荐":
		// hot_flag：0=热门置顶段，1=普通段。hotIDs 为空时本表无热门内容，hot_flag 恒为 1，
		// 用常量 1 代替 CASE（不要直接退化成「最新」谓词——游标在置顶段时，
		// 普通段的全部行都应放行，退化成时间序会漏掉它们）。
		hotExpr := "1"
		if len(q.HotIDs) > 0 {
			hotExpr = "CASE WHEN id IN (" + joinUintIDs(q.HotIDs) + ") THEN 0 ELSE 1 END"
		}
		pubPred, pubArgs := buildPublishKeyset(c, tiebreak)
		args := append([]interface{}{c.IsHot, c.IsHot}, pubArgs...)
		return "(" + hotExpr + " > ? OR (" + hotExpr + " = ? AND " + pubPred + "))", args

	default: // 最新
		pubPred, pubArgs := buildPublishKeyset(c, tiebreak)
		return pubPred, pubArgs
	}
}

// buildPublishKeyset 时间序谓词：排在游标 (publish_time, id) 之后的行；tiebreak 为真时额外放行完全并列的行
func buildPublishKeyset(c *FeedCursor, tiebreak bool) (string, []interface{}) {
	return buildKeyset([]string{"publish_time", "id"},
		[]interface{}{c.Publish, c.ID}, tiebreak)
}

// buildKeyset 把「元组比较 (k1,...,kn) < (v1,...,vn)」展开成 OR 形式。
//
// 为什么不直接写元组比较：MySQL 8.0 的优化器不会把它下推成索引范围。实测
// `(publish_time, id) < (?, ?)` 的 plan 是 type=index——从索引尾部反向全扫、
// 逐行过上层 Filter（58 行）；展开形式是 type=range，直接跳到范围起点（5 行）。
// 两者语义完全等价：
//
//	(a, b) < (A, B)  ⇔  a < A OR (a = A AND b < B)
//
// 这个差别正是游标分页的全部意义：元组形式在深翻页时退化成 O(N)（游标越靠前，
// 要走的索引越长），展开形式才是 O(log N + limit)。
//
// tiebreak=true 时最后一级用 <=，用于放行「与游标完全并列」的跨表行（见 feedKeysetPredicate）。
func buildKeyset(keys []string, vals []interface{}, tiebreak bool) (string, []interface{}) {
	var build func(i int) (string, []interface{})
	build = func(i int) (string, []interface{}) {
		if i == len(keys)-1 {
			cmp := "<"
			if tiebreak {
				cmp = "<="
			}
			return keys[i] + " " + cmp + " ?", []interface{}{vals[i]}
		}
		inner, innerArgs := build(i + 1)
		return keys[i] + " < ? OR (" + keys[i] + " = ? AND (" + inner + "))",
			append([]interface{}{vals[i], vals[i]}, innerArgs...)
	}
	pred, args := build(0)
	return "(" + pred + ")", args
}

// feedKeysetOrder 各模式的 ORDER BY。单表内 id 唯一，所以 (排序键..., id) 已是严格全序，
// 无需在 SQL 里再带 isNote（它是表内恒等项，只在内存归并和跨表谓词里有意义）。
func feedKeysetOrder(q FeedKeysetQuery) string {
	switch q.Mode {
	case "最热":
		return "heat DESC, like_count DESC, id DESC"
	case "推荐":
		if len(q.HotIDs) > 0 {
			return "CASE WHEN id IN (" + joinUintIDs(q.HotIDs) + ") THEN 0 ELSE 1 END ASC, publish_time DESC, id DESC"
		}
		return "publish_time DESC, id DESC"
	default: // 最新
		return "publish_time DESC, id DESC"
	}
}
