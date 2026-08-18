package personal

import (
	stdErrors "errors"
	"testing"
	"time"

	"tuyou/internal/model/entity"
	repoActivity "tuyou/internal/repository/web/activity"
	repoPersonal "tuyou/internal/repository/web/personal"
	appErrors "tuyou/pkg/errors"
	"tuyou/pkg/sensitive"

	"gorm.io/gorm"
)

// mockPubRepo 只实现 publish-activity 路径用到的 PersonalRepository 方法
type mockPubRepo struct {
	repoPersonal.PersonalRepository
	activities   []entity.Activity
	total        int64
	participants map[uint][]entity.Member
}

func (m mockPubRepo) FindPublishActivities(memberID uint, page, size int) ([]entity.Activity, int64, error) {
	return m.activities, m.total, nil
}

func (m mockPubRepo) FindParticipantsByActivityIDs(activityIDs []uint) (map[uint][]entity.Member, error) {
	return m.participants, nil
}

// mockActRepo 只实现 GetMembersByIDs
type mockActRepo struct {
	repoActivity.ActivityRepository
	members []entity.Member
}

func (m mockActRepo) GetMembersByIDs(ids []uint) ([]entity.Member, error) {
	return m.members, nil
}

func TestBuildParticipants_MasksNickname(t *testing.T) {
	svc := &personalService{sensitive: sensitive.New([]string{"赌博"})}
	got := svc.buildParticipants([]entity.Member{
		{BaseEntity: entity.BaseEntity{ID: 28}, Nickname: "赌博乙", Avatar: "a1", Phone: "13800000028"},
	})
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
	if got[0].ID != "28" {
		t.Fatalf("id 应为字符串，got %q", got[0].ID)
	}
	if got[0].Nickname != "**乙" {
		t.Fatalf("昵称应打星，got %q", got[0].Nickname)
	}
	if got[0].Phone != "13800000028" {
		t.Fatalf("手机号应透传，got %q", got[0].Phone)
	}
}

func TestGetPublishActivities_Participants(t *testing.T) {
	now := time.Now()
	act := entity.Activity{
		BaseEntity:     entity.BaseEntity{ID: 1},
		LeaderMemberID: 27,
		Title:          "周末徒步",
		StartTime:      now,
		EndTime:        now.Add(2 * time.Hour),
		SignUpStart:    now,
		SignUpEnd:      now.Add(-time.Hour),
	}
	repo := mockPubRepo{
		activities: []entity.Activity{act},
		total:      1,
		participants: map[uint][]entity.Member{
			1: {{BaseEntity: entity.BaseEntity{ID: 28}, Nickname: "赌博乙", Avatar: "a2", Phone: "13800000028"}},
		},
	}
	svc := &personalService{
		repo:         repo,
		activityRepo: mockActRepo{members: []entity.Member{{BaseEntity: entity.BaseEntity{ID: 27}, Nickname: "测试甲"}}},
		sensitive:    sensitive.New([]string{"赌博"}),
	}

	list, total, err := svc.GetPublishActivities(27, 1, 10)
	if err != nil {
		t.Fatalf("GetPublishActivities: %v", err)
	}
	if total != 1 || len(list) != 1 {
		t.Fatalf("total/list = %d/%d, want 1/1", total, len(list))
	}
	if len(list[0].Participants) != 1 {
		t.Fatalf("participants len = %d, want 1", len(list[0].Participants))
	}
	if list[0].Participants[0].Nickname != "**乙" {
		t.Fatalf("报名人昵称应打星，got %q", list[0].Participants[0].Nickname)
	}
	if list[0].Participants[0].Phone != "13800000028" {
		t.Fatalf("报名人手机号应透传，got %q", list[0].Participants[0].Phone)
	}
}

// mockCardRepo 资料卡路径用到的 PersonalRepository 方法
type mockCardRepo struct {
	repoPersonal.PersonalRepository
	member         *entity.Member
	myFollowed     map[uint]bool // currentMemberID 关注集合
	targetFollowed map[uint]bool // targetID 关注集合
	findErr        error
}

func (m mockCardRepo) FindByID(id uint) (*entity.Member, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	return m.member, nil
}

func (m mockCardRepo) FindFollowedIDs(followerID uint) (map[uint]bool, error) {
	if m.member != nil && followerID == m.member.ID {
		return m.targetFollowed, nil
	}
	return m.myFollowed, nil
}

func TestGetMemberCard(t *testing.T) {
	svc := &personalService{
		repo: mockCardRepo{
			member: &entity.Member{
				BaseEntity:    entity.BaseEntity{ID: 28},
				Nickname:      "赌博乙",
				Avatar:        "a",
				Bio:           "带队十年",
				BgImg:         "b",
				Role:          0,
				Status:        1,
				FollowCount:   5,
				FansCount:     7,
				PublishCount:  3,
				FavoriteCount: 2,
				LeadingCount:  nil,
			},
			myFollowed:     map[uint]bool{28: true},
			targetFollowed: map[uint]bool{27: true},
		},
		sensitive: sensitive.New([]string{"赌博"}),
	}

	card, err := svc.GetMemberCard(27, 28)
	if err != nil {
		t.Fatalf("GetMemberCard: %v", err)
	}
	if card.Nickname != "**乙" {
		t.Fatalf("昵称应打星，got %q", card.Nickname)
	}
	if card.Role != "领队" {
		t.Fatalf("role 应为领队，got %q", card.Role)
	}
	if card.Status != 1 {
		t.Fatalf("status = %d, want 1", card.Status)
	}
	if card.LeadingCount != 0 {
		t.Fatalf("leadingCount nil 应为 0，got %d", card.LeadingCount)
	}
	if !card.IsFollowed {
		t.Fatalf("isFollowed 应为 true")
	}
	if !card.IsFollower {
		t.Fatalf("isFollower 应为 true")
	}
}

func TestGetMemberCard_NotFound(t *testing.T) {
	svc := &personalService{repo: mockCardRepo{findErr: gorm.ErrRecordNotFound}}
	_, err := svc.GetMemberCard(27, 999)
	if !stdErrors.Is(err, appErrors.ErrUserNotFound) {
		t.Fatalf("应为 ErrUserNotFound，got %v", err)
	}
}
