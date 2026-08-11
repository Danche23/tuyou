package menus

import (
	"context"
	"errors"
	"math"
	"strconv"

	request "tuyou/internal/model/dto/request/admin/permission/menus"
	response "tuyou/internal/model/dto/response/admin/permission/menus"
	"tuyou/internal/model/entity"
	"tuyou/internal/repository/admin/permission/menus"
	appErrors "tuyou/pkg/errors"
)

type menuService struct {
	menuRepo menus.MenuRepository
}

func NewMenuService(menuRepo menus.MenuRepository) MenuService {
	return &menuService{menuRepo: menuRepo}
}

// menuTypeToInt 菜单类型中文转 int：目录=1，菜单=2
func menuTypeToInt(t string) int {
	if t == "菜单" {
		return 2
	}
	return 1
}

// menuTypeName 菜单类型 int 转中文
func menuTypeName(t int) string {
	switch t {
	case 1:
		return "目录"
	case 2:
		return "菜单"
	default:
		return ""
	}
}

// statusToInt 状态中文转 int：启用=1，禁用=0
func statusToInt(s string) int {
	if s == "禁用" {
		return 0
	}
	return 1
}

// boolToInt bool 转 int：true=1，false=0
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// ListMenus 查询菜单列表
func (s *menuService) ListMenus(ctx context.Context, req *request.ListMenusReq) (*response.ListMenuResp, error) {
	filters := map[string]interface{}{}
	if req.MenuName != "" {
		filters["menu_name"] = req.MenuName
	}

	if req.Status != "" {
		filters["status"] = statusToInt(req.Status)
	}

	total, err := s.menuRepo.Count(ctx, filters)
	if err != nil {
		return nil, appErrors.ErrInternalError
	}

	list, err := s.menuRepo.List(ctx, filters, int(req.Page), int(req.Size))
	if err != nil {
		return nil, appErrors.ErrInternalError
	}

	// 补充菜单类型中文名
	for i := range list {
		list[i].MenuTypeName = menuTypeName(list[i].MenuType)
	}

	return &response.ListMenuResp{
		Total:     int(total),
		Page:      int(req.Page),
		Size:      int(req.Size),
		TotalPage: int(math.Ceil(float64(total) / float64(req.Size))),
		List:      list,
	}, nil
}

// ListAllMenus 查询全部菜单（父菜单选择器用）
func (s *menuService) ListAllMenus(ctx context.Context) ([]response.MenuOptionItem, error) {
	list, err := s.menuRepo.ListAll(ctx)
	if err != nil {
		return nil, appErrors.ErrInternalError
	}
	return list, nil
}

// resolveParentID 将父菜单名解析为父菜单 ID，空串返回根（0）
func (s *menuService) resolveParentID(ctx context.Context, parentName string) (uint, error) {
	if parentName == "" {
		return 0, nil
	}
	parent, err := s.menuRepo.FindByName(ctx, parentName)
	if err != nil {
		if errors.Is(err, appErrors.ErrNotFound) {
			return 0, appErrors.New(appErrors.CodeResourceNotFound, "父菜单不存在")
		}
		return 0, err
	}
	return parent.ID, nil
}

