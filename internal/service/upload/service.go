package upload

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
	"time"

	dtoUpload "tuyou/internal/model/dto/response/upload"
	"tuyou/internal/storage"
	"tuyou/pkg/errors"

	"github.com/gabriel-vasile/mimetype"
	"golang.org/x/image/draw"
)

const (
	maxImages     = 5
	maxImageBytes = 5 * 1024 * 1024   // 单张图片最大 5MB
	maxVideoBytes = 300 * 1024 * 1024 // 单个视频最大 300MB
	maxEdge       = 1920              // 图片最长边（px），超过则等比压缩
	jpegQuality   = 85

	// videoSnapshotParam OSS 视频截帧参数：t=0 取封面帧，输出 jpg（访问 URL 时即时生成）
	videoSnapshotParam = "x-oss-process=video/snapshot,t_0,f_jpg"
)

// allowedImages 允许的图片类型（按文件头检测）→ 扩展名
var allowedImages = map[string]string{
	"image/jpeg": "jpg",
	"image/png":  "png",
	"image/gif":  "gif",
	"image/webp": "webp",
}

// allowedVideos 允许的视频类型（按文件头检测）→ 扩展名
var allowedVideos = map[string]string{
	"video/mp4":        "mp4",
	"video/quicktime":  "mov",
	"video/x-msvideo":  "avi",
	"video/x-matroska": "mkv",
	"video/webm":       "webm",
}

// uploadService 文件上传服务实现
type uploadService struct {
	storage storage.Storage
}

// NewUploadService 创建上传服务
func NewUploadService(storage storage.Storage) UploadService {
	return &uploadService{storage: storage}
}

// UploadImages 批量上传图片（1~5 张，单张 ≤5MB），返回 URL 列表
func (s *uploadService) UploadImages(ctx context.Context, files []*multipart.FileHeader) (*dtoUpload.UploadImagesResponse, error) {
	if len(files) == 0 {
		return nil, errors.New(errors.CodeInvalidParam, "请选择要上传的图片")
	}
	if len(files) > maxImages {
		return nil, errors.New(errors.CodeInvalidParam, "图片最多上传 5 张")
	}

	urls := make([]string, 0, len(files))
	for _, fh := range files {
		if fh.Size > maxImageBytes {
			return nil, errors.New(errors.CodeInvalidParam, "单张图片不能超过 5MB")
		}

		src, err := fh.Open()
		if err != nil {
			return nil, errors.NewWithErr(errors.CodeInternalError, "读取图片失败", err)
		}
		data, err := io.ReadAll(src)
		_ = src.Close()
		if err != nil {
			return nil, errors.NewWithErr(errors.CodeInternalError, "读取图片失败", err)
		}

		// 按文件头（magic bytes）校验真实类型，不信任扩展名
		mime := mimetype.Detect(data).String()
		ext, ok := allowedImages[mime]
		if !ok {
			return nil, errors.New(errors.CodeInvalidParam, "仅支持 jpg/jpeg/png/gif/webp 格式")
		}

		// 压缩 + 去 EXIF（jpeg/png 重编码，gif/webp 保持原样）
		processed, err := processImage(data, mime)
		if err != nil {
			return nil, errors.New(errors.CodeInvalidParam, "图片文件无效")
		}

		objectKey := genObjectKey("images", ext)
		url, err := s.storage.Upload(ctx, objectKey, bytes.NewReader(processed), int64(len(processed)), mime)
		if err != nil {
			return nil, errors.NewWithErr(errors.CodeInternalError, "图片上传失败", err)
		}
		urls = append(urls, url)
	}
	return &dtoUpload.UploadImagesResponse{Urls: urls}, nil
}

// UploadVideo 上传单个视频（≤100MB），返回视频 URL 和封面 URL。
// 封面不单独存文件：由 OSS 视频截帧在访问封面 URL 时即时生成（videoURL + 截帧参数）。
func (s *uploadService) UploadVideo(ctx context.Context, file *multipart.FileHeader) (*dtoUpload.UploadVideoResponse, error) {
	if file == nil {
		return nil, errors.New(errors.CodeInvalidParam, "请选择要上传的视频")
	}
	if file.Size > maxVideoBytes {
		return nil, errors.New(errors.CodeInvalidParam, "单个视频不能超过 100MB")
	}

	src, err := file.Open()
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "读取视频失败", err)
	}
	defer src.Close()

	// 按文件头校验真实视频类型
	head, err := io.ReadAll(io.LimitReader(src, 4096))
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "读取视频失败", err)
	}
	mime := mimetype.Detect(head).String()
	ext, ok := allowedVideos[mime]
	if !ok {
		return nil, errors.New(errors.CodeInvalidParam, "仅支持 mp4/mov/avi/mkv/webm 格式")
	}

	// 回到文件头，整体流式上传（不整块加载进内存）
	if _, err := src.Seek(0, io.SeekStart); err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "读取视频失败", err)
	}

	objectKey := genObjectKey("video", ext)
	videoURL, err := s.storage.Upload(ctx, objectKey, src, file.Size, mime)
	if err != nil {
		return nil, errors.NewWithErr(errors.CodeInternalError, "视频上传失败", err)
	}

	coverURL := videoURL + "?" + videoSnapshotParam
	return &dtoUpload.UploadVideoResponse{VideoURL: videoURL, CoverURL: coverURL}, nil
}

// processImage 处理图片：jpeg/png 重编码（去 EXIF + 压缩），gif/webp 原样返回（避免丢动画/无法重编码）
func processImage(data []byte, mime string) ([]byte, error) {
	switch mime {
	case "image/jpeg", "image/png":
		return reencodeImage(data, mime)
	default:
		return data, nil
	}
}

// reencodeImage 解码后重编码，去除 EXIF 并按最长边 maxEdge 等比缩放
func reencodeImage(data []byte, mime string) ([]byte, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	resized := resizeIfNeeded(img)

	var buf bytes.Buffer
	if mime == "image/png" {
		if err := png.Encode(&buf, resized); err != nil {
			return nil, err
		}
	} else {
		if err := jpeg.Encode(&buf, resized, &jpeg.Options{Quality: jpegQuality}); err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

// resizeIfNeeded 最长边超过 maxEdge 时按比例缩放
func resizeIfNeeded(img image.Image) image.Image {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= maxEdge && h <= maxEdge {
		return img
	}
	ratio := float64(maxEdge) / float64(w)
	if float64(h)*ratio > float64(maxEdge) {
		ratio = float64(maxEdge) / float64(h)
	}
	nw, nh := int(float64(w)*ratio), int(float64(h)*ratio)
	dst := image.NewRGBA(image.Rect(0, 0, nw, nh))
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, b, draw.Over, nil)
	return dst
}

// genObjectKey 生成存储对象键：{dir}/{yyyy}/{MM}/{dd}/{yyyyMMdd_HHmmss}_{随机hex}.{ext}
// 文件名用时间戳 + 短随机后缀：可读、无中文、不撞名（秒级时间戳 + 6位hex兜底）
func genObjectKey(dir, ext string) string {
	now := time.Now()
	return fmt.Sprintf("%s/%d/%02d/%02d/%s_%s.%s",
		dir, now.Year(), now.Month(), now.Day(),
		now.Format("20060102_150405"), randHex(3), ext)
}

// randHex 生成 n 字节随机 hex 字符串（3字节 = 6位hex）
func randHex(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		// crypto/rand 失败概率极低，退回时间戳避免空 key
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}
