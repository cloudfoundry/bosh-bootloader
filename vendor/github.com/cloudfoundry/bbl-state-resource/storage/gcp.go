package storage

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"

	gcs "cloud.google.com/go/storage"
	"github.com/mholt/archiver"
	oauthgoogle "golang.org/x/oauth2/google"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// untested gcs api instantiation
type objectHandleWrapper struct {
	objectHandle *gcs.ObjectHandle
}

func (o objectHandleWrapper) Version() (Version, error) {
	r, err := o.objectHandle.Attrs(context.Background())
	if err == gcs.ErrObjectNotExist {
		return Version{}, ObjectNotFoundError
	}
	if err != nil {
		return Version{}, err
	}

	return Version{Name: r.Name, Ref: hex.EncodeToString(r.MD5), Updated: r.Updated}, nil
}

func (o objectHandleWrapper) NewReader() (io.ReadCloser, error) {
	r, err := o.objectHandle.NewReader(context.Background())
	if err == gcs.ErrObjectNotExist {
		return nil, ObjectNotFoundError
	}
	return r, err
}

func (o objectHandleWrapper) NewWriter() io.WriteCloser {
	return o.objectHandle.NewWriter(context.Background())
}

type bucketHandleWrapper struct {
	bucketHandle *gcs.BucketHandle
}

func (b bucketHandleWrapper) GetAllObjects() ([]Object, error) {
	objectIter := b.bucketHandle.Objects(context.Background(), nil)

	var objects []Object
	for {
		next, err := objectIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		handle := objectHandleWrapper{objectHandle: b.bucketHandle.Object(next.Name)}
		objects = append(objects, handle)
	}
	return objects, nil
}

func (b bucketHandleWrapper) Delete() error {
	objectIter := b.bucketHandle.Objects(context.Background(), nil)

	for {
		next, err := objectIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}
		err = b.bucketHandle.Object(next.Name).Delete(context.Background())
		if err != nil {
			return err
		}
	}
	return b.bucketHandle.Delete(context.Background())
}

func NewGCSStorage(serviceAccountKey, objectName, bucketName string) (Storage, error) {
	storageJwtConf, err := oauthgoogle.JWTConfigFromJSON([]byte(serviceAccountKey), gcs.ScopeReadWrite)
	if err != nil {
		return Storage{}, fmt.Errorf("failed to form JWT config from GCP storage account key: %s", err)
	}
	ctx := context.Background()
	tokenSource := storageJwtConf.TokenSource(ctx)

	storageClient, err := gcs.NewClient(ctx, option.WithTokenSource(tokenSource))
	if err != nil {
		return Storage{}, fmt.Errorf("failed to instantiate storageclient: %s", err)
	}

	p := struct {
		ProjectId string `json:"project_id"`
	}{}
	if err := json.Unmarshal([]byte(serviceAccountKey), &p); err != nil {
		return Storage{}, fmt.Errorf("Unmarshalling account key for project id: %s", err)
	}
	bucket := storageClient.Bucket(bucketName).UserProject(p.ProjectId)

	_, err = bucket.Attrs(ctx)
	if err != nil && err != gcs.ErrBucketNotExist {
		return Storage{}, fmt.Errorf("Failed to get bucket: %s", err)
	} else if err == gcs.ErrBucketNotExist {
		err = bucket.Create(ctx, p.ProjectId, nil)
	}
	if err != nil {
		return Storage{}, fmt.Errorf("Failed to create bucket: %s", err)
	}

	object := bucket.Object(objectName)

	return Storage{
		Name: objectName,
		Bucket: bucketHandleWrapper{
			bucketHandle: bucket,
		},
		Object: objectHandleWrapper{
			objectHandle: object,
		},
		Archiver: archiver.TarGz,
	}, nil
}
