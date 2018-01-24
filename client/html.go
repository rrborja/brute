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

func (element *AfterIdElement) Class(value string) *AfterIdElement {
	class := attribs.Class(value)
	element.Class_ = append(element.Class_, class)
	return element
}

func (element *AfterIdElement) Attributes(attribs ...Attr) *AfterIdElement {
	element.Attributes_ = attribs
	return element
}

var root = &RenderStack{child: make([]*RenderStack, 0)}

type content struct {
	value interface{}
	attribs func() (Tag, *attribs.Id, []attribs.Class, []Attr)
}

func (c content) String() string {
	return fmt.Sprintf("%s", c.value)
}

// Union-like C++ Equivalent
type RenderStack struct {
	*content

	child []*RenderStack
}

func attribPack(tag Tag, id *attribs.Id, classes []attribs.Class, attrs []Attr) func() (Tag, *attribs.Id, []attribs.Class, []Attr) {
	return func() (Tag, *attribs.Id, []attribs.Class, []Attr) {
		return tag, id, classes, attrs
	}
}

func Queue(value RenderElement) {
	switch val := value.Content.(type) {
	case func():
		prev := root
		root = &RenderStack{child: make([]*RenderStack, 0)}
		val()
		prev.child = root.child
		root = prev
	default:
		root.child = append(root.child, &RenderStack{content: &content{value.Content,
		attribPack(value.Tag, value.Id_, value.Class_, value.Attributes_),
		}})
	}
}

type RenderElement struct {
	*AfterIdElement
}

func (element *AfterIdElement) Value(value interface{}) *RenderElement {
	element.Content = value
	e := RenderElement{element}
	Queue(e)
	return &e
}

func (element *RenderElement) Render() string {
	return evaluate(element.Tag, element.Id_, element.Class_, element.Attributes_)
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

func Escape(content string) string {
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

	return fmt.Sprintf("<%s %s %s>", tag.Name, initial, joinAttrs(attribs...))
}

func renderEndTag(tag Tag) string {
	return "</" + string(tag.Name) + ">"
}

func evaluate(tag Tag, id *attribs.Id, class []attribs.Class, attribs []Attr) string {
	begin := renderBeginTag(tag, id, class, attribs)
	end := renderEndTag(tag)

	if root.child != nil {
		value := make([]string, len(root.child))
		for _, node := range root.child {
			prev := root
			root = node
			value = append(value, evaluate(node.attribs()))
			root = prev
		}
		return begin + strings.Join(value, "") + end
	} else {
		return begin + fmt.Sprintf("%s", *(root.content)) + end
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
