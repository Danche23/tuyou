# 发布笔记（POST /api/v1/notes）实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 实现发布笔记接口，用户通过 JWT 认证后创建笔记（含标签关联、图片写入、发布计数）。

**Architecture:** 严格遵循项目分层：DTO → Repository → Service → Controller → Router Assembly。Service 层管理事务边界，Repository 层只做数据访问，Controller 只做参数绑定和响应。

**Tech Stack:** Go 1.24, Gin, GORM, 现有 pkg/errors, pkg/response, middleware.Auth()

## Global Constraints

- 所有 API 返回统一响应格式 `{code: int, message: string, data: any}`，HTTP 状态码始终 200
- JWT 认证使用现有 `middleware.Auth()`（已 set `user_id` 到 context）
- Repository 写操作接受 `tx *gorm.DB`，Service 层管理事务
- 接口与实现分离：`xxx_interface.go` + `xxx_repository.go` / `xxx_service.go`
- 错误使用 `errors.New(code, message)` 创建，Controller 用 `response.BizError(ctx, err)` 返回
- Category 枚举：1=徒步, 2=露营, 3=骑行, 4=登山, 5=跑步, 6=攀岩, 7=溯溪
- 图片/视频互斥：有 videoUrl → images 恰好 1 张（第一帧），无 videoUrl → images 1~5 张
- 封面 = images[0]

---

### Task 1: 创建请求 DTO

**Files:**
- Create: `internal/model/dto/request/note.go`

**Interfaces:**
- Consumes: `entity.Note` (现有的 note entity)
- Produces: `request.CreateNoteRequest` 结构体

- [ ] **Step 1: 写入请求 DTO**

```go
package request

// CreateNoteRequest 发布笔记请求
type CreateNoteRequest struct {
	Title    string   `json:"title"    binding:"required,max=50"`
	Content  string   `json:"content"  binding:"required,max=2000"`
	Category int      `json:"category" binding:"required,min=1,max=7"`
	Tags     []string `json:"tags"     binding:"required,min=1,max=3"`
	Images   []string `json:"images"   binding:"required,min=1,max=5"`
	VideoURL string   `json:"videoUrl" binding:"omitempty"`
}
```

- [ ] **Step 2: 编译验证**

```bash
cd D:/tuyou && go build ./internal/model/dto/request/
```

- [ ] **Step 3: 提交**

```bash
git add internal/model/dto/request/note.go
git commit -m "feat: 添加发布笔记请求 DTO"
```

---

### Task 2: 创建响应 DTO

**Files:**
- Create: `internal/model/dto/response/note.go`

**Interfaces:**
- Consumes: `entity.Note` (现有的)
- Produces: `response.AuthorInfo`, `response.NoteResponse`, `response.ToNoteResponse()`

- [ ] **Step 1: 写入响应 DTO**

```go
package response

import (
	"tuyou/internal/model/entity"
)

// AuthorInfo 作者简要信息
type AuthorInfo struct {
	ID       uint   `json:"id"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

// NoteResponse 笔记响应
type NoteResponse struct {
	ID           uint        `json:"id"`
	Type         string      `json:"type"`
	Author       *AuthorInfo `json:"author"`
	Title        string      `json:"title"`
	CoverImg     string      `json:"coverImg"`
	Images       []string    `json:"images"`
	VideoURL     string      `json:"videoUrl"`
	Content      string      `json:"content"`
	ViewCount    int64       `json:"viewCount"`
	LikeCount    int64       `json:"likeCount"`
	CollectCount int64       `json:"collectCount"`
	CommentCount int64       `json:"commentCount"`
	IsLiked      bool        `json:"isLiked"`
	IsFavorited  bool        `json:"isFavorited"`
	Category     int         `json:"category"`
	Tags         []string    `json:"tags"`
	Status       int         `json:"status"`
	PublishTime  string      `json:"publishTime"`
}

