# 笔记模块设计文档

> 日期：2026-07-25 | 开发者：李浩坤 | 分支：develop

---

## 一、概述

户外旅行平台前台笔记模块。用户发布笔记（图文/视频），支持查看详情、点赞 toggle、收藏 toggle、删除。

本轮**上传走 Mock**（只校验 URL 格式和数量），不接 MinIO。

---

## 二、API 接口（5 个）

全部需要 JWT 认证，路由前缀 `/api/v1`。

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /notes | 发布笔记 |
| GET | /notes/:id | 笔记详情（含 isLiked/isFavorited） |
| POST | /notes/:id/like | 点赞/取消点赞 (toggle) |
| POST | /notes/:id/favorite | 收藏/取消收藏 (toggle) |
| DELETE | /notes/:id | 删除笔记（作者/管理员） |

### 2.1 发布笔记

**请求体：**

```json
{
  "title":    "string, 必填, max=50字",
  "content":  "string, 必填, max=2000字",
  "category": "int, 必填, 1~7",
  "tags":     ["string数组, 必填, 1~3个"],
  "images":   ["string数组, 必填, 1~5个(视频模式固定1个)"],
  "videoUrl": "string, 可选, max=1个"
}
```

**图片/视频规则：**
- 选视频 → videoUrl 有值，images 恰好 1 张（前端截第一帧）
- 选图片 → videoUrl 空，images 1~5 张
- 封面永远是 images[0]

**Category 枚举：** 1=徒步, 2=露营, 3=骑行, 4=登山, 5=跑步, 6=攀岩, 7=溯溪

**响应：** data 为 NoteResponse 结构

### 2.2 笔记详情

**响应：** data 为 NoteResponse 结构（含 isLiked + isFavorited）

浏览量异步 +1（goroutine，不阻塞响应）

### 2.3 点赞/取消

**响应：**
```json
{
  "code": 0,
  "data": { "isLiked": true, "likeCount": 236 }
}
```

### 2.4 收藏/取消

**响应：**
```json
{
  "code": 0,
  "data": { "isFavorited": true }
}
```

### 2.5 删除笔记

| 操作者 | 行为 |
|--------|------|
| 作者 | status=0 + DeletedAt=now + 级联软删关联表（note_image/note_tag/note_like/note_collect） |
| 管理员 | 物理删除 note + 级联物理删所有关联表 |

响应统一格式 `{code: 0, message: "成功"}`

---

## 三、响应结构

### NoteResponse（笔记统一响应）

```json
{
  "id": 1,
  "type": "note",
  "author": { "id": 1, "nickname": "山林旅人", "avatar": "https://..." },
  "title": "周末骑行千岛湖环线",
  "coverImg": "https://...",
  "images": ["https://..."],
  "videoUrl": "https://...",
  "content": "正文内容...",
  "viewCount": 1024,
  "likeCount": 235,
  "collectCount": 88,
  "commentCount": 12,
  "isLiked": false,
  "isFavorited": false,
  "category": 3,
  "tags": ["千岛湖", "环湖"],
  "status": 1,
  "publishTime": "2026-04-18T12:00:00"
}
```

---

## 四、文件结构

### 新建文件（8 个）

```
internal/model/dto/request/note.go         # CreateNoteRequest
internal/model/dto/response/note.go        # AuthorInfo / NoteResponse / LikeResponse / FavoriteResponse / ToXxxResponse
internal/repository/note_interface.go      # NoteRepository 接口
internal/repository/note_repository.go     # NoteRepository 实现
internal/service/note_interface.go         # NoteService 接口
internal/service/note_service.go           # NoteService 实现
internal/api/v1/note/controller.go         # 5 个 Handler
internal/api/v1/note/routes.go             # RegisterRoutes
```

### 修改文件（2 个）

```
internal/api/router.go     # 添加 noteCtrl 字段 + 注册路由
internal/app/app.go        # 依赖注入链
```

---

## 五、分层设计

### DTO

**请求 DTO：**

```go
type CreateNoteRequest struct {
    Title    string   `json:"title"    binding:"required,max=50"`
    Content  string   `json:"content"  binding:"required,max=2000"`
    Category int      `json:"category" binding:"required,min=1,max=7"`
    Tags     []string `json:"tags"     binding:"required,min=1,max=3"`
    Images   []string `json:"images"   binding:"required,min=1,max=5"`
    VideoURL string   `json:"videoUrl" binding:"omitempty"`
}
```

**响应 DTO：**

- `AuthorInfo`: ID + Nickname + Avatar（查 member 表组装）
- `NoteResponse`: 完整 Post 结构，含 isLiked/isFavorited
- `LikeResponse`: isLiked + likeCount
- `FavoriteResponse`: isFavorited
- `ToNoteResponse(note, author, tags, images, isLiked, isFav)`: 组装函数

### Repository

