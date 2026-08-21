package message

import (
	"context"

	request "tuyou/internal/model/dto/request/admin/message"
	response "tuyou/internal/model/dto/response/admin/message"
)

// MessageService 后台消息管理服务接口
type MessageService interface {
	// ListInteractions 分页查询互动消息（msg_type=1）
	ListInteractions(ctx context.Context, req *request.ListMessageReq) (*response.ListMessageResp, error)
	// ListSystem 分页查询系统消息（msg_type=2）
	ListSystem(ctx context.Context, req *request.ListMessageReq) (*response.ListMessageResp, error)
	// PublishSystemMsg 发布系统消息：全体/指定会员，落库后推前台未读数
	PublishSystemMsg(ctx context.Context, req *request.PublishSystemMsgReq) error
	// Delete 硬删除一条消息（连带其全部接收记录）
	Delete(ctx context.Context, id uint) error
}
