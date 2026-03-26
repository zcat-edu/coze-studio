/*
 * Copyright 2025 coze-dev Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package impl

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/coze-dev/coze-studio/backend/infra/imagex"
	"github.com/coze-dev/coze-studio/backend/infra/storage"
	"github.com/coze-dev/coze-studio/backend/infra/storage/impl/minio"
	"github.com/coze-dev/coze-studio/backend/infra/storage/impl/s3"
	"github.com/coze-dev/coze-studio/backend/infra/storage/impl/tos"
	"github.com/coze-dev/coze-studio/backend/pkg/envkey"
	"github.com/coze-dev/coze-studio/backend/types/consts"
)

type Storage = storage.Storage

func New(ctx context.Context) (Storage, error) {
	storageType := os.Getenv(consts.StorageType)
	switch storageType {
	case "minio":
		return minio.New(
			ctx,
			os.Getenv(consts.MinIOEndpoint),
			os.Getenv(consts.MinIOAK),
			os.Getenv(consts.MinIOSK),
			os.Getenv(consts.StorageBucket),
			envkey.GetBoolD("MINIO_USE_SSL", false),
		)
	case "tos":
		return tos.New(
			ctx,
			os.Getenv(consts.TOSAccessKey),
			os.Getenv(consts.TOSSecretKey),
			os.Getenv(consts.StorageBucket),
			os.Getenv(consts.TOSEndpoint),
			os.Getenv(consts.TOSRegion),
		)
	case "s3":
		return s3.New(
			ctx,
			os.Getenv(consts.S3AccessKey),
			os.Getenv(consts.S3SecretKey),
			os.Getenv(consts.StorageBucket),
			os.Getenv(consts.S3Endpoint),
			os.Getenv(consts.S3Region),
		)
	default:
		// For local or Windows development, return nil
		return nil, nil
	}

	return nil, fmt.Errorf("unknown storage type: %s", storageType)
}

// mockStorage is a simple mock implementation of Storage for local development
type mockStorage struct{}

func (m *mockStorage) PutObject(ctx context.Context, objectKey string, content []byte, opts ...func(interface{})) error {
	return nil
}

func (m *mockStorage) PutObjectWithReader(ctx context.Context, objectKey string, content interface{}, opts ...func(interface{})) error {
	return nil
}

func (m *mockStorage) GetObject(ctx context.Context, objectKey string) ([]byte, error) {
	return []byte{}, nil
}

func (m *mockStorage) DeleteObject(ctx context.Context, objectKey string) error {
	return nil
}

func (m *mockStorage) GetObjectUrl(ctx context.Context, objectKey string, opts ...func(interface{})) (string, error) {
	return "http://localhost:8888/static/" + objectKey, nil
}

func (m *mockStorage) HeadObject(ctx context.Context, objectKey string, opts ...func(interface{})) (interface{}, error) {
	return nil, nil
}

func (m *mockStorage) ListAllObjects(ctx context.Context, prefix string, opts ...func(interface{})) ([]interface{}, error) {
	return []interface{}{}, nil
}

func (m *mockStorage) ListObjectsPaginated(ctx context.Context, input interface{}, opts ...func(interface{})) (interface{}, error) {
	return nil, nil
}

func NewImagex(ctx context.Context) (imagex.ImageX, error) {
	storageType := os.Getenv(consts.StorageType)
	switch storageType {
	case "minio":
		return minio.NewStorageImagex(
			ctx,
			os.Getenv(consts.MinIOEndpoint),
			os.Getenv(consts.MinIOAK),
			os.Getenv(consts.MinIOSK),
			os.Getenv(consts.StorageBucket),
			envkey.GetBoolD("MINIO_USE_SSL", false),
		)
	case "tos":
		return tos.NewStorageImagex(
			ctx,
			os.Getenv(consts.TOSAccessKey),
			os.Getenv(consts.TOSSecretKey),
			os.Getenv(consts.StorageBucket),
			os.Getenv(consts.TOSEndpoint),
			os.Getenv(consts.TOSRegion),
		)
	case "s3":
		return s3.NewStorageImagex(
			ctx,
			os.Getenv(consts.S3AccessKey),
			os.Getenv(consts.S3SecretKey),
			os.Getenv(consts.StorageBucket),
			os.Getenv(consts.S3Endpoint),
			os.Getenv(consts.S3Region),
		)
	default:
		// For local or Windows development, return nil
		return nil, nil
	}
	return nil, fmt.Errorf("unknown storage type: %s", storageType)
}

// mockImageX is a simple mock implementation of ImageX for local development
type mockImageX struct{}

func (m *mockImageX) GetUploadAuth(ctx context.Context, opt ...interface{}) (interface{}, error) {
	return nil, nil
}

func (m *mockImageX) GetUploadAuthWithExpire(ctx context.Context, expire time.Duration, opt ...interface{}) (interface{}, error) {
	return nil, nil
}

func (m *mockImageX) GetResourceURL(ctx context.Context, uri string, opts ...interface{}) (interface{}, error) {
	return &struct {
		CompactURL string `json:"CompactURL"`
		URL        string `json:"URL"`
	}{
		CompactURL: "http://localhost:8888/static/" + uri,
		URL:        "http://localhost:8888/static/" + uri,
	}, nil
}

func (m *mockImageX) Upload(ctx context.Context, data []byte, opts ...interface{}) (interface{}, error) {
	return &struct {
		Result    interface{} `json:"Results"`
		RequestId string      `json:"RequestId"`
		FileInfo  interface{} `json:"PluginResult"`
	}{
		Result: &struct {
			Uri       string `json:"Uri"`
			UriStatus int    `json:"UriStatus"`
		}{
			Uri:       "image.jpg",
			UriStatus: 2000,
		},
		RequestId: "mock-request-id",
		FileInfo: &struct {
			Name        string `json:"FileName"`
			Uri         string `json:"ImageUri"`
			ImageWidth  int    `json:"ImageWidth"`
			ImageHeight int    `json:"ImageHeight"`
			Md5         string `json:"ImageMd5"`
			ImageFormat string `json:"ImageFormat"`
			ImageSize   int    `json:"ImageSize"`
			FrameCnt    int    `json:"FrameCnt"`
			Duration    int    `json:"Duration"`
		}{
			Name:        "image.jpg",
			Uri:         "image.jpg",
			ImageWidth:  100,
			ImageHeight: 100,
			Md5:         "mock-md5",
			ImageFormat: "jpg",
			ImageSize:   len(data),
			FrameCnt:    1,
			Duration:    0,
		},
	}, nil
}

func (m *mockImageX) GetServerID() string {
	return "mock-server-id"
}

func (m *mockImageX) GetUploadHost(ctx context.Context) string {
	return "http://localhost:8888/upload"
}
