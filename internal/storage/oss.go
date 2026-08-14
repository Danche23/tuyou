package storage

import (
	"context"
	"fmt"
	"io"

	"tuyou/pkg/config"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
)

// videoSnapshotParam OSS 视频截帧参数：t=0 取封面帧，输出 jpg（访问封面 URL 时即时生成，不占存储）
const videoSnapshotParam = "x-oss-process=video/snapshot,t_0,f_jpg"

// ossStorage 阿里云 OSS 存储实现（SDK v2，与 blog 项目一致）。
// 对象写入 OSS Bucket，访问 URL 固定用默认域名 https://{bucket}.oss-{region}.aliyuncs.com/{key}
type ossStorage struct {
	client *oss.Client
	cfg    config.OSSStorageConfig
}

// objectURL 拼接对象访问 URL（固定默认域名）
func (s *ossStorage) objectURL(objectKey string) string {
	return fmt.Sprintf("https://%s.oss-%s.aliyuncs.com/%s", s.cfg.Bucket, s.cfg.Endpoint, objectKey)
}

// newOSSStorage 校验配置并创建 OSS 客户端。
func newOSSStorage(cfg config.OSSStorageConfig) (Storage, error) {
	if cfg.AccessKeyID == "" || cfg.AccessKeySecret == "" || cfg.Bucket == "" || cfg.Endpoint == "" {
		return nil, fmt.Errorf("oss 配置不完整：access_key_id/access_key_secret/bucket/endpoint 均必填")
	}
	ossCfg := oss.LoadDefaultConfig().
		WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.AccessKeySecret, "")).
		WithRegion(cfg.Endpoint)
	return &ossStorage{client: oss.NewClient(ossCfg), cfg: cfg}, nil
}

// Upload 流式上传对象到 OSS。
// contentType 不传，OSS 会按 objectKey 扩展名自动识别 Content-Type。
func (s *ossStorage) Upload(ctx context.Context, objectKey string, r io.Reader, _ int64, _ string) (string, error) {
	request := &oss.PutObjectRequest{
		Bucket: oss.Ptr(s.cfg.Bucket),
		Key:    oss.Ptr(objectKey),
		Body:   r,
	}
	if _, err := s.client.PutObject(ctx, request); err != nil {
		return "", err
	}
	return s.objectURL(objectKey), nil
}

// VideoCover 返回视频封面 URL（懒加载：访问时 OSS 即时截帧，不读文件、不占存储）
func (s *ossStorage) VideoCover(_ context.Context, videoKey string) (string, error) {
	return s.objectURL(videoKey) + "?" + videoSnapshotParam, nil
}
