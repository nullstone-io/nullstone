package nsaws

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type S3Uploader struct {
	// BucketName refers to the S3 bucket name that files will be uploaded
	BucketName string
	// ObjectDirectory allows placing files under a subdirectory within S3
	ObjectDirectory string
	// OnObjectUpload allows performing an operation (like logging) when a file completes upload
	OnObjectUpload func(objectKey string)
}

func (u *S3Uploader) UploadDir(ctx context.Context, cfg aws.Config, baseDir string, filepaths []string) error {
	uploader := manager.NewUploader(s3.NewFromConfig(cfg))
	for _, fp := range filepaths {
		if err := u.uploadOne(ctx, uploader, baseDir, fp); err != nil {
			return fmt.Errorf("error uploading %q: %w", fp, err)
		}
	}
	return nil
}

func (u *S3Uploader) uploadOne(ctx context.Context, uploader *manager.Uploader, baseDir, fp string) error {
	localFilepath := filepath.Join(baseDir, fp)
	file, err := os.Open(localFilepath)
	if err != nil {
		return fmt.Errorf("error opening local file %q: %w", localFilepath, err)
	}
	defer file.Close()
	objectKey := path.Join(u.ObjectDirectory, strings.Replace(fp, string(filepath.Separator), "/", -1))
	_, err = uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket: aws.String(u.BucketName),
		Key:    aws.String(objectKey),
		Body:   file,
	})
	if err != nil {
		return fmt.Errorf("error uploading file %q: %w", objectKey, err)
	}
	if u.OnObjectUpload != nil {
		u.OnObjectUpload(objectKey)
	}
	return nil
}
