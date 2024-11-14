package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
)

type MinioService interface {
	MakeBucket(bucketName string) error
	NewPresignedUrl(bucketName string, objectName string, method string, expiresSec int64) (string, error)
	StatObject(bucketName string, objectName string) (minio.ObjectInfo, error)
	GetObject(bucketName string, objectName string) ([]byte, error)
	UpdateObject(bucketName string, objectName string, data []byte) error
	CopyObject(srcBucketName string, srcObjectName string, dstBucketName string, dstObjectName string) error
	RemoveObject(bucketName string, objectName string) error
}

type minioService struct {
	ctx    *context.Context
	client *minio.Client
}

func NewMinioService(ctx *context.Context, cli *minio.Client) MinioService {
	return &minioService{ctx: ctx, client: cli}
}

func (ms *minioService) MakeBucket(bucketName string) error {
	if err := ms.client.MakeBucket(*ms.ctx, bucketName, minio.MakeBucketOptions{}); err != nil {
		return err
	}

	policy := fmt.Sprintf(`{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Principal": {
					"AWS": [
						"*"
					]
				},
				"Action": [
					"s3:GetBucketLocation",
					"s3:ListBucket"
				],
				"Resource": [
					"arn:aws:s3:::%s"
				]
			},
			{
				"Effect": "Allow",
				"Principal": {
					"AWS": [
						"*"
					]
				},
				"Action": [
					"s3:GetObject"
				],
				"Resource": [
					"arn:aws:s3:::%s/*"
				]
			}
		]
	}`, bucketName, bucketName)

	if err := ms.client.SetBucketPolicy(*ms.ctx, bucketName, policy); err != nil {
		return err
	}

	return nil
}

func (ms *minioService) NewPresignedUrl(bucketName string, objectName string, method string, expiresSec int64) (string, error) {
	var url *url.URL
	var err error
	expires := time.Duration(expiresSec) * time.Second

	switch method {
	case "GET":
		url, err = ms.client.PresignedGetObject(*ms.ctx, bucketName, objectName, expires, nil)
	case "PUT":
		url, err = ms.client.PresignedPutObject(*ms.ctx, bucketName, objectName, expires)
	default:
		return "", fmt.Errorf("invalid method")
	}

	if err != nil {
		return "", err
	}

	return url.String(), nil
}

func (ms *minioService) StatObject(bucketName string, objectName string) (minio.ObjectInfo, error) {
	return ms.client.StatObject(*ms.ctx, bucketName, objectName, minio.StatObjectOptions{})
}

func (ms *minioService) GetObject(bucketName string, objectName string) ([]byte, error) {
	fmt.Println("Getting object", objectName, "from bucket", bucketName)

	obj, err := ms.client.GetObject(*ms.ctx, bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		fmt.Println("Failed to get object", objectName, "from bucket", bucketName, ":", err)
		return nil, err
	}

	defer obj.Close()

	data, err := io.ReadAll(obj)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (ms *minioService) UpdateObject(bucketName string, objectName string, data []byte) error {
	_, err := ms.client.PutObject(*ms.ctx, bucketName, objectName, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{})
	return err
}

func (ms *minioService) CopyObject(srcBucketName string, srcObjectName string, dstBucketName string, dstObjectName string) error {
	_, err := ms.client.CopyObject(*ms.ctx, minio.CopyDestOptions{
		Bucket: dstBucketName,
		Object: dstObjectName,
	}, minio.CopySrcOptions{
		Bucket: srcBucketName,
		Object: srcObjectName,
	})

	return err
}

func (ms *minioService) RemoveObject(bucketName string, objectName string) error {
	return ms.client.RemoveObject(*ms.ctx, bucketName, objectName, minio.RemoveObjectOptions{
		ForceDelete: true,
	})
}