// ToNoteResponse 组装 NoteResponse
func ToNoteResponse(note *entity.Note, author *AuthorInfo, tags []string, images []string) *NoteResponse {
	coverImg := ""
	if len(images) > 0 {
		coverImg = images[0]
	}
	return &NoteResponse{
		ID:           note.ID,
		Type:         "note",
		Author:       author,
		Title:        note.Title,
		CoverImg:     coverImg,
		Images:       images,
		VideoURL:     note.VideoURL,
		Content:      note.Content,
		ViewCount:    note.ViewCount,
		LikeCount:    note.LikeCount,
		CollectCount: note.CollectCount,
		CommentCount: note.CommentCount,
		IsLiked:      false,
		IsFavorited:  false,
		Category:     note.Category,
		Tags:         tags,
		Status:       note.Status,
		PublishTime:  note.PublishTime.Format("2006-01-02 15:04:05"),
	}
}
```

- [ ] **Step 2: 编译验证**

```bash
cd D:/tuyou && go build ./internal/model/dto/response/
```

- [ ] **Step 3: 提交**

```bash
git add internal/model/dto/response/note.go
git commit -m "feat: 添加笔记响应 DTO"
```

---

### Task 3: 创建 NoteRepository 接口

**Files:**
- Create: `internal/repository/note_interface.go`

**Interfaces:**
- Consumes: `entity.Note`, `entity.Tag`, `entity.NoteTag`, `entity.NoteImage`, `entity.Member`, `response.AuthorInfo`
- Produces: `repository.NoteRepository` 接口

- [ ] **Step 1: 写入接口**

```go
package repository

import (
	"tuyou/internal/model/dto/response"
	"tuyou/internal/model/entity"

	"gorm.io/gorm"
)

// NoteRepository 笔记仓储接口
type NoteRepository interface {
	// Create 创建笔记
	Create(tx *gorm.DB, note *entity.Note) error

	// FindOrCreateTags 根据标签名查找或创建标签
	FindOrCreateTags(tx *gorm.DB, tagNames []string) ([]entity.Tag, error)

	// CreateNoteTags 批量创建笔记-标签关联
	CreateNoteTags(tx *gorm.DB, noteID uint, tagIDs []uint) error

	// CreateNoteImages 批量创建笔记图片
	CreateNoteImages(tx *gorm.DB, noteID uint, imageURLs []string) error

	// GetAuthor 获取作者简要信息
	GetAuthor(memberID uint) (*response.AuthorInfo, error)

	// IncrPublishCount 发布计数 +1
	IncrPublishCount(tx *gorm.DB, memberID uint) error
}
```

- [ ] **Step 2: 编译验证**

```bash
cd D:/tuyou && go build ./internal/repository/
```

- [ ] **Step 3: 提交**

```bash
git add internal/repository/note_interface.go
git commit -m "feat: 添加 NoteRepository 接口"
```

---

### Task 4: 实现 NoteRepository

**Files:**
- Create: `internal/repository/note_repository.go`

**Interfaces:**
- Consumes: `NoteRepository` 接口（Task 3）、`entity.Note`, `entity.Tag`, `entity.NoteTag`, `entity.NoteImage`, `entity.Member`
- Produces: `repository.NewNoteRepository(db *gorm.DB) NoteRepository`

- [ ] **Step 1: 写入实现**

```go
package repository

import (
	"tuyou/internal/model/dto/response"
	"tuyou/internal/model/entity"

	"gorm.io/gorm"
)

type noteRepository struct {
	db *gorm.DB
}

// NewNoteRepository 创建笔记仓储实例
func NewNoteRepository(db *gorm.DB) NoteRepository {
	return &noteRepository{db: db}
}

// Create 创建笔记
func (r *noteRepository) Create(tx *gorm.DB, note *entity.Note) error {
	return tx.Create(note).Error
}

// FindOrCreateTags 根据标签名查找或创建标签
func (r *noteRepository) FindOrCreateTags(tx *gorm.DB, tagNames []string) ([]entity.Tag, error) {
	tags := make([]entity.Tag, 0, len(tagNames))
	for _, name := range tagNames {
		var tag entity.Tag
		err := tx.Where("tag_name = ?", name).First(&tag).Error
		if err == gorm.ErrRecordNotFound {
			tag = entity.Tag{TagName: name, Status: 1}
			if err := tx.Create(&tag).Error; err != nil {
				return nil, err
			}
		} else if err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}
	return tags, nil
}

// CreateNoteTags 批量创建笔记-标签关联
func (r *noteRepository) CreateNoteTags(tx *gorm.DB, noteID uint, tagIDs []uint) error {
	noteTags := make([]entity.NoteTag, 0, len(tagIDs))
	for _, tagID := range tagIDs {
		noteTags = append(noteTags, entity.NoteTag{
			NoteID: noteID,
			TagID:  tagID,
		})
	}
	return tx.Create(&noteTags).Error
}

