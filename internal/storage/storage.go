package storage

import (
	"context"
	"fmt"
	"io"

	"tuyou/pkg/config"
)

// Storage 统一文件存储接口。
// 由配置文件 storage.driver 决定具体实现（oss | local），
// 上传模块只依赖本接口，不感知存储后端。
type Storage interface {
	// Upload 将文件内容流式上传到指定对象键（objectKey），返回可访问的永久 URL。
	// objectKey 形如 "images/2026/08/08/abc123.jpg"，由调用方生成。
	// 使用 io.Reader 流式写入，避免大文件（视频 ≤300MB）整块加载进内存。
	Upload(ctx context.Context, objectKey string, r io.Reader, size int64, contentType string) (string, error)

	// VideoCover 生成视频封面 URL。
	//   - oss：返回 x-oss-process 懒加载截帧 URL（不读文件、不占存储）；
	//   - local：对已落盘的视频文件用 ffmpeg 截第一帧，另存封面文件，返回真实封面 URL。
	// 失败返回 err，调用方 best-effort 处理（日志 + 空封面，不阻断上传）。
	VideoCover(ctx context.Context, videoKey string) (string, error)
}

// New 根据配置创建存储实现。
func New(cfg config.StorageConfig) (Storage, error) {
	switch cfg.Driver {
	case "oss":
		return newOSSStorage(cfg.OSS)
	case "local":
		return newLocalStorage(cfg.Local)
	default:
		return nil, fmt.Errorf("不支持的存储驱动: %s（当前仅支持 oss / local）", cfg.Driver)
	}
}
