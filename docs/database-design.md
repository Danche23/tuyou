# tuyou 数据库设计文档

> 数据库：MySQL 8.0  
> ORM：GORM（Go），使用 AutoMigrate 管理表结构  
> 最后更新：2026-10-15

---

## 1. 总体说明

tuyou 是户外旅行平台，提供前台会员端和后台运营管理端两套业务。数据库共 **28 张表**，分为六个域：

| 域 | 表 | 说明 |
|---|---|---|
| 会员与社交 | `member` `follow` | 会员主表、关注关系 |
| 内容 | `note` `note_image` `note_like` `note_collect` `note_tag` `strategy` `strategy_image` `strategy_like` `strategy_collect` `strategy_tag` `activity` `activity_join` `tag` | 笔记、攻略、活动及互动 |
| 评论 | `comment` `comment_like` | 笔记/攻略共用的两级评论 |
| 消息 | `msg_info` `msg_template` `msg_user_record` | 系统消息与互动通知 |
| 后台权限（RBAC） | `sys_admin` `sys_role` `sys_menu` `sys_role_menu` `sys_permission` `sys_role_perm` | 管理员、角色、菜单、接口权限 |
| 其他 | `sensitive_word` `search_history` | 敏感词词库、搜索历史 |

---

## 2. 通用约定

| 项目 | 约定 |
|---|---|
| 主键 | `uint`，GORM 默认自增 |
| 时间列 | `create_time` / `update_time` / `delete_time`（来自 `BaseEntity`） |
| 软删除 | 所有表继承 `gorm.DeletedAt`（`delete_time`），查询自动过滤已删除行 |
| 字符集 | `utf8mb4` |
| 命名 | 表名 `snake_case`，列名 `snake_case` |
| 枚举 | 用 `tinyint` + 注释说明取值，不用 MySQL ENUM |

**BaseEntity（所有表共用）：**

```go
type BaseEntity struct {
    ID        uint           // primary key, auto-increment
    CreatedAt time.Time      // column: create_time
    UpdatedAt time.Time      // column: update_time
    DeletedAt gorm.DeletedAt // column: delete_time, index（软删除）
}
```

---

## 3. 会员与社交

### 3.1 member（会员主表）

| 列 | 类型 | 说明 |
|---|---|---|
| id | uint PK | |
| avatar | varchar(255) | 头像地址，默认值见代码常量 |
| nickname | varchar(20) NOT NULL | 昵称 |
| phone | varchar(20) NOT NULL | 登录手机号，注销后可重新注册（代码级唯一性校验） |
| password | varchar(255) NOT NULL | bcrypt 哈希，JSON 序列化时隐藏 |
| bio | varchar(100) | 个人简介 |
| bg_img | varchar(255) | 主页背景图 |
| role | tinyint index | 0=领队，1=普通会员 |
| status | tinyint index | 0=禁用，1=正常，2=已注销 |
| last_login_time | datetime | 最近登录时间 |
| follow_count | bigint default 0 | 关注数（冗余计数） |
| favorite_count | bigint default 0 | 收藏数（冗余计数） |
| publish_count | bigint default 0 | 发布数（笔记+攻略） |
| fans_count | bigint default 0 | 粉丝数（冗余计数） |
| leading_count | bigint default 0 | 带队次数（领队专用，普通会员为 NULL） |

### 3.2 follow（关注关系表）

| 列 | 类型 | 说明 |
|---|---|---|
| follower_id | uint uniqueIndex(follower,followee) | 关注者 |
| followee_id | uint uniqueIndex(follower,followee) index | 被关注者 |
| status | tinyint NOT NULL default 1 | 1=已关注，0=已取关（位翻转，不物理删除） |

> 设计说明：关注/取关不物理删除记录，用 `status` 位翻转表达状态，保留历史并便于统计。

---

## 4. 内容域

笔记（note）和攻略（strategy）结构高度同构，复用同一套标签表和评论表。

### 4.1 note（游记笔记）

| 列 | 类型 | 说明 |
|---|---|---|
| member_id | uint index | 发布用户 |
| title | varchar(100) NOT NULL | 标题 |
| video_url | varchar(255) | 视频地址 |
| content | text | 正文 |
| is_sensitive | tinyint default 0 index | 发布时敏感词命中标记，0=否，1=是 |
| view_count | bigint default 0 | 浏览次数（Redis 缓存后异步落库） |
| like_count | bigint default 0 | 点赞数 |
| collect_count | bigint default 0 | 收藏数 |
| comment_count | bigint default 0 | 评论数 |
| status | tinyint default 1 index | 0=下架，1=正常 |
| publish_time | datetime NOT NULL index | 发布时间 |
| is_slide | tinyint default 0 index | 1=首页幻灯片推荐 |
| category | tinyint index | 内容分类 |
| heat | int default 0 index | 热度（定时任务周期重算） |

### 4.2 note_image（笔记图片）