// CreateNoteImages 批量创建笔记图片
func (r *noteRepository) CreateNoteImages(tx *gorm.DB, noteID uint, imageURLs []string) error {
	images := make([]entity.NoteImage, 0, len(imageURLs))
	for i, url := range imageURLs {
		images = append(images, entity.NoteImage{
			NoteID: noteID,
			URL:    url,
			Sort:   i, // 0=封面, 1=第二张...
		})
	}
	return tx.Create(&images).Error
}

// GetAuthor 获取作者简要信息
func (r *noteRepository) GetAuthor(memberID uint) (*response.AuthorInfo, error) {
	var member entity.Member
	if err := r.db.First(&member, memberID).Error; err != nil {
		return nil, err
	}
	return &response.AuthorInfo{
		ID:       member.ID,
		Nickname: member.Nickname,
		Avatar:   member.Avatar,
	}, nil
}

// IncrPublishCount 发布计数 +1
func (r *noteRepository) IncrPublishCount(tx *gorm.DB, memberID uint) error {
	return tx.Model(&entity.Member{}).Where("id = ?", memberID).
		UpdateColumn("publish_count", gorm.Expr("publish_count + 1")).Error
}
```

- [ ] **Step 2: 编译验证**

```bash
cd D:/tuyou && go build ./internal/repository/
```

- [ ] **Step 3: 提交**

```bash
git add internal/repository/note_repository.go
git commit -m "feat: 实现 NoteRepository"
```

---

### Task 5: 创建 NoteService 接口

**Files:**
- Create: `internal/service/note_interface.go`

**Interfaces:**
- Consumes: `request.CreateNoteRequest`, `response.NoteResponse`
- Produces: `service.NoteService` 接口

- [ ] **Step 1: 写入接口**

```go
package service

import (
	"tuyou/internal/model/dto/request"
	"tuyou/internal/model/dto/response"
)

// NoteService 笔记服务接口
type NoteService interface {
	// Create 发布笔记
	Create(memberID uint, req *request.CreateNoteRequest) (*response.NoteResponse, error)
}
```

- [ ] **Step 2: 编译验证**

```bash
cd D:/tuyou && go build ./internal/service/
```

- [ ] **Step 3: 提交**

```bash
git add internal/service/note_interface.go
git commit -m "feat: 添加 NoteService 接口"
```

---

### Task 6: 实现 NoteService

**Files:**
- Create: `internal/service/note_service.go`

**Interfaces:**
- Consumes: `NoteRepository` (Task 3-4), `NoteService` (Task 5), `request.CreateNoteRequest`, `response.NoteResponse`
- 依赖: `entity.Note`, `pkg/errors`
- Produces: `service.NewNoteService(repo NoteRepository) NoteService`

- [ ] **Step 1: 写入实现**

```go
package service

import (
	"time"

	"tuyou/internal/model/dto/request"
	"tuyou/internal/model/dto/response"
	"tuyou/internal/model/entity"
	"tuyou/internal/repository"
	"tuyou/pkg/errors"

	"gorm.io/gorm"
)

type noteService struct {
	noteRepo repository.NoteRepository
	db       *gorm.DB
}

// NewNoteService 创建笔记服务实例
func NewNoteService(noteRepo repository.NoteRepository, db *gorm.DB) NoteService {
	return &noteService{noteRepo: noteRepo, db: db}
}

