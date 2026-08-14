package upload

import (
	"context"
	"mime/multipart"

	dtoUpload "tuyou/internal/model/dto/response/upload"
)

// UploadService 文件上传服务接口
type UploadService interface {
	// UploadImages 批量上传图片（1~5 张，单张 ≤5MB），返回 URL 列表
	UploadImages(ctx context.Context, files []*multipart.FileHeader) (*dtoUpload.UploadImagesResponse, error)

	// UploadVideo 上传单个视频（≤300MB），返回视频 URL 和封面 URL
	UploadVideo(ctx context.Context, file *multipart.FileHeader) (*dtoUpload.UploadVideoResponse, error)
}
