package client

import (
	"github.com/rrborja/brute/client/attribs"
	"github.com/rrborja/brute/client/html/meta"
	"errors"
	"github.com/rrborja/brute/client/html/meta/mime"
	"fmt"
	"strings"
)

var html Html

var title string

var headElements = make([]Element, 0)
var bodyElements = make([]Element, 0)

var metaInfo *meta.MetaInfo
var metaCharset *Attr

var (
	MetaInformationAlreadySuppliedError = errors.New("basic meta information already supplied")
	CharsetAlreadyDefinedError = errors.New("charset already defined")
)

type Html struct {
	Head Element
	Body Element
}

type RenderStack struct {
	begin string
	content string
	end string

	next *RenderStack
}

var root = new(RenderStack)

type TagAttr interface {
	Name() string
	Value() interface{}
}

type Attr struct {
	name string
	value interface{}
}

func (attr *Attr) Name() string {
	return attr.name
}

func (attr *Attr) Value() interface{} {
	return attr.value
}

func (attr Attr) String() string {
	switch v := attr.value.(type) {
	case bool: if v { return fmt.Sprintf(`%s`, attr.Name()) }
	default: return fmt.Sprintf(`%s="%s"`, attr.Name(), attr.Value())
	}
	return ""
}

type SelfAttr struct {
	name string
}

func (selfAttr *SelfAttr) Name() string {
	return selfAttr.name
}

func (selfAttr *SelfAttr) Value() interface{} {
	return nil
}

func (selfAttr SelfAttr) String() string {
	return selfAttr.Name()
}

func init() {
	html = Html{}
}

func PageDescription(author string, description string, keywords ...string) (err error) {
	if metaInfo != nil {
		err = MetaInformationAlreadySuppliedError
	}
	metaInfo = &meta.MetaInfo{Author:author, Description:description, Keywords:keywords}
	return
}

func Charset(charSet CharSet) (err error) {
	if metaCharset != nil {
		err = CharsetAlreadyDefinedError
	}
	metaCharset = &Attr{name: "charset", value: charSet}
	return
}

func Title(name string) {
	title = name
}

func Meta(metaType attribs.Value, content string) {
	addHeadElements(Element {
		Tag: Tag {
			Name: "meta",
			Attributes: []TagAttr{
				&Attr{metaType.String(), content},
			},
			SelfEnd: true,
		},
	})
}

func PreloadStylesheet(href string) {
	addHeadElements(Element {
		Tag: Tag {
			Name: "link",
			Attributes: []TagAttr{
				&Attr{"rel", "stylesheet"},
				&Attr{"type", mime.TextCss},
				&Attr{"href", href},
			},
			SelfEnd: true,
		},
	})
}

func PreloadScript(href string, async ...bool) {
	attr := []TagAttr{
		&Attr{"src", href},
	}
	if len(async) > 0 && async[0] {
		attr = append(attr, &SelfAttr{"async"})
	}
	addHeadElements(Element {
		Tag: Tag {
			Name: "script",
			Attributes: attr,
		},
	})
}

func EmbedScript(code string, async ...bool) {
	attr := make([]TagAttr, 0)
	if len(async) > 0 && async[0] {
		attr = append(attr, &SelfAttr{"async"})
	}
	// TODO
}

func addHeadElements(element Element) {
	headElements = append(headElements, element)
}

func Escape(content string) string {
	content = strings.Replace(content, "&", "&amp;", -1)
	content = strings.Replace(content,"<", "&lt;", -1)
	content = strings.Replace(content, ">", "&gt;", -1)
	return content
}

type syntax struct {
	context interface{}
	children []syntax
}

func renderBeginTag(tag Tag, id *attribs.Id, class []attribs.Class, attribs []Attr) string {
	var initial string

	if id != nil {
		entry := strings.Join([]string{"id", fmt.Sprintf(`"%s"`, *id)}, "=")
		initial = strings.Join(append([]string{initial}, entry), " ")
	}

	if len(class) > 0 {
		entry := strings.Join([]string{"class", fmt.Sprintf(`"%s"`, joinClass(class...))}, "=")
		initial = strings.Join(consolidate(append([]string{initial}, entry)...), " ")
	}

	var selfEnd string

	if tag.SelfEnd {
		selfEnd = "/"
	}

	return fmt.Sprintf("<%s>",
		strings.Join(consolidate(string(tag.Name), initial, joinAttrs(attribs...), selfEnd), " "))
}

func consolidate(values ...string) []string {
	finalValue := make([]string, len(values))
	var counter int
	for _, value := range values {
		if len(value) > 0 {
			finalValue[counter] = value
			counter ++
		}
	}
	return finalValue[:counter]
}

func renderEndTag(tag Tag) string {
	return "</" + string(tag.Name) + ">"
}

func evaluate(element Element) string {
	tag := element.Tag
	id := element.Id_
	classes := element.Class_
	attribs := element.Attributes_

	var begin, end string
	if tag.Name != selfValue {
		begin = renderBeginTag(tag, id, classes, attribs)
		end = renderEndTag(tag)
	}

	switch val := element.Content.(type) {
	case func():
		root.begin = begin
		root.next = new(RenderStack)

		prev := root
		root = root.next

		val()

		result := begin + root.content + end

		root = prev
		root.end = begin
		root.content =  string(append([]byte(root.content), result...))

		return result
	default:
		result := begin + Escape(fmt.Sprintf("%v", val)) + end
		root.content = string(append([]byte(root.content), result...))

		return result
	}
}

func joinClass(classes ...attribs.Class) string {
	values := make([]string, len(classes))
	for i, class := range classes {
		values[i] = string(class)
	}
	return strings.Join(values[:], " ")
}

func joinAttrs(attrs ...Attr) string {
	values := make([]string, len(attrs))
	for i, attr := range attrs {
		values[i] = attr.String()
	}
	return strings.Join(values[:], " ")
}