# 「我的」会员中心 — 接口设计文档

**日期：** 2026-08-06
**状态：** 待开发（新模块）
**基础路径：** `/api/v1`

> 依据 `preview.html`（我的页面 H5）与 `member` / `note` / `strategy` / `activity` 等数据库表设计。
> 登录、注册、验证码、重置密码复用已有 `/auth/*` 接口；笔记、攻略、活动的新增/详情/点赞/收藏复用已有接口，本文档仅补充会员中心所需的新接口。

---

## 通用约定

**统一响应格式：**（与现有 `pkg/response` 一致）

```json
{
  "code": 0,
  "message": "成功",
  "data": {}
}
```

**认证：** 除标注外，全部接口需携带登录返回的 Token：

```
Authorization: Bearer <token>
```

**分页参数：** `page`（页码，从 1 开始，必填）、`size`（每页条数，最大 100，必填）。
分页响应统一为：

```json
{
  "code": 0,
  "message": "成功",
  "data": {
    "list": [],
    "total": 0,
    "page": 1,
    "size": 10,
    "total_page": 0
  }
}
```

**错误码：** 复用 `pkg/errors` 已有定义，常用：
- `401` 未授权 / 令牌失效（`1005` 无效令牌、`1006` 令牌过期）
- `400` 参数错误（`2001` 参数错误、`2002` 缺少必要参数）
- `1001` 用户不存在、`1004` 用户已被禁用
- `3001` 资源不存在

---

## 接口总览

| # | 方法 | 路径 | 说明 | 页面 |
|---|------|------|------|------|
| 1 | GET | `/member/center` | 获取我的主页信息（头像/昵称/签名/统计数） | 首页 |
| 2 | GET | `/member/profile` | 获取个人资料（编辑弹窗回显） | 编辑资料 |
| 3 | PUT | `/member/profile` | 更新个人资料（昵称/签名/头像） | 编辑资料 |
| 4 | POST | `/member/avatar` | 上传头像（multipart） | 编辑资料 |
| 5 | GET | `/member/publish` | 我的发布列表（笔记/攻略） | 我的发布 |
| 6 | GET | `/member/collect` | 我的收藏列表（笔记/攻略） | 我的收藏 |
| 7 | GET | `/member/like` | 我的点赞列表（笔记/攻略） | 我的点赞 |
| 8 | GET | `/member/activity` | 我的活动列表（发布/参与） | 我的活动 |
| 9 | GET | `/member/follows` | 我的关注列表 | 关注 |
| 10 | GET | `/member/fans` | 我的粉丝列表 | 粉丝 |
| 11 | POST | `/auth/logout` | 退出登录 | 设置 |
| 12 | POST | `/member/close` | 注销账号 | 设置 |

> 复用已有接口：登录 `/auth/login`、注册 `/auth/register`、发送验证码 `/auth/send-code`、重置密码 `/auth/reset-password`（对应设置页「忘记密码」）。
> 设置页「隐私设置」「消息通知」功能暂未定义，待需求明确后补充。

---

## 1. 获取我的主页信息

首页用户卡片 + 底部统计数（笔记攻略 / 关注 / 粉丝）。

```
GET /api/v1/member/center
```

**请求参数：** 无（当前登录用户）

**响应参数：**

| 字段 | 类型 | 说明 |
|------|------|------|
| data.id | int | 会员ID |
| data.avatar | string | 头像地址 |
| data.nickname | string | 昵称 |
| data.bio | string | 个性签名 |
| data.bgImg | string | 个人主页背景图 |
| data.role | int | 用户类型：0=领队，1=普通会员 |
| data.publishCount | int | 发布数（笔记+攻略）→ 首页「笔记攻略」 |
| data.followCount | int | 关注数 → 首页「关注」 |
| data.fansCount | int | 粉丝数 → 首页「粉丝」 |
| data.favoriteCount | int | 收藏数 |
| data.leadingCount | int | 带队次数（发布活动数，普通会员为 null） |

**响应示例：**

```json
{
  "code": 0,
  "message": "成功",
  "data": {
    "id": 1,
    "avatar": "https://picsum.photos/200/200?random=1",
    "nickname": "旅行爱好者阿途",
    "bio": "走遍山河，记录美好攻略",
    "bgImg": "",
    "role": 1,
    "publishCount": 28,
    "followCount": 136,
    "fansCount": 249,
    "favoriteCount": 45,
    "leadingCount": 0
  }
}
```

---

## 2. 获取个人资料

编辑资料弹窗回显。

```
GET /api/v1/member/profile
```

**请求参数：** 无（当前登录用户）

**响应参数：**

| 字段 | 类型 | 说明 |
|------|------|------|
| data.id | int | 会员ID |
| data.avatar | string | 头像地址 |
| data.nickname | string | 昵称 |
| data.bio | string | 个性签名 |
| data.phone | string | 手机号（脱敏） |
| data.bgImg | string | 个人主页背景图 |
| data.role | int | 用户类型 |