// InsertMenu 新增菜单
func (s *menuService) InsertMenu(ctx context.Context, req *request.InsertMenuReq) (*response.InsertMenuResp, error) {
	// 重名校验
	exist, err := s.menuRepo.FindByName(ctx, req.MenuName)
	if err != nil && !errors.Is(err, appErrors.ErrNotFound) {
		return nil, err
	}
	if exist != nil {
		return nil, appErrors.New(appErrors.CodeResourceAlreadyExists, "菜单名称已存在")
	}

	parentID, err := s.resolveParentID(ctx, req.ParentName)
	if err != nil {
		return nil, err
	}

	status := statusToInt(req.Status)
	menuType := menuTypeToInt(req.MenuType)

	menu := &entity.SysMenu{
		MenuName:   req.MenuName,
		ParentID:   parentID,
		MenuType:   menuType,
		RoutePath:  req.RoutePath,
		Sort:       req.Sort,
		Status:     status,
		Component:  req.Component,
		Icon:       req.Icon,
		HideInMenu: boolToInt(req.HideInMenu),
	}
	// 类型规则：目录(1)组件路径为空；子菜单(有父级)图标为空，一级菜单保留图标
	if menuType == 1 {
		menu.Component = "" // 目录不展示组件路径
	}
	if parentID != 0 {
		menu.Icon = "" // 子菜单不展示图标
	}
	if err := s.menuRepo.Insert(ctx, menu); err != nil {
		return nil, err
	}

	// 类型联动：父菜单有子菜单了，类型自动改为"目录"；目录不展示组件路径，component 一并清空
	if parentID != 0 {
		if parent, err := s.menuRepo.FindByID(ctx, parentID); err == nil && parent.MenuType == 2 {
			_ = s.menuRepo.Update(ctx, parentID, map[string]interface{}{"menu_type": 1, "component": ""})
		}
	}

	return &response.InsertMenuResp{
		ID:       strconv.Itoa(int(menu.ID)),
		MenuName: menu.MenuName,
	}, nil
}

// UpdateMenu 编辑菜单
func (s *menuService) UpdateMenu(ctx context.Context, id uint, req *request.UpdateMenuReq) error {
	// 校验菜单存在，并记录当前类型（用于类型规则）
	menu, err := s.menuRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	updates := map[string]interface{}{}
	newParentID := menu.ParentID // 更新后的父级，用于判断是否为子菜单
	if req.MenuName != nil {
		updates["menu_name"] = *req.MenuName
	}
	if req.ParentName != nil {
		parentID, err := s.resolveParentID(ctx, *req.ParentName)
		if err != nil {
			return err
		}
		newParentID = parentID
		updates["parent_id"] = parentID
	}
	if req.MenuType != nil {
		updates["menu_type"] = menuTypeToInt(*req.MenuType)
	}
	if req.RoutePath != nil {
		updates["route_path"] = *req.RoutePath
	}
	if req.Sort != nil {
		updates["sort"] = *req.Sort
	}
	if req.Status != "" {
		updates["status"] = statusToInt(req.Status)
	}
	if req.Component != nil {
		updates["component"] = *req.Component
	}
	if req.Icon != nil {
		updates["icon"] = *req.Icon
	}
	if req.HideInMenu != nil {
		updates["hide_in_menu"] = boolToInt(*req.HideInMenu)
	}

	if len(updates) == 0 {
		return appErrors.ErrMissingParam
	}

	// 类型规则：目录(1)组件路径为空；子菜单(有父级)图标为空，一级菜单保留图标
	menuType := menu.MenuType
	if req.MenuType != nil {
		menuType = menuTypeToInt(*req.MenuType)
	}
	if menuType == 1 {
		updates["component"] = ""
	}
	if newParentID != 0 {
		updates["icon"] = ""
	}

	return s.menuRepo.Update(ctx, id, updates)
}

// DeleteMenu 删除菜单（级联删除全部子孙菜单并清理角色关联）
func (s *menuService) DeleteMenu(ctx context.Context, id uint) error {
	// 校验菜单存在，并记录父菜单
	menu, err := s.menuRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	parentID := menu.ParentID

	// 广度优先收集全部子孙 id
	ids := []uint{id}
	frontier := []uint{id}
	for len(frontier) > 0 {
		childIDs, err := s.menuRepo.FindChildrenIDs(ctx, frontier)
		if err != nil {
			return appErrors.ErrInternalError
		}
		if len(childIDs) == 0 {
			break
		}
		ids = append(ids, childIDs...)
		frontier = childIDs
	}

	if err := s.menuRepo.DeleteCascade(ctx, ids); err != nil {
		return err
	}

	// 类型联动：父菜单已无任何子菜单，类型改回"菜单"
	if parentID != 0 {
		if cnt, err := s.menuRepo.CountChildren(ctx, parentID); err == nil && cnt == 0 {
			_ = s.menuRepo.Update(ctx, parentID, map[string]interface{}{"menu_type": 2})
		}
	}
	return nil
}
