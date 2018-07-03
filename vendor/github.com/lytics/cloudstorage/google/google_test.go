package google_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/lytics/cloudstorage"
	"github.com/lytics/cloudstorage/google"
	"github.com/lytics/cloudstorage/testutils"
)

/*

# to use Google Cloud Storage ensure you create a google-cloud jwt token

export CS_GCS_JWTKEY="{\"project_id\": \"lio-testing\", \"private_key_id\": \"

*/

var config = &cloudstorage.Config{
	Type:       google.StoreType,
	AuthMethod: google.AuthJWTKeySource,
	Project:    "tbd",
	Bucket:     "liotesting-int-tests-nl",
	TmpDir:     "/tmp/localcache/google",
}

func TestAll(t *testing.T) {
	jwtVal := os.Getenv("CS_GCS_JWTKEY")
	if jwtVal == "" {
		t.Skip("Not testing no CS_GCS_JWTKEY env var")
		return
	}
	jc := &cloudstorage.JwtConf{}

	if err := json.Unmarshal([]byte(jwtVal), jc); err != nil {
		t.Fatalf("Could not read CS_GCS_JWTKEY %v", err)
		return
	}
	if jc.ProjectID != "" {
		config.Project = jc.ProjectID
	}
	config.JwtConf = jc

	store, err := cloudstorage.NewStore(config)
	if err != nil {
		t.Fatalf("Could not create store: config=%+v  err=%v", config, err)
	}
	testutils.RunTests(t, store)
}
