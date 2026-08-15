package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"tuyou/pkg/config"

	ffmpeg "github.com/u2takey/ffmpeg-go"
)

// videoCoverTimeout ffmpeg 截帧超时（防止坏视频卡死上传请求）
const videoCoverTimeout = 30 * time.Second

// localStorage 本地磁盘存储实现（driver=local）。
// 文件写 {baseDir}/{objectKey}，访问 URL = {baseURL}/static/{objectKey}（静态路由在 app 层挂载）。
type localStorage struct {
	baseDir    string
	baseURL    string
	ffmpegPath string
}

// newLocalStorage 校验配置并创建本地存储。
func newLocalStorage(cfg config.LocalStorageConfig) (Storage, error) {
	if cfg.BaseDir == "" || cfg.BaseURL == "" {
		return nil, fmt.Errorf("local 配置不完整：base_dir/base_url 均必填")
	}
	if err := os.MkdirAll(cfg.BaseDir, 0o755); err != nil {
		return nil, fmt.Errorf("创建本地存储目录失败: %w", err)
	}
	return &localStorage{
		baseDir:    cfg.BaseDir,
		baseURL:    strings.TrimRight(cfg.BaseURL, "/"),
		ffmpegPath: cfg.FFmpegPath,
	}, nil
}

// Upload 流式写文件到本地磁盘（自动建目录），返回可访问 URL。
// contentType 忽略：静态文件按扩展名由浏览器识别。
func (s *localStorage) Upload(_ context.Context, objectKey string, r io.Reader, _ int64, _ string) (string, error) {
	fullPath := filepath.Join(s.baseDir, filepath.FromSlash(objectKey))
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return "", err
	}
	dst, err := os.Create(fullPath)
	if err != nil {
		return "", err
	}
	defer dst.Close()
	if _, err := io.Copy(dst, r); err != nil {
		return "", err
	}
	return s.objectURL(objectKey), nil
}

// objectURL 拼接本地文件访问 URL
func (s *localStorage) objectURL(objectKey string) string {
	return s.baseURL + "/static/" + objectKey
}

// VideoCover 对已落盘的视频文件用 ffmpeg 截第一帧，另存封面文件，返回真实封面 URL。
// 截帧失败返回 err（上传服务 best-effort 处理，不阻断视频上传）。
func (s *localStorage) VideoCover(_ context.Context, videoKey string) (string, error) {
	videoPath := filepath.Join(s.baseDir, filepath.FromSlash(videoKey))
	coverKey := coverKeyFromVideo(videoKey)
	coverPath := filepath.Join(s.baseDir, filepath.FromSlash(coverKey))
	if err := os.MkdirAll(filepath.Dir(coverPath), 0o755); err != nil {
		return "", err
	}

	cmd := ffmpeg.Input(videoPath).
		Output(coverPath, ffmpeg.KwArgs{"vframes": 1, "f": "image2"}).
		OverWriteOutput().
		Silent(true) // 关闭 ffmpeg-go 的 compiled command 回显（默认打到 stderr，IDE 里显红）
	if s.ffmpegPath != "" {
		cmd = cmd.SetFfmpegPath(s.ffmpegPath)
	}
	// 带超时，坏视频不卡死上传请求
	cmd = cmd.WithTimeout(videoCoverTimeout)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("ffmpeg 截帧失败 videoKey=%s err=%w", videoKey, err)
	}
	return s.objectURL(coverKey), nil
}

// coverKeyFromVideo 由 videoKey 推导封面 key：video/2026/08/14/xxx.mp4 → covers/2026/08/14/xxx.jpg
func coverKeyFromVideo(videoKey string) string {
	ext := filepath.Ext(videoKey)                               // ".mp4"
	base := strings.TrimSuffix(filepath.ToSlash(videoKey), ext) // "video/2026/08/14/xxx"
	return strings.Replace(base, "video/", "covers/", 1) + ".jpg"
}
