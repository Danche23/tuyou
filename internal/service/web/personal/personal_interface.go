package personal

import (
	responsePersonal "tuyou/internal/model/dto/response/web/personal"
)

type PersonalService interface {
	GetFollowedMembers(memberID uint) ([]responsePersonal.FollowedMember, error)
	GetFansMembers(memberID uint) ([]responsePersonal.FansMember, error)
}
