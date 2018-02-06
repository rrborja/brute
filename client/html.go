package client

import (
	"github.com/rrborja/brute/client/attribs"
	"github.com/rrborja/brute/client/html/meta"
	"errors"
	"github.com/rrborja/brute/client/html/meta/mime"
	"fmt"
	"strings"
	"github.com/silentred/gid"
	"html"
	"io"
	"sync"
)

//var title *string
//
//var headElements = make([]interface{}, 0)
//
//var metaInfo *meta.MetaInfo
//var metaCharset *Attr

var (
	MetaInformationAlreadySuppliedError = errors.New("basic meta information already supplied")
	CharsetAlreadyDefinedError = errors.New("charset already defined")
)

var htmlStream chan string

var sessionWriter map[int64]*RenderStackHolder
var sessionWriterMutex = sync.RWMutex{}

type Html struct {
	Head Element
	Body Element
}

type RenderStackHolder struct {
	title *string

	headElements []interface{}

	metaInfo *meta.MetaInfo
	metaCharset *Attr

	body interface{}

	root *RenderStack
	writer io.Writer
}

type RenderStack struct {
	begin string
	content string
	end string

	next *RenderStack
}

//var root = new(RenderStack)

type TagAttr interface {
	Name() string
	Value() interface{}
	String() string
}

type Attr struct {
	name string
	value interface{}
}

func NewAttr(name string, value interface{}) *Attr {
	return &Attr{name, value}
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
	sessionWriter = make(map[int64]*RenderStackHolder)
}

func Charset(charSet CharSet) (err error) {
	sessionId := gid.Get()
	renderStackHolder, ok := Writer(sessionId)

	if !ok {
		return fmt.Errorf("cannot set page's charset since the session %v expired or does not exist", sessionId)
	}

	metaCharset := &renderStackHolder.metaCharset

	if *metaCharset != nil {
		err = CharsetAlreadyDefinedError
	}
	*metaCharset = &Attr{name: "charset", value: charSet}
	return
}

func Title(name string) (err error) {
	sessionId := gid.Get()
	renderStackHolder, ok := Writer(sessionId)

	if !ok {
		return fmt.Errorf("cannot set page's title since the session %v expired or does not exist", sessionId)
	}

	title := &renderStackHolder.title

	if *title != nil {
		return
	}

	*title = &name
	addHeadElements(Element {
		Tag: Tag {
			Name: "title",
		},
		Content: name,
	})

	return
}

func Author(name string) {
	Meta("author", name)
}

func Description(value string) {
	if len(value) > 160 {
		value = value[:157] + "..."
	}
	Meta("description", value)
}

func Copyright(value string) {
	Meta("copyright", value)
}

func Meta(metaType string, content string) {
	addHeadElements(SelfElement {
		Tag: Tag {
			Name: "meta",
			SelfEnd: true,
		},
		Attributes_: []TagAttr{
			&Attr{"name", metaType},
			&Attr{"content",  content},
		},
	})
}

func PreloadStylesheet(href string) {
	addHeadElements(Element{
		Tag: Tag {
			Name: "link",
			SelfEnd: true,
		},
		Attributes_: []TagAttr{
			&Attr{"rel", "stylesheet"},
			&Attr{"type", mime.TextCss},
			&Attr{"href", href},
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
	addHeadElements(Element{
		Tag: Tag {
			Name: "script",
			Attributes: attr,
		},
		Attributes_: attr,
	})
}

func EmbedScript(code string, async ...bool) {
	attr := make([]TagAttr, 0)
	if len(async) > 0 && async[0] {
		attr = append(attr, &SelfAttr{"async"})
	}
	// TODO
}

func addHeadElements(element interface{}) (err error) {
	sessionId := gid.Get()
	renderStackHolder, ok := Writer(sessionId)

	if !ok {
		return fmt.Errorf("cannot set page's head content since the session %v expired or does not exist", sessionId)
	}

	headElements := &renderStackHolder.headElements

	*headElements = append(*headElements, element)

	return
}

func Escape(content string) string {
	return strings.Replace(html.EscapeString(content), "\n", " ", -1)
}

type syntax struct {
	context interface{}
	children []syntax
}

func Writer(sessionId int64) (renderStackHolder *RenderStackHolder, ok bool) {
	sessionWriterMutex.RLock()
	renderStackHolder, ok = sessionWriter[sessionId]
	sessionWriterMutex.RUnlock()
	return
}

func SetWriter(sessionId int64, renderStackHolder *RenderStackHolder) {
	sessionWriterMutex.Lock()
	sessionWriter[sessionId] = renderStackHolder
	sessionWriterMutex.Unlock()
}

func SetContentType(writer io.Writer, mimeName mime.Mime) {
	writer.Write([]byte("~ct" + mimeName))
}

func checkNewBody() (renderStackHolder *RenderStackHolder, ok bool) {
	id := gid.Get()

	renderStackHolder, ok = Writer(id)
	if !ok {
		panic("Expected a writer. Endpoint crashed.")
	}

	if renderStackHolder.body == nil {
		defer func() { renderStackHolder.body = new(interface{}) }()

		writer := renderStackHolder.writer

		SetContentType(writer, mime.TextHtml)

		writer.Write([]byte("<!DOCTYPE html><html><head>"))
		for _, _headElement := range renderStackHolder.headElements {
			headElement := _headElement
			switch v := headElement.(type) {
			case Element: renderStackHolder.evaluate(v)
			case SelfElement: renderStackHolder.evaluate(Element(v))
			}
		}
		writer.Write([]byte("</head><body>"))
	}

	return
}

func NewElement(tag HtmlTag) *Element {

	renderStackHolder, _ := checkNewBody()

	element := new(Element)
	element.Tag = Tag{Name: tag}
	element.stack = renderStackHolder

	return element
}

func renderBeginTag(tag Tag, id *attribs.Id, class []attribs.Class, attribs []TagAttr) string {
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

func (stack *RenderStackHolder) evaluate(element Element) string {
	tag := element.Tag
	id := element.Id_
	classes := element.Class_
	attribs := element.Attributes_

	var begin, end string
	if tag.Name != selfValue {
		begin = renderBeginTag(tag, id, classes, attribs)
		if tag.SelfEnd {
			stack.writer.Write([]byte(begin))
			return begin
		}
		end = renderEndTag(tag)
	}

	switch val := element.Content.(type) {
	case func():
		stack.root.begin = begin
		stack.root.next = new(RenderStack)

		prev := stack.root
		stack.root = stack.root.next

		stack.writer.Write([]byte(begin))

		val()

		stack.writer.Write([]byte(end))

		result := begin + stack.root.content + end

		stack.root = prev
		stack.root.end = begin
		stack.root.content =  string(append([]byte(stack.root.content), result...))

		return result
	default:
		result := begin + Escape(fmt.Sprintf("%v", val)) + end
		stack.root.content = string(append([]byte(stack.root.content), result...))

		stack.writer.Write([]byte(result))

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

func joinAttrs(attrs ...TagAttr) string {
	values := make([]string, len(attrs))
	for i, attr := range attrs {
		values[i] = attr.String()
	}
	return strings.Join(values[:], " ")
}