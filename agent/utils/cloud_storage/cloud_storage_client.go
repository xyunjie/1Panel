package cloud_storage

import (
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/utils/cloud_storage/client"
)

type CloudStorageClient interface {
	ListBuckets() ([]interface{}, error)
	ListObjects(prefix string) ([]string, error)
	Exist(path string) (bool, error)
	Delete(path string) (bool, error)
	Upload(src, target string) (bool, error)
	Download(src, target string) (bool, error)

	Size(path string) (int64, error)
}

func NewCloudStorageClient(backupType string, vars map[string]interface{}) (CloudStorageClient, error) {
	switch backupType {
	case constant.Local:
		return client.NewLocalClient(vars)
	case constant.S3:
		return client.NewS3Client(vars)
	case constant.Sftp:
		return client.NewSftpClient(vars)
	case constant.MinIo:
		return client.NewMinIoClient(vars)
	default:
		return nil, buserr.WithName("ErrNotSupportType", backupType)
	}
}
