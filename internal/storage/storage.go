package storage

import (
	"context"
	"fmt"
	"io"

	"tuyou/pkg/config"
)

// Storage 统一文件存储接口。
// 由配置文件 storage.driver 决定具体实现（当前仅 oss，local 预留），
// 上传模块只依赖本接口，不感知存储后端。
type Storage interface {
	// Upload 将文件内容流式上传到指定对象键（objectKey），返回可访问的永久 URL。
	// objectKey 形如 "images/2026/08/08/abc123.jpg"，由调用方生成。
	// 使用 io.Reader 流式写入，避免大文件（视频 ≤100MB）整块加载进内存。
	Upload(ctx context.Context, objectKey string, r io.Reader, size int64, contentType string) (string, error)
}

// New 根据配置创建存储实现。
func New(cfg config.StorageConfig) (Storage, error) {
	switch cfg.Driver {
	case "oss":
		return newOSSStorage(cfg.OSS)
	default:
		return nil, fmt.Errorf("不支持的存储驱动: %s（当前仅支持 oss）", cfg.Driver)
	}
}