**响应示例：**

```json
{
  "code": 0,
  "message": "成功",
  "data": {
    "id": 1,
    "avatar": "https://picsum.photos/200/200?random=1",
    "nickname": "旅行爱好者阿途",
    "bio": "走遍山河，记录美好攻略",
    "phone": "138****0001",
    "bgImg": "",
    "role": 1
  }
}
```

---

## 3. 更新个人资料

保存编辑弹窗内容（昵称/签名/头像）。仅更新传入的字段，未传字段不修改。

```
PUT /api/v1/member/profile
```

**请求参数：**

| 字段 | 类型 | 必填 | 说明 |
|------|------|:---:|------|
| nickname | string | 否 | 昵称，最大 20 字 |
| bio | string | 否 | 个性签名，最大 100 字 |
| avatar | string | 否 | 头像地址（可由上传头像接口返回后回填） |

**请求示例：**

```json
{
  "nickname": "旅行爱好者阿途",
  "bio": "走遍山河，记录美好攻略"
}
```

**响应示例：**

```json
{
  "code": 0,
  "message": "资料修改成功",
  "data": {}
}
```

---

## 4. 上传头像

```
POST /api/v1/member/avatar
```

**请求参数（multipart/form-data）：**

| 字段 | 类型 | 必填 | 说明 |
|------|------|:---:|------|
| file | file | 是 | 头像图片文件 |

**响应参数：**

| 字段 | 类型 | 说明 |
|------|------|------|
| data.url | string | 上传成功后图片访问地址 |

**响应示例：**

```json
{
  "code": 0,
  "message": "成功",
  "data": {
    "url": "https://cdn.example.com/avatar/xxx.jpg"
  }
}
```

> 说明：图片地址需持久化时，前端可在保存资料时随 `avatar` 字段一起提交（接口 3）。

---

## 5. 我的发布列表

对应「我的发布」页，按标签切换查询笔记/攻略。

```
GET /api/v1/member/publish?type=note&page=1&size=10
```

**请求参数：**

| 字段 | 类型 | 必填 | 说明 |
|------|------|:---:|------|
| type | string | 是 | 内容类型：`note` 笔记 / `strategy` 攻略 |
| page | int | 是 | 页码，从 1 开始 |
| size | int | 是 | 每页条数，最大 100 |

**数据来源：** `note` / `strategy` 表，`member_id = 当前用户`。

**响应参数（data.list 元素，笔记与攻略字段差异）：**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | int | 内容ID |
| type | string | `note` / `strategy` |
| title | string | 标题 |
| coverImg | string | 封面图 |
| videoUrl | string | 视频地址 |
| content | string | 正文（列表可为摘要） |
| viewCount | int | 浏览量 |
| likeCount | int | 点赞数 |
| collectCount | int | 收藏数 |
| commentCount | int | 评论数 |
| category | int | 分类 |
| status | int | 上下架：0=下架，1=正常 |
| publishTime | string | 发布时间 |
| routeInfo | object | 仅攻略返回：distance/duration/difficulty/transport |

**响应示例：**

```json
{
  "code": 0,
  "message": "成功",
  "data": {
    "list": [
      {
        "id": 11,
        "type": "note",
        "title": "周末新乡周边短途游玩推荐",
        "coverImg": "https://...",
        "content": "周末不知道去哪……",
        "viewCount": 128,
        "likeCount": 18,
        "collectCount": 5,
        "commentCount": 2,
        "category": 1,
        "status": 1,
        "publishTime": "2026-07-28 10:00:00"
      }
    ],
    "total": 28,
    "page": 1,
    "size": 10,
    "total_page": 3
  }
}
```

---

## 6. 我的收藏列表

对应「我的收藏」页，按标签切换查询笔记/攻略。

```
GET /api/v1/member/collect?type=note&page=1&size=10
```

**请求参数：**

| 字段 | 类型 | 必填 | 说明 |
|------|------|:---:|------|
| type | string | 是 | `note` 笔记 / `strategy` 攻略 |
| page | int | 是 | 页码 |
| size | int | 是 | 每页条数 |

**数据来源：** `note_collect` / `strategy_collect` 表（`member_id = 当前用户`）关联 `note` / `strategy`。

**响应参数：** 同接口 5（data.list 元素一致），按收藏时间倒序。

---

## 7. 我的点赞列表

对应「我的点赞」页，按标签切换查询笔记/攻略。

```
GET /api/v1/member/like?type=note&page=1&size=10
```

**请求参数：**

| 字段 | 类型 | 必填 | 说明 |
|------|------|:---:|------|
| type | string | 是 | `note` 笔记 / `strategy` 攻略 |
| page | int | 是 | 页码 |
| size | int | 是 | 每页条数 |

**数据来源：** `note_like` / `strategy_like` 表（`member_id = 当前用户`）关联 `note` / `strategy`。