// Create 发布笔记
func (s *noteService) Create(memberID uint, req *request.CreateNoteRequest) (*response.NoteResponse, error) {
	// 校验图片/视频互斥
	if req.VideoURL != "" && len(req.Images) != 1 {
		return nil, errors.New(errors.CodeInvalidParam, "视频模式下只能上传1张图片（第一帧截图）")
	}

	// 构建 Note 实体
	note := &entity.Note{
		MemberID:    memberID,
		Title:       req.Title,
		Content:     req.Content,
		Category:    req.Category,
		VideoURL:    req.VideoURL,
		Status:      1,
		PublishTime: time.Now(),
	}

	// 开启事务
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 1. 创建笔记
		if err := s.noteRepo.Create(tx, note); err != nil {
			return err
		}

		// 2. 处理标签
		tags, err := s.noteRepo.FindOrCreateTags(tx, req.Tags)
		if err != nil {
			return err
		}
		tagIDs := make([]uint, len(tags))
		for i, t := range tags {
			tagIDs[i] = t.ID
		}
		if err := s.noteRepo.CreateNoteTags(tx, note.ID, tagIDs); err != nil {
			return err
		}

		// 3. 创建图片
		if err := s.noteRepo.CreateNoteImages(tx, note.ID, req.Images); err != nil {
			return err
		}

		// 4. 更新发布计数
		if err := s.noteRepo.IncrPublishCount(tx, memberID); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "发布笔记失败", err)
	}

	// 组装响应（事务外，避免长事务）
	author, err := s.noteRepo.GetAuthor(memberID)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "获取作者信息失败", err)
	}

	return response.ToNoteResponse(note, author, req.Tags, req.Images), nil
}
```

- [ ] **Step 2: 编译验证**

```bash
cd D:/tuyou && go build ./internal/service/
```

- [ ] **Step 3: 提交**

```bash
git add internal/service/note_service.go
git commit -m "feat: 实现发布笔记 Service"
```

---

### Task 7: 创建 Controller

**Files:**
- Create: `internal/api/v1/note/controller.go`

**Interfaces:**
- Consumes: `NoteService` (Task 5-6), `middleware.GetUserID()`, `pkg/response`
- Produces: `note.NewController(noteSvc service.NoteService) *Controller`

- [ ] **Step 1: 创建目录并写入 Controller**

```bash
mkdir -p D:/tuyou/internal/api/v1/note
```

```go
package note

import (
	"net/http"

	"tuyou/internal/middleware"
	"tuyou/internal/model/dto/request"
	"tuyou/internal/service"
	"tuyou/pkg/response"

	"github.com/gin-gonic/gin"
)

// Controller 笔记控制器
type Controller struct {
	noteSvc service.NoteService
}

// NewController 创建控制器实例
func NewController(noteSvc service.NoteService) *Controller {
	return &Controller{noteSvc: noteSvc}
}

// Create 发布笔记
// POST /api/v1/notes
func (c *Controller) Create(ctx *gin.Context) {
	memberID := middleware.GetUserID(ctx)
	if memberID == 0 {
		response.Error(ctx, http.StatusUnauthorized, "未登录")
		return
	}

	var req request.CreateNoteRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "参数校验失败: "+err.Error())
		return
	}

	resp, err := c.noteSvc.Create(memberID, &req)
	if err != nil {
		response.BizError(ctx, err)
		return
	}

	response.Success(ctx, resp)
}
```

- [ ] **Step 2: 编译验证**

```bash
cd D:/tuyou && go build ./internal/api/v1/note/
```

- [ ] **Step 3: 提交**

```bash
git add internal/api/v1/note/controller.go
git commit -m "feat: 添加发布笔记 Controller"
```

---

### Task 8: 创建路由注册

**Files:**
- Create: `internal/api/v1/note/routes.go`

**Interfaces:**
- Consumes: `Controller` (Task 7), `middleware.Auth()`
- Produces: 路由注册

- [ ] **Step 1: 写入路由文件**

```go
package note

