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

func TestJsonObjectWithSecondDepthMultipleElements(t *testing.T) {
	reader := bytes.NewBufferString("")
	AddSession(gid.Get(), reader)

	Map("k1", "v1")
	Map("k2", func() {
		Map(
			"k2.3", "v2.3",
		)
		Map(
			"k2.4","v2.4",
		)
		Map(
			"anotherDepth", func() {
				Map("k4", "reached")
		})
	})
	Map("k3", true)

	CloseSession(gid.Get())

	assert.Equal(t, `{"k1":"v1","k2":{"k2.3":"v2.3","k2.4":"v2.4","anotherDepth":{"k4":"reached"}},"k3":true}`, reader.String())
}

func TestJsonList(t *testing.T) {
	reader := bytes.NewBufferString("")
	AddSession(gid.Get(), reader)

	List(
		Element(func() {
			Map("key1", "val1")
		}),
		Element(func() { Map("key2", true) }),
		Element(false),
	)

	CloseSession(gid.Get())

	assert.Equal(t, `[{"key1":"val1"},{"key2":true},false]`, reader.String())
}

func TestMapWithList(t *testing.T) {
	reader := bytes.NewBufferString("")
	AddSession(gid.Get(), reader)

	Map(
		"list", List(1,2,3),
	)
	Map(
		"element", true,
	)

	CloseSession(gid.Get())

	assert.Equal(t, `{"list":[1,2,3],"element":true}`, reader.String())
}

func TestListWithinList(t *testing.T) {
	reader := bytes.NewBufferString("")
	AddSession(gid.Get(), reader)

	List(
		Element(
			List(1,2,3)),
		4,5,
	)

	CloseSession(gid.Get())

	assert.Equal(t, `[[1,2,3],4,5]`, reader.String())
}

func TestListAndAnotherList(t *testing.T) {
	reader := bytes.NewBufferString("")
	AddSession(gid.Get(), reader)

	List(
		1,2,3,
	)()
	List(
		4,5,
	)()

	CloseSession(gid.Get())

	assert.Equal(t, `[1,2,3,4,5]`, reader.String())
}