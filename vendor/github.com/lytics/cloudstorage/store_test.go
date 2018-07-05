package cloudstorage_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/lytics/cloudstorage"
	"github.com/lytics/cloudstorage/localfs"
	"github.com/lytics/cloudstorage/testutils"
)

func TestAll(t *testing.T) {
	localFsConf := &cloudstorage.Config{
		Type:       localfs.StoreType,
		AuthMethod: localfs.AuthFileSystem,
		LocalFS:    "/tmp/mockcloud",
		TmpDir:     "/tmp/localcache",
	}

	store, err := cloudstorage.NewStore(localFsConf)
	if err != nil {
		t.Fatalf("Could not create store: config=%+v  err=%v", localFsConf, err)
		return
	}
	testutils.RunTests(t, store)
	// verify cleanup
	cloudstorage.CleanupCacheFiles(time.Minute*1, localFsConf.TmpDir)
}

func TestStore(t *testing.T) {
	invalidConf := &cloudstorage.Config{}

	store, err := cloudstorage.NewStore(invalidConf)
	assert.NotEqual(t, nil, err)
	assert.Equal(t, nil, store)

	missingStoreConf := &cloudstorage.Config{
		Type: "non-existent-store",
	}

	store, err = cloudstorage.NewStore(missingStoreConf)
	assert.NotEqual(t, nil, err)
	assert.Equal(t, nil, store)

	// test missing temp dir, assign local temp
	localFsConf := &cloudstorage.Config{
		Type:       localfs.StoreType,
		AuthMethod: localfs.AuthFileSystem,
		LocalFS:    "/tmp/mockcloud",
	}

	store, err = cloudstorage.NewStore(localFsConf)
	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, store)
}

func TestJwtConf(t *testing.T) {
	configInput := `
	{
		"JwtConf": {
			"type": "service_account",
			"project_id": "testing",
			"private_key_id": "abcdefg",
			"private_key": "aGVsbG8td29ybGQ=",
			"client_email": "testing@testing.iam.gserviceaccount.com",
			"client_id": "117058426251532209964",
			"scopes": [
				"https://www.googleapis.com/auth/devstorage.read_write"
			]
		}
	}`

	// v := base64.StdEncoding.EncodeToString([]byte("hello-world"))
	// t.Logf("b64  %q", v)
	conf := &cloudstorage.Config{}
	err := json.Unmarshal([]byte(configInput), conf)
	assert.Equal(t, nil, err)
	conf.JwtConf.PrivateKey = "------helo-------\naGVsbG8td29ybGQ=\n-----------------end--------"
	assert.NotEqual(t, nil, conf.JwtConf)
	assert.Equal(t, nil, conf.JwtConf.Validate())
	assert.Equal(t, "aGVsbG8td29ybGQ=", conf.JwtConf.PrivateKey)
	assert.Equal(t, "service_account", conf.JwtConf.Type)

	// note on this one the "keytype" & "private_keybase64"
	configInput = `
	{
		"JwtConf": {
			"keytype": "service_account",
			"project_id": "testing",
			"private_key_id": "abcdefg",
			"private_keybase64": "aGVsbG8td29ybGQ=",
			"client_email": "testing@testing.iam.gserviceaccount.com",
			"client_id": "117058426251532209964",
			"scopes": [
				"https://www.googleapis.com/auth/devstorage.read_write"
			]
		}
	}`
	conf = &cloudstorage.Config{}
	err = json.Unmarshal([]byte(configInput), conf)
	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, conf.JwtConf)
	assert.Equal(t, nil, conf.JwtConf.Validate())
	assert.Equal(t, "aGVsbG8td29ybGQ=", conf.JwtConf.PrivateKey)
	assert.Equal(t, "service_account", conf.JwtConf.Type)
}
