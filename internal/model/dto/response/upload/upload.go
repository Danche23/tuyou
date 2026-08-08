package upload

// UploadImagesResponse 图片批量上传响应
type UploadImagesResponse struct {
	Urls []string `json:"urls"`
}

// UploadVideoResponse 视频上传响应
type UploadVideoResponse struct {
	VideoURL string `json:"videoUrl"`
	CoverURL string `json:"coverUrl"`
}
