package cloudstorage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestETAG(t *testing.T) {
	assert.Equal(t, "hello", CleanETag("hello"))
	assert.Equal(t, "hello", CleanETag(`"hello"`))
	assert.Equal(t, "hello", CleanETag(`\"hello\"`))
	assert.Equal(t, "hello", CleanETag("\"hello\""))
}
func TestContentType(t *testing.T) {
	assert.Equal(t, "text/csv; charset=utf-8", ContentType("data.csv"))
	assert.Equal(t, "application/json", ContentType("data.json"))
	assert.Equal(t, "application/octet-stream", ContentType("data.unknown"))
}
