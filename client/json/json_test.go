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

func TestJsonObjectWithSecondDepth(t *testing.T) {
	reader := bytes.NewBufferString("")
	AddSession(gid.Get(), reader)

	Map("k1", "v1")
	Map("k2", func() {
		Map(
			"k2.3", "v2.3",
		)
	})
	Map("k3", true)

	CloseSession(gid.Get())

	assert.Equal(t, `{"k1":"v1","k2":{"k2.3":"v2.3"},"k3":true}`, reader.String())
}