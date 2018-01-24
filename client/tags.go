package client

import (
	"github.com/rrborja/brute/client/attribs"
)

type Tag struct {
	Name HtmlTag
	Attributes []TagAttr
	SelfEnd bool
}



type P interface {
	Description() (attribs.Id, attribs.Class)
	Attribs() []Tag
	Content() Html
}

type HtmlTag string


var (
	div HtmlTag = "div"
)
