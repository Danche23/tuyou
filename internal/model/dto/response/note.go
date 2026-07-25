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
