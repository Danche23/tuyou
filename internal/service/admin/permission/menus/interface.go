package menus

import (
	"context"
	request "tuyou/internal/model/dto/request/admin/permission/menus"
	response "tuyou/internal/model/dto/response/admin/permission/menus"
)

// MenuService 菜单管理服务接口
type MenuService interface {
	// ListMenus 查询菜单列表
	ListMenus(ctx context.Context, req *request.ListMenusReq) (*response.ListMenuResp, error)

	// ListAllMenus 查询全部菜单（父菜单选择器用）
	ListAllMenus(ctx context.Context) ([]response.MenuOptionItem, error)

	// InsertMenu 新增菜单
	InsertMenu(ctx context.Context, req *request.InsertMenuReq) (*response.InsertMenuResp, error)

	// UpdateMenu 编辑菜单
	UpdateMenu(ctx context.Context, id uint, req *request.UpdateMenuReq) error

	// DeleteMenu 删除菜单（级联删除子孙菜单）
	DeleteMenu(ctx context.Context, id uint) error
}
