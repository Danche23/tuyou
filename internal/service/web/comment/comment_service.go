package comment

import (
	responseComment "tuyou/internal/model/dto/response/web/comment"
	repoComment "tuyou/internal/repository/web/comment"
	"tuyou/pkg/errors"

	"gorm.io/gorm"
)

type commentService struct {
	db   *gorm.DB
	repo repoComment.CommentRepository
}

func NewCommentService(db *gorm.DB, repo repoComment.CommentRepository) CommentService {
	return &commentService{db: db, repo: repo}
}

// ToggleLike 点赞/取消点赞（翻转）
func (s *commentService) ToggleLike(commentID uint, memberID uint) (*responseComment.LikeToggleResponse, error) {
	// 先查评论存在（查不到顺带挡掉 commentID=0 的非法请求）
	comment, err := s.repo.FindByID(commentID)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeNotFound, "评论不存在", err)
	}

	var isLiked bool
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 先试取消：删到了 → 之前已点赞，本次取消，like_count -1
		deleted, err := s.repo.DeleteCommentLike(tx, commentID, memberID)
		if err != nil {
			return err
		}
		if deleted {
			if err := s.repo.DecrLikeCount(tx, commentID); err != nil {
				return err
			}
			isLiked = false
			return nil
		}
		// 没删到 → 之前未点赞，本次点赞，like_count +1
		if err := s.repo.CreateCommentLike(tx, commentID, memberID); err != nil {
			return err
		}
		if err := s.repo.IncrLikeCount(tx, commentID); err != nil {
			return err
		}
		isLiked = true
		return nil
	})
	if err != nil {
		return nil, err
	}

	// 最新点赞数 = 进来时读到的数 ± 1
	likeCount := comment.LikeCount
	if isLiked {
		likeCount++
	} else {
		likeCount--
	}
	return &responseComment.LikeToggleResponse{IsLiked: isLiked, LikeCount: likeCount}, nil
}
