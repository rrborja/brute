package meta

import (
	"fmt"
	"github.com/rrborja/brute/client"
)

type HttpEquiv string

const (
	ContentType_ HttpEquiv = "content-type"
	DefaultStyle HttpEquiv = "default-style"
	Refresh HttpEquiv = "refresh"
)

type MetaName string

const (
	ApplicationName MetaName = "application-name"
	Author MetaName = "author"
	Description MetaName = "description"
	Generator MetaName = "generator"
	Keywords MetaName = "keywords"
	Viewport MetaName = "viewport"
)

type ContentType struct {
	Value client.Mime
}

func (contentType ContentType) Render() [2]string {
	return [2]string {`http-equiv="content-type"`, fmt.Sprintf(`content="%s"`, contentType.Value)}
}

type MetaInfo struct {
	Author string
	Description string
	Keywords []string
}