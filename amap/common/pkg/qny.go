package pkg

import (
	"common/config"
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-uuid"
	"github.com/qiniu/go-sdk/v7/storagev2/credentials"
	"github.com/qiniu/go-sdk/v7/storagev2/http_client"
	"github.com/qiniu/go-sdk/v7/storagev2/uploader"
)

func validateImageFile(file *multipart.FileHeader) error {
	if file == nil {
		return errors.New("未上传文件")
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".png" && ext != ".jpg" && ext != ".jpeg" && ext != ".webp" {
		return errors.New("只允许上传 .png、.jpg、.jpeg、.webp 格式图片")
	}
	if file.Size > 5*1024*1024 {
		return errors.New("图片大小不能超过5M")
	}
	return nil
}

// QiNiuYun 上传图片到七牛云，返回可访问 URL
func QiNiuYun(file *multipart.FileHeader) (string, error) {
	if err := validateImageFile(file); err != nil {
		return "", err
	}
	data := config.DataConfig.Qny
	if data.AccessKey == "" || data.SecretKey == "" || data.Bucket == "" {
		return "", errors.New("七牛云配置不完整")
	}
	if data.Domain == "" {
		return "", errors.New("七牛云域名未配置")
	}
	id, err := uuid.GenerateUUID()
	if err != nil {
		return "", errors.New("生成文件名失败")
	}
	key := fmt.Sprintf("img/%s%s", id, filepath.Ext(file.Filename))
	mac := credentials.NewCredentials(data.AccessKey, data.SecretKey)
	uploadManager := uploader.NewUploadManager(&uploader.UploadManagerOptions{
		Options: http_client.Options{
			Credentials: mac,
		},
	})
	reader, err := file.Open()
	if err != nil {
		return "", errors.New("读取上传文件失败")
	}
	defer reader.Close()
	err = uploadManager.UploadReader(context.Background(), reader, &uploader.ObjectOptions{
		BucketName: data.Bucket,
		ObjectName: &key,
		FileName:   file.Filename,
	}, nil)
	if err != nil {
		fmt.Println("七牛云上传失败:", err)
		return "", errors.New("图片上传失败")
	}
	domain := strings.TrimRight(data.Domain, "/")
	return domain + "/" + key, nil
}