import (
	"tuyou/internal/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册笔记路由
func (c *Controller) RegisterRoutes(r *gin.RouterGroup) {
	notes := r.Group("/notes")
	notes.Use(middleware.Auth())
	{
		notes.POST("", c.Create)
	}
}
```

- [ ] **Step 2: 编译验证**

```bash
cd D:/tuyou && go build ./internal/api/v1/note/
```

- [ ] **Step 3: 提交**

```bash
git add internal/api/v1/note/routes.go
git commit -m "feat: 注册笔记路由"
```

---

### Task 9: 修改 Router 装配

**Files:**
- Modify: `internal/api/router.go`

**Interfaces:**
- Consumes: `note.Controller` (Task 7)
- Produces: 路由器支持 noteCtrl

- [ ] **Step 1: 读取现有文件**

- [ ] **Step 2: 修改 router.go**

原文：
```go
package api

import (
	"tuyou/internal/middleware"

	"github.com/gin-gonic/gin"
)

// Router 路由
type Router struct {
	// ========== 在这里添加你的 Controller 字段 ==========
	// userCtrl  *user.Controller
	// authCtrl  *auth.Controller
}

// NewRouter 创建路由
func NewRouter(
// ========== 在这里添加你的 Service 参数（依赖注入） ==========
// userService service.UserService,
// authService service.AuthService,
) *Router {
	return &Router{
		// userCtrl: user.NewController(userService),
		// authCtrl: auth.NewController(authService),
	}
}
```

替换为：
```go
package api

import (
	"tuyou/internal/api/v1/note"
	"tuyou/internal/middleware"

	"github.com/gin-gonic/gin"
)

// Router 路由
type Router struct {
	noteCtrl *note.Controller
	// ========== 在这里添加你的 Controller 字段 ==========
	// userCtrl  *user.Controller
	// authCtrl  *auth.Controller
}

// NewRouter 创建路由
func NewRouter(noteCtrl *note.Controller) *Router {
	return &Router{
		noteCtrl: noteCtrl,
	}
}
```

然后在 `Setup` 方法的 `apiGroup` 块内添加：

```go
r.noteCtrl.RegisterRoutes(apiGroup)
```

- [ ] **Step 3: 编译验证**

```bash
cd D:/tuyou && go build ./internal/api/
```

- [ ] **Step 4: 提交**

```bash
git add internal/api/router.go
git commit -m "feat: Router 装配 noteCtrl"
```

---

### Task 10: 修改 App 依赖注入

**Files:**
- Modify: `internal/app/app.go`

**Interfaces:**
- Consumes: `repository.NewNoteRepository`, `service.NewNoteService`, `note.NewController`, `api.NewRouter`
- Produces: 完整依赖注入链

- [ ] **Step 1: 读取现有文件**

- [ ] **Step 2: 修改 initDependencies()**

在 `initDependencies()` 中，替换注释块为实际代码：

```go
// initDependencies 初始化依赖注入
func (a *App) initDependencies() {
	// ========== 创建 Repository ==========
	noteRepo := repository.NewNoteRepository(a.mysqlDB)

	// ========== 创建 Service ==========
	noteSvc := service.NewNoteService(noteRepo, a.mysqlDB)

	// ========== 创建 Controller ==========
	noteCtrl := note.NewController(noteSvc)

	// ========== 创建 Router ==========
	a.router = api.NewRouter(noteCtrl)
}
```

需要在文件头部添加 import：
```go
import (
	// ... 现有 imports ...
	"tuyou/internal/api/v1/note"
	"tuyou/internal/repository"
	"tuyou/internal/service"
)
```

- [ ] **Step 3: 编译验证**

```bash
cd D:/tuyou && go build ./internal/app/
```

- [ ] **Step 4: 提交**

```bash
git add internal/app/app.go
git commit -m "feat: App 依赖注入笔记模块"
```

---

### Task 11: 编译验证

**Files:**
- 无新建

- [ ] **Step 1: 全量编译**

```bash
cd D:/tuyou && go build ./...
```

预期：编译通过，无错误。

- [ ] **Step 2: 编译二进制**

```bash
cd D:/tuyou && go build -o bin/tuyou-api.exe cmd/server/main.go
```

预期：生成 `bin/tuyou-api.exe`。

---

### Task 12: 启动验证

**Files:**
- 无新建

- [ ] **Step 1: 确认 MySQL 和 Redis 可用**

确保 `configs/config.yaml` 中数据库配置正确，MySQL 中有 `tuyou` 数据库。

- [ ] **Step 2: 启动服务**

```bash
cd D:/tuyou && go run cmd/server/main.go
```

预期日志：
```
=========================================
欢迎使用 tuyou-api
版本: 1.0.0
模式: debug
配置加载成功
=========================================
MySQL 连接成功 ...
数据库迁移完成
HTTP 服务器启动 addr=:9300 mode=debug
```

- [ ] **Step 3: 测试接口**

```bash
# 先注册/登录获取 token
curl -X POST http://localhost:9300/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"phone":"13800000001","nickname":"测试用户","password":"123456"}'

# 用返回的 token 发布笔记
curl -X POST http://localhost:9300/api/v1/notes \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "title": "测试笔记标题",
    "content": "这是一篇测试笔记的内容",
    "category": 1,
    "tags": ["徒步", "登山"],
    "images": ["https://example.com/img1.jpg", "https://example.com/img2.jpg"]
  }'
```

预期响应：
```json
{
  "code": 0,
  "message": "成功",
  "data": {
    "id": 1,
    "type": "note",
    "author": {"id": 1, "nickname": "测试用户", "avatar": ""},
    "title": "测试笔记标题",
    "coverImg": "https://example.com/img1.jpg",
    "images": ["https://example.com/img1.jpg", "https://example.com/img2.jpg"],
    ...
    "tags": ["徒步", "登山"],
    "category": 1,
    "status": 1,
    ...
  }
}
```

- [ ] **Step 4: 提交**

```bash
git add -A
git commit -m "chore: 发布笔记接口验证通过"
```