| 列 | 类型 | 说明 |
|---|---|---|
| note_id | uint index | 关联笔记 |
| url | varchar(255) NOT NULL | 图片 CDN 地址 |
| sort | int default 0 | 排序，0=封面图 |

### 4.3 note_like / note_collect（点赞/收藏）

两表结构相同：`note_id + member_id` 组成联合唯一索引，不重复记录。

### 4.4 strategy（旅行攻略）

与 note 同构，额外字段：

| 列 | 类型 | 说明 |
|---|---|---|
| distance | varchar(20) | 距离（公里） |
| duration | varchar(20) | 总时长 |
| difficulty | tinyint | 难度：入门/中等/中高等/高等 |
| transport | varchar(20) | 交通方式 |

### 4.5 tag（通用标签，笔记/攻略共用）

| 列 | 类型 | 说明 |
|---|---|---|
| tag_name | varchar(30) NOT NULL uniqueIndex | 标签名 |
| status | tinyint default 1 index | 0=禁用，1=启用 |
| is_hot | tinyint default 0 index | 1=热门搜索关键词 |

### 4.6 note_tag / strategy_tag（内容-标签关联）

`{note/strategy}_id + tag_id` 联合唯一索引。

### 4.7 activity（活动）

| 列 | 类型 | 说明 |
|---|---|---|
| leader_member_id | uint index | 发布人，代码校验 role=0（领队） |
| title | varchar(50) NOT NULL | |
| cover_img | varchar(255) | 封面图 |
| activity_content | text | 活动详情 |
| address | varchar(200) | 线下地点 |
| is_sensitive | tinyint default 0 index | 敏感词标记 |
| start_time / end_time | datetime NOT NULL index | 活动时间段 |
| sign_up_start / sign_up_end | datetime NOT NULL | 报名时间段 |
| max_people | int default 20 | 最大报名人数 |
| current_people | int default 0 | 当前已报名人数 |
| price | decimal(10,2) default 0 | 费用 |
| activity_status | tinyint default 1 index | 1=报名中，2=进行中，3=已结束 |
| difficulty | tinyint | 难度 |
| activity_type | tinyint index | 活动类型 |
| is_slide | tinyint default 0 index | 1=首页幻灯片推荐 |
| is_del | tinyint default 0 | 逻辑删除（额外字段，独立于 BaseEntity 的软删除） |

### 4.8 activity_join（活动报名记录）

| 列 | 类型 | 说明 |
|---|---|---|
| activity_id | uint uniqueIndex(activity,member) | |
| member_id | uint uniqueIndex(activity,member) | |
| is_cancel | tinyint default 0 index | 0=正常，1=用户取消 |

---

## 5. 评论域

### 5.1 comment（评论主表，笔记/攻略共用）

| 列 | 类型 | 说明 |
|---|---|---|
| target_type | tinyint NOT NULL index(target) | 1=笔记评论，2=攻略评论 |
| target_id | uint NOT NULL index(target) | 关联笔记/攻略 ID |
| member_id | uint NOT NULL index | 评论者 |
| content | text NOT NULL | 评论内容 |
| reply_to_id | uint NULL index | 回复目标评论 ID，NULL=一级评论 |
| reply_to_member_id | uint NULL index | 被回复的用户，NULL=一级评论 |
| parent_id | uint NULL index | 直接父评论 ID，消息中心回复列表按此绑定 |
| like_count | bigint default 0 | 点赞数 |
| heat | bigint default 0 | 热度值（含时间因子，异步更新） |
| heat_score | bigint default 0 | 累计热度分（只增不减） |
| heat_scored | tinyint default 0 | 二级评论是否已计入父级热度 |
| is_sensitive | tinyint default 0 | 敏感词命中标记 |
| hit_words | varchar(255) default '' | 命中的敏感词，分号分隔 |
| status | tinyint default 1 index | 0=下架/删除，1=正常 |

> 两级评论模型：一级评论 `parent_id=NULL`；二级评论 `parent_id` 指向一级评论，`reply_to_id` 指向实际被回复那条（可能是二级），`reply_to_member_id` 用于展示"回复 @xxx"。

### 5.2 comment_like（评论点赞）

`comment_id + member_id` 联合唯一索引。

---

## 6. 消息域

### 6.1 msg_template（消息模板）

| 列 | 类型 | 说明 |
|---|---|---|
| template_name | varchar(50) NOT NULL | |
| msg_type | tinyint index | 1=互动，2=系统 |
| content_template | text | 支持变量占位符 |
| fixed_tip | varchar(500) | 活动通知专用固定提醒文案 |

### 6.2 msg_info（消息主表）

| 列 | 类型 | 说明 |
|---|---|---|
| msg_type | tinyint index | 1=互动，2=系统通知 |
| receive_obj_type | tinyint index | 0=全体用户，1=指定会员，2=指定管理员 |
| sender_id | uint default 0 index | 0=系统，其他=管理员/会员 id |
| template_id | uint index | 关联模板 |
| msg_content | text | 渲染后的最终消息正文 |
| send_time | datetime index | 消息生成时间 |
| comment_id | uint default 0 index | 互动消息关联评论（用于回复定位） |
| is_del | tinyint default 0 | 后台逻辑删除 |