**接口方法：**

| 分类 | 方法 | 说明 |
|------|------|------|
| CRUD | Create | 插入 note |
| CRUD | GetByID | 查单篇（DeletedAt IS NULL） |
| CRUD | SoftDelete | 作者删除：status=0, DeletedAt |
| CRUD | PhysicalDelete | 管理员删除：Unscoped().Delete |
| 浏览量 | IncrViewCount | view_count +1 |
| 标签 | FindOrCreateTags | 按 tagName 查找或创建 |
| 标签 | CreateNoteTags | 批量插入 note_tag |
| 标签 | DeleteNoteTags | 级联软删 note_tag |
| 标签 | GetNoteTags | 查标签名列表（JOIN tag） |
| 图片 | CreateNoteImages | 批量插入 note_image |
| 图片 | SoftDeleteNoteImages | 级联软删 |
| 图片 | GetNoteImages | 查图片 URL 列表 |
| 点赞 | GetLike | 查单条 note_like |
| 点赞 | ToggleLike | 有则真删, 无则创建 |
| 点赞 | SoftDeleteNoteLikes | 级联软删 |
| 收藏 | GetCollect | 查单条 note_collect |
| 收藏 | ToggleCollect | 有则真删, 无则创建 |
| 收藏 | SoftDeleteNoteCollects | 级联软删 |
| 作者 | GetAuthor | 查 member：id, nickname, avatar |

**写操作均接受 `tx *gorm.DB`，由 Service 管理事务。**

### Service

**接口方法：**

```go
type NoteService interface {
    Create(memberID uint, req *CreateNoteRequest) (*NoteResponse, error)
    GetByID(noteID, currentMemberID uint) (*NoteResponse, error)
    ToggleLike(noteID, memberID uint) (*LikeResponse, error)
    ToggleFavorite(noteID, memberID uint) (*FavoriteResponse, error)
    Delete(noteID, memberID uint, isAdmin bool) error
}
```

**Create 流程：**
1. 校验图片/视频模式（videoUrl 非空 → len(images)==1）
2. 开启事务
3. INSERT note（status=1, publishTime=now）
4. FindOrCreate tags → INSERT note_tag
5. INSERT note_image
6. UPDATE member.publish_count +1
7. 提交事务
8. 组装 NoteResponse 返回

**GetByID 流程：**
1. 查 note（DeletedAt IS NULL）
2. goroutine IncrViewCount
3. 查 author / images / tags
4. 查 isLiked / isFavorited
5. 组装 NoteResponse

**Delete 流程：**
1. 查 note 存在且未删除
2. 作者删除：校验 memberID，事务中 SET status=0, DeletedAt + 级联软删
3. 管理员删除：事务中物理删除 note + 级联物理删所有关联表

### Controller

每个 Handler 从 `middleware.GetUserID(ctx)` 取当前用户 ID。

JWT 认证使用现有的 `middleware.Auth()`（内部已 `c.Set("user_id", claims.GetUserID())`）。

路由组：`/api/v1` → `/notes` 子组，5 个路由全部在 `Auth()` 中间件下。

### 装配（app.go）

```go
func (a *App) initDependencies() {
    noteRepo := repository.NewNoteRepository(a.mysqlDB)
    noteSvc := service.NewNoteService(noteRepo)
    noteCtrl := note.NewController(noteSvc)
    a.router = api.NewRouter(noteCtrl)
}
```

---

## 六、关联表说明

| 表 | 归属 | 操作方式 |
|----|------|----------|
| note | 自己 | CRUD |
| note_image | 自己 | 写入/软删/查询 |
| note_tag | 自己 | 写入/软删/查询 |
| note_like | 自己 | Toggle/软删/查询 |
| note_collect | 自己 | Toggle/软删/查询 |
| tag | 张峰源 | FindOrCreate（直接操作表） |
| member | 张兆焓 | SELECT（只读 author 信息） |

---

## 七、错误码

| 场景 | 错误码 | 说明 |
|------|--------|------|
| 参数校验失败 | 400 + binding 详情 | Gin 自动 |
| 未登录 | 401 | 中间件返回 |
| 非作者删除 | 403 | Service 层 |
| 笔记不存在 | 3001 | 资源不存在 |
| 图片视频同时存在 | 2001 | 参数错误 |
| 图片数量不正确 | 2001 | 视频模式 images!=1 |
| 标签数量不正确 | 2001 | tags 不在 1~3 范围 |
| 服务器错误 | 500 | 通用 |

---

## 八、后续阶段衔接

- **攻略模块**：note 和 strategy 子表结构一致，Repository/Service 代码可复制改表名
- **上传切换 MinIO**：只需改 Service 层的 URL 校验逻辑（dive,url → 查 MinIO 资源是否存在），接口不变
- **定时任务**：view_count 和 heat 热度字段由后续定时任务统一计算