**响应参数：** 同接口 5（data.list 元素一致），按点赞时间倒序。

---

## 8. 我的活动列表

对应「我的活动」页，按标签切换查询我发布的/我参与的活动。

```
GET /api/v1/member/activity?type=publish&page=1&size=10
```

**请求参数：**

| 字段 | 类型 | 必填 | 说明 |
|------|------|:---:|------|
| type | string | 是 | `publish` 我发布的活动 / `join` 我参与的活动 |
| page | int | 是 | 页码 |
| size | int | 是 | 每页条数 |

**数据来源：**
- `publish`：`activity` 表，`leader_member_id = 当前用户`（仅领队有数据，普通会员返回空列表）
- `join`：`activity_join` 表（`member_id = 当前用户` 且 `is_cancel = 0`）关联 `activity`

**响应参数（data.list 元素）：**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | int | 活动ID |
| title | string | 活动标题 |
| coverImg | string | 封面图 |
| address | string | 活动地点 |
| startTime | string | 活动开始时间 |
| endTime | string | 活动结束时间 |
| maxPeople | int | 最大报名人数 |
| currentPeople | int | 当前报名人数 |
| price | number | 费用 |
| activityStatus | int | 状态：1报名中，2进行中，3已结束 |
| difficulty | int | 难度 |
| activityType | int | 活动类型 |
| isJoin | bool | 当前用户是否已报名（publish 场景为 false） |

**响应示例：**

```json
{
  "code": 0,
  "message": "成功",
  "data": {
    "list": [
      {
        "id": 3,
        "title": "新乡周末爬山结伴",
        "coverImg": "https://...",
        "address": "新乡市凤凰山",
        "startTime": "2026-08-10 08:00:00",
        "endTime": "2026-08-10 18:00:00",
        "maxPeople": 20,
        "currentPeople": 6,
        "price": 0,
        "activityStatus": 1,
        "difficulty": 1,
        "activityType": 1,
        "isJoin": true
      }
    ],
    "total": 1,
    "page": 1,
    "size": 10,
    "total_page": 1
  }
}
```

---

## 9. 我的关注列表

```
GET /api/v1/member/follows?page=1&size=10
```

**数据来源：** `follow` 表，`follower_id = 当前用户` 关联 `member`（被关注者）。

**响应参数（data.list 元素）：**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | int | 被关注用户ID |
| nickname | string | 昵称 |
| avatar | string | 头像 |
| bio | string | 个性签名 |
| role | int | 用户类型：0=领队，1=普通会员 |
| isFollowed | bool | 是否已关注（恒为 true） |

**响应示例：**

```json
{
  "code": 0,
  "message": "成功",
  "data": {
    "list": [
      {
        "id": 2,
        "nickname": "行走的骆驼",
        "avatar": "https://...",
        "bio": "热爱徒步和露营",
        "role": 1,
        "isFollowed": true
      }
    ],
    "total": 136,
    "page": 1,
    "size": 10,
    "total_page": 14
  }
}
```

---

## 10. 我的粉丝列表

```
GET /api/v1/member/fans?page=1&size=10
```

**数据来源：** `follow` 表，`followee_id = 当前用户` 关联 `member`（粉丝）。

**响应参数（data.list 元素）：** 同接口 9，其中 `isFollowed` 表示当前用户是否已回关该粉丝。

---

## 11. 退出登录

对应设置页「退出登录」。

```
POST /api/v1/auth/logout
```

**请求参数：** 无（携带 Token 即可）

**响应示例：**

```json
{
  "code": 0,
  "message": "退出登录成功",
  "data": {}
}
```

> 说明：客户端清除本地 Token；如需服务端强制失效，可后续将 Token 加入 Redis 黑名单。

---

## 12. 注销账号

对应设置页「注销账号」。

```
POST /api/v1/member/close
```

**请求参数：**

| 字段 | 类型 | 必填 | 说明 |
|------|------|:---:|------|
| password | string | 是 | 登录密码，用于身份校验 |

**响应示例：**

```json
{
  "code": 0,
  "message": "账号已注销",
  "data": {}
}
```

**业务逻辑：** 校验密码通过后，将 `member.status` 置为 `0`（禁用）并记录 `delete_time` 逻辑删除。

---

## 数据来源表说明

| 页面/功能 | 数据表 |
|-----------|--------|
| 首页用户卡片 / 统计数 | `member`（nickname/avatar/bio/publish_count/follow_count/fans_count） |
| 编辑资料 | `member`（nickname/bio/avatar/bg_img） |
| 我的发布 | `note`、`strategy`（member_id） |
| 我的收藏 | `note_collect`、`strategy_collect` |
| 我的点赞 | `note_like`、`strategy_like` |
| 我的活动（发布/参与） | `activity`（leader_member_id）、`activity_join` |
| 关注 / 粉丝 | `follow` |
