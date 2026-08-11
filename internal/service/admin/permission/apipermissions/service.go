package apipermissions

import (
	"context"
	"errors"
	"math"
	"strconv"

	request "tuyou/internal/model/dto/request/admin/permission/apipermissions"
	response "tuyou/internal/model/dto/response/admin/permission/apipermissions"
	"tuyou/internal/model/entity"
	"tuyou/internal/repository/admin/permission/apipermissions"
	appErrors "tuyou/pkg/errors"
)

type apiPermService struct {
	apiPermRepo apipermissions.ApiPermRepository
}

func NewApiPermService(apiPermRepo apipermissions.ApiPermRepository) ApiPermService {
	return &apiPermService{apiPermRepo: apiPermRepo}
}

// statusToInt 状态中文转 int：启用=1，禁用=0
func statusToInt(s string) int {
	if s == "禁用" {
		return 0
	}
	return 1
}

// ListApiPerms 查询接口权限列表
func (s *apiPermService) ListApiPerms(ctx context.Context, req *request.ListApiPermsReq) (*response.ListApiPermResp, error) {
	filters := map[string]interface{}{}
	if req.PermName != "" {
		filters["perm_name"] = req.PermName
	}
	if req.Module != "" {
		filters["module"] = req.Module
	}
	if req.Status != "" {
		filters["status"] = statusToInt(req.Status)
	}

	total, err := s.apiPermRepo.Count(ctx, filters)
	if err != nil {
		return nil, appErrors.ErrInternalError
	}

	list, err := s.apiPermRepo.List(ctx, filters, int(req.Page), int(req.Size))
	if err != nil {
		return nil, appErrors.ErrInternalError
	}

	return &response.ListApiPermResp{
		Total:     int(total),
		Page:      int(req.Page),
		Size:      int(req.Size),
		TotalPage: int(math.Ceil(float64(total) / float64(req.Size))),
		List:      list,
	}, nil
}

// GetModules 获取所有业务模块
func (s *apiPermService) GetModules(ctx context.Context) ([]response.ModuleItem, error) {
	modules, err := s.apiPermRepo.GetModules(ctx)
	if err != nil {
		return nil, appErrors.ErrInternalError
	}
	items := make([]response.ModuleItem, 0, len(modules))
	for i, name := range modules {
		items = append(items, response.ModuleItem{
			ID:   strconv.Itoa(i + 1),
			Name: name,
		})
	}
	return items, nil
}

// InsertApiPerm 新增接口权限
func (s *apiPermService) InsertApiPerm(ctx context.Context, req *request.InsertApiPermReq) (*response.InsertApiPermResp, error) {
	// 权限标识唯一性校验
	exist, err := s.apiPermRepo.FindByPermKey(ctx, req.PermKey)
	if err != nil && !errors.Is(err, appErrors.ErrNotFound) {
		return nil, err
	}
	if exist != nil {
		return nil, appErrors.New(appErrors.CodeResourceAlreadyExists, "权限标识已存在")
	}

	perm := &entity.SysPermission{
		PermName:      req.PermName,
		Module:        req.Module,
		PermKey:       req.PermKey,
		RequestMethod: req.RequestMethod,
		APIPath:       req.APIPath,
		Status:        statusToInt(req.Status),
	}
	if err := s.apiPermRepo.Insert(ctx, perm); err != nil {
		return nil, err
	}

	return &response.InsertApiPermResp{
		ID:       strconv.Itoa(int(perm.ID)),
		PermName: perm.PermName,
	}, nil
}

// UpdateApiPerm 编辑接口权限
func (s *apiPermService) UpdateApiPerm(ctx context.Context, id uint, req *request.UpdateApiPermReq) error {
	// 校验权限存在
	if _, err := s.apiPermRepo.FindByID(ctx, id); err != nil {
		return err
	}

	// 权限标识唯一性校验（排除自身）
	exist, err := s.apiPermRepo.FindByPermKey(ctx, req.PermKey)
	if err != nil && !errors.Is(err, appErrors.ErrNotFound) {
		return err
	}
	if exist != nil && exist.ID != id {
		return appErrors.New(appErrors.CodeResourceAlreadyExists, "权限标识已存在")
	}

	updates := map[string]interface{}{
		"perm_name":      req.PermName,
		"module":         req.Module,
		"perm_key":       req.PermKey,
		"request_method": req.RequestMethod,
		"api_path":       req.APIPath,
		"status":         statusToInt(req.Status),
	}

	return s.apiPermRepo.Update(ctx, id, updates)
}

// DeleteApiPerm 删除接口权限（级联清理角色-权限关联）
func (s *apiPermService) DeleteApiPerm(ctx context.Context, id uint) error {
	// 校验权限存在
	if _, err := s.apiPermRepo.FindByID(ctx, id); err != nil {
		return err
	}

	return s.apiPermRepo.DeleteCascade(ctx, id)
}
