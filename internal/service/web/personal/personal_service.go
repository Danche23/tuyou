package personal

import (
	responsePersonal "tuyou/internal/model/dto/response/web/personal"
	repoPersonal "tuyou/internal/repository/web/personal"
)

type personalService struct {
	repo repoPersonal.PersonalRepository
}

// NewPersonalService 创建服务实例
func NewPersonalService(repo repoPersonal.PersonalRepository) PersonalService {
	return &personalService{repo: repo}
}

// GetFollowedMembers 获取我关注的用户列表
func (s *personalService) GetFollowedMembers(memberID uint) ([]responsePersonal.FollowedMember, error) {
	members, err := s.repo.FindFollowedMembers(memberID)
	if err != nil {
		return nil, err
	}

	result := make([]responsePersonal.FollowedMember, 0, len(members))
	for _, m := range members {
		result = append(result, responsePersonal.FollowedMember{
			ID:         m.ID,
			Nickname:   m.Nickname,
			Avatar:     m.Avatar,
			Bio:        m.Bio,
			MemberType: m.MemberType,
		})
	}
	return result, nil
}

func (s *personalService) GetFansMembers(memberID uint) ([]responsePersonal.FansMember, error) {
	members, err := s.repo.FindFansMembers(memberID)
	if err != nil {
		return nil, err
	}
	result := make([]responsePersonal.FansMember, 0, len(members))
	for _, m := range members {
		result = append(result, responsePersonal.FansMember{
			ID:         m.ID,
			Nickname:   m.Nickname,
			Avatar:     m.Avatar,
			Bio:        m.Bio,
			MemberType: m.MemberType,
		})
	}
	return result, nil
}
