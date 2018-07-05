package cloudstorage_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lytics/cloudstorage"
)

func TestRegistry(t *testing.T) {
	cloudstorage.Register("teststore", fakeProvider)
	paniced := didPanic(func() {
		cloudstorage.Register("teststore", fakeProvider)
	})
	assert.True(t, paniced)
}
func didPanic(f func()) (dp bool) {
	defer func() {
		if r := recover(); r != nil {
			dp = true
		}
	}()
	f()
	return dp
}

func fakeProvider(conf *cloudstorage.Config) (cloudstorage.Store, error) {
	return nil, fmt.Errorf("Not Implemented")
}
