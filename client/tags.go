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
	body HtmlTag = "body"
	div HtmlTag = "div"
	ul HtmlTag = "ul"
	ol HtmlTag = "ol"
	li HtmlTag = "li"
)