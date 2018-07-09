package storage

type StorageClient interface {
	Download(filePath string) (Version, error)
	Upload(filePath string) (Version, error)
	Version() (Version, error)
	GetAllNewerVersions(watermark Version) ([]Version, error)
	DeleteBucket() error // test cleanup only
}

func NewStorageClient(gcpServiceAccountKey, objectName, bucketName string) (StorageClient, error) {
	return NewGCSStorage(gcpServiceAccountKey, objectName, bucketName)
}