### 6.3 msg_user_record（用户消息阅读记录）

| 列 | 类型 | 说明 |
|---|---|---|
| msg_id | uint NOT NULL index | 关联消息主表 |
| receive_member_id | uint NOT NULL index | 接收会员 |
| is_read | tinyint default 0 index | 0=未读，1=已读 |
| is_user_del | tinyint default 0 | 用户是否删除该消息 |

---

## 7. 后台权限（RBAC）

权限模型：管理员通过 `role_id` 绑定**单一角色**（多对一）；角色与菜单、接口权限均为**多对多**。

即：`sys_admin.role_id → sys_role`，以及 `sys_role →(sys_role_menu)→ sys_menu`、`sys_role →(sys_role_perm)→ sys_permission`。

### 7.1 sys_admin（管理员）

| 列 | 类型 | 说明 |
|---|---|---|
| username | varchar(20) NOT NULL uniqueIndex | 后台登录账号 |
| password | varchar(255) NOT NULL | 加密登录密码（bcrypt），JSON 序列化时隐藏 |
| admin_name | varchar(20) | 管理员真实姓名 |
| role_id | uint index | 关联 sys_role 角色表 id（单角色） |
| avatar | varchar(255) | 管理员头像路径 |
| status | tinyint default 1 index | 0=禁用，1=正常 |
| last_login_time | datetime | 最后登录时间 |

### 7.2 sys_role（角色）

| 列 | 类型 | 说明 |
|---|---|---|
| role_name | varchar(50) NOT NULL | 角色名称 |
| role_code | varchar(50) NOT NULL uniqueIndex | 鉴权唯一编码 |
| description | varchar(500) | 描述 |
| status | tinyint default 1 index | 0=禁用，1=正常 |

### 7.3 sys_menu（菜单）

| 列 | 类型 | 说明 |
|---|---|---|
| menu_name | varchar(50) NOT NULL | 菜单展示名称 |
| parent_id | uint default 0 index | 父菜单 ID，0=一级菜单 |
| menu_type | tinyint index | 1=目录，2=页面菜单 |
| route_path | varchar(255) | 前端页面路由地址 |
| sort | int | 菜单排序序号 |
| status | tinyint default 1 index | 0=隐藏菜单，1=展示菜单 |
| component | varchar(255) | 前端组件路径 |
| icon | varchar(100) | 菜单图标 |
| hide_in_menu | tinyint default 0 | 是否在侧边栏隐藏，0=显示，1=隐藏 |

### 7.4 sys_permission（接口权限）

| 列 | 类型 | 说明 |
|---|---|---|
| perm_name | varchar(30) NOT NULL | 权限名称 |
| module | varchar(40) index | 所属业务模块 |
| perm_key | varchar(100) uniqueIndex | 后端鉴权唯一标识 |
| request_method | varchar(10) | GET/POST/PUT/DELETE |
| api_path | varchar(255) | 接口路径 |
| status | tinyint default 1 index | 0=禁用，1=启用 |

### 7.5 sys_role_menu / sys_role_perm（关联中间表）

各自以 `role_id + {menu/perm}_id` 联合唯一索引。

---

## 8. 其他

### 8.1 sensitive_word（敏感词）

| 列 | 类型 | 说明 |
|---|---|---|
| word | varchar(100) NOT NULL uniqueIndex | 词内容 |
| status | tinyint default 1 index | 1=启用，0=停用 |

> 启动时从表加载启用词构建内存 Trie；后台变更后原子替换内存词库，无需重启。加载失败直接 Fatal，不带病上线。

### 8.2 search_history（搜索历史）

| 列 | 类型 | 说明 |
|---|---|---|
| member_id | uint uniqueIndex(member,keyword) | |
| keyword | varchar(50) NOT NULL uniqueIndex(member,keyword) | |

> 每位用户去重保留最近 20 条，超出时删除最旧记录。

---

## 9. 关键设计决策

**浏览量异步落库**：前台请求先写 Redis，由 `stream.Consumer` 批量回写 MySQL，避免高频写打满数据库。

**热度定时重算**：点赞/评论热度不在请求链路内计算，由 `task/heat.go`、`task/comment_heat.go` 周期性校准，`heat_score` 只增不减作为历史基线。

**冗余计数字段**：`member.follow_count`、`note.like_count` 等冗余计数在业务写路径内同步更新，并由定时校准任务（`member_publish.go`）对齐防止漂移。

**敏感词快照**：内容发布时将命中结果快照存入 `is_sensitive`、`hit_words`，原文不变，便于审核和追溯，不影响用户已发布内容。

**follow 位翻转**：关注/取关不物理删除，用 `status` 位翻转，保留关系历史，统计关注数时只计 `status=1`。
