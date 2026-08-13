package center

import (
	"context"
	stdErrors "errors"
	"math"
	"strconv"

	"gorm.io/gorm"
	response "tuyou/internal/model/dto/response/admin/auth"
	responseCenter "tuyou/internal/model/dto/response/admin/center"
	repoAuth "tuyou/internal/repository/admin/auth"
	repoCenter "tuyou/internal/repository/admin/center"
	"tuyou/internal/stream"
	appErrors "tuyou/pkg/errors"
)

type centerService struct {
	centerRepo   repoCenter.CenterRepository
	adminRepo    repoAuth.AdminRepository
	adminMsgRepo repoCenter.AdminMsgRepository
	adminHub     *stream.SSEHub
}

func NewCenterService(centerRepo repoCenter.CenterRepository, adminRepo repoAuth.AdminRepository, adminMsgRepo repoCenter.AdminMsgRepository, adminHub *stream.SSEHub) CenterService {
	return &centerService{centerRepo: centerRepo, adminRepo: adminRepo, adminMsgRepo: adminMsgRepo, adminHub: adminHub}
}

// UpdateAvatar 更新头像
func (s *centerService) UpdateAvatar(ctx context.Context, adminID uint, avatar string) (*response.AdminInfo, error) {
	if err := s.centerRepo.UpdateAvatar(ctx, adminID, avatar); err != nil {
		return nil, err
	}
	return s.getProfile(ctx, adminID)
}

// UpdateUsername 更新账号名
func (s *centerService) UpdateUsername(ctx context.Context, adminID uint, username string) (*response.AdminInfo, error) {
	if err := s.centerRepo.UpdateUsername(ctx, adminID, username); err != nil {
		return nil, err
	}
	return s.getProfile(ctx, adminID)
}

// GetNotifications 分页获取当前管理员个人消息
func (s *centerService) GetNotifications(ctx context.Context, adminID uint, page, size int) (*responseCenter.AdminMsgPage, error) {
	page, size = normalizePage(page, size)
	rows, total, err := s.adminMsgRepo.List(ctx, adminID, page, size)
	if err != nil {
		return nil, appErrors.NewWithErr(appErrors.CodeInternalError, "查询消息列表失败", err)
	}
	return &responseCenter.AdminMsgPage{
		List:      buildNotifications(rows),
		Total:     total,
		Page:      page,
		Size:      size,
		TotalPage: int(math.Ceil(float64(total) / float64(size))),
	}, nil
}

// GetUnreadCount 获取当前管理员未读消息数
func (s *centerService) GetUnreadCount(ctx context.Context, adminID uint) (*responseCenter.AdminUnreadCount, error) {
	cnt, err := s.adminMsgRepo.UnreadCount(ctx, adminID)
	if err != nil {
		return nil, appErrors.NewWithErr(appErrors.CodeInternalError, "查询未读数失败", err)
	}
	return &responseCenter.AdminUnreadCount{UnreadCount: cnt}, nil
}

// MarkRead 标记单条消息已读
func (s *centerService) MarkRead(ctx context.Context, adminID, msgID uint) error {
	err := s.adminMsgRepo.MarkRead(ctx, adminID, msgID)
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			return appErrors.New(appErrors.CodeNotFound, "消息不存在")
		}
		return appErrors.NewWithErr(appErrors.CodeInternalError, "标记已读失败", err)
	}
	s.pushUnread(ctx, adminID)
	return nil
}

// MarkAllRead 标记全部消息已读
func (s *centerService) MarkAllRead(ctx context.Context, adminID uint) error {
	if err := s.adminMsgRepo.MarkAllRead(ctx, adminID); err != nil {
		return appErrors.NewWithErr(appErrors.CodeInternalError, "标记全部已读失败", err)
	}
	s.pushUnread(ctx, adminID)
	return nil
}

// Delete 删除单条消息（软删除）
func (s *centerService) Delete(ctx context.Context, adminID, msgID uint) error {
	err := s.adminMsgRepo.Delete(ctx, adminID, msgID)
	if err != nil {
		if stdErrors.Is(err, gorm.ErrRecordNotFound) {
			return appErrors.New(appErrors.CodeNotFound, "消息不存在")
		}
		return appErrors.NewWithErr(appErrors.CodeInternalError, "删除消息失败", err)
	}
	s.pushUnread(ctx, adminID)
	return nil
}

// ClearAll 清空所有消息（软删除）
func (s *centerService) ClearAll(ctx context.Context, adminID uint) error {
	if err := s.adminMsgRepo.ClearAll(ctx, adminID); err != nil {
		return appErrors.NewWithErr(appErrors.CodeInternalError, "清空消息失败", err)
	}
	s.pushUnread(ctx, adminID)
	return nil
}

// pushUnread 推送该管理员最新未读数（角标实时更新），best-effort
func (s *centerService) pushUnread(ctx context.Context, adminID uint) {
	if uc, err := s.adminMsgRepo.UnreadCount(ctx, adminID); err == nil {
		s.adminHub.Publish(adminID, "unread", &responseCenter.AdminUnreadCount{UnreadCount: uc})
	}
}

// normalizePage 分页参数归一化：page 默认 1，size 默认 10，上限 100
func normalizePage(page, size int) (int, int) {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 10
	}
	if size > 100 {
		size = 100
	}
	return page, size
}

// buildNotifications 联查结果 → 响应项
func buildNotifications(rows []repoCenter.AdminMsgRow) []responseCenter.AdminNotification {
	items := make([]responseCenter.AdminNotification, 0, len(rows))
	for _, r := range rows {
		items = append(items, responseCenter.AdminNotification{
			ID:       r.ID,
			Type:     r.Type,
			Content:  r.Content,
			IsRead:   r.IsRead,
			SendTime: r.SendTime,
		})
	}
	return items
}

// statusStr 将账号状态映射为展示文案：1启用，0禁用
func statusStr(status int) string {
	if status == 1 {
		return "启用"
	}
	return "禁用"
}

// getProfile 查询更新后的管理员信息
func (s *centerService) getProfile(ctx context.Context, adminID uint) (*response.AdminInfo, error) {
	admin, err := s.adminRepo.FindByID(ctx, adminID)
	if err != nil {
		return nil, err
	}
	role, err := s.adminRepo.FindRoleNameByRoleID(ctx, admin.RoleID)
	if err != nil {
		return nil, err
	}
	return &response.AdminInfo{
		ID:        strconv.Itoa(int(admin.ID)),
		Username:  admin.Username,
		Adminname: admin.AdminName,
		Role:      role,
		Avatar:    admin.Avatar,
		Status:    statusStr(admin.Status),
	}, nil
}
