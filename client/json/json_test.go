package json

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"bytes"
	"github.com/silentred/gid"
)

func TestJsonObject(t *testing.T) {
	reader := bytes.NewBufferString("")
	AddSession(gid.Get(), reader)

	Map("key1", "value1")
	Map("key2", 2)
	Map("key3", false)

	CloseSession(gid.Get())

	assert.Equal(t, `{"key1":"value1","key2":2,"key3":false}`, reader.String())
}