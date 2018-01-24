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

type Element struct {
	Tag
	Id_ *attribs.Id
	Class_ []attribs.Class
	Attributes_ []Attr
	Content interface{}
}

type AfterIdElement Element

func (element *Element) Id(value string) *AfterIdElement {
	id := attribs.Id(value)
	element.Id_ = &id
	e := AfterIdElement(*element)
	return &e
}

func (element Element) Class(value string) *AfterIdElement {
	e := AfterIdElement(element)
	return e.Class(value)
}

func (element Element) Value(value interface{}) string {
	e := AfterIdElement(element)
	return e.Value(value)
}

func (element *AfterIdElement) Class(value string) *AfterIdElement {
	class := attribs.Class(value)
	element.Class_ = append(element.Class_, class)
	return element
}

func (element *AfterIdElement) Attributes(attribs ...Attr) *AfterIdElement {
	element.Attributes_ = attribs
	return element
}

type content struct {
	value interface{}
	attribs func() (Tag, *attribs.Id, []attribs.Class, []Attr)
}

func (c content) String() string {
	return fmt.Sprintf("%s", c.value)
}

// Union-like C++ Equivalent
type RenderStack struct {
	begin string
	content string
	end string

	next *RenderStack
}

var root = new(RenderStack)

func attribPack(tag Tag, id *attribs.Id, classes []attribs.Class, attrs []Attr) func() (Tag, *attribs.Id, []attribs.Class, []Attr) {
	return func() (Tag, *attribs.Id, []attribs.Class, []Attr) {
		return tag, id, classes, attrs
	}
}

type ListElement Element

func (element ListElement) Value(items ...interface{}) string {
	e := AfterIdElement(element)
	return e.Value(func() {
		for _, item := range items {
			element := new(Element)
			element.Tag = Tag{Name: li}
			element.Value(item)
		}
	})
}

type RenderElement struct {
	*AfterIdElement
}

func (element *AfterIdElement) Value(value interface{}) string {
	element.Content = value
	return evaluate(Element(*element))
}

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
	return fmt.Sprintf(`%s="%s"`, attr.Name(), attr.Value())
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


func Div() *Element {
	element := new(Element)
	element.Tag = Tag{Name: div}
	return element
}

func UnorderedList() *ListElement {
	element := new(ListElement)
	element.Tag = Tag{Name: ul}
	return element
}

func OrderedList() *ListElement {
	element := new(ListElement)
	element.Tag = Tag{Name: ol}
	return element
}

//func Item() *Element {
//
//}

func Escape(content string) string {
	content = strings.Replace(content,"<", "&lt;", -1)
	content = strings.Replace(content, ">", "&gt;", -1)
	content = strings.Replace(content, "&", "&amp;", -1)
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
		initial = strings.Join(append([]string{initial}, entry), " ")
	}

	return fmt.Sprintf("<%s%s%s>", tag.Name, initial, joinAttrs(attribs...))
}

func renderEndTag(tag Tag) string {
	return "</" + string(tag.Name) + ">"
}

func evaluate(element Element) string {
	tag := element.Tag
	id := element.Id_
	classes := element.Class_
	attribs := element.Attributes_

	begin := renderBeginTag(tag, id, classes, attribs)
	end := renderEndTag(tag)

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
		result := begin + Escape(fmt.Sprintf("%s", val)) + end
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

func A(attr ...Attr) []Attr {
	return attr
}
