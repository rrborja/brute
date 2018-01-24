package client

type Tag struct {
	Name HtmlTag
	Attributes []TagAttr
	SelfEnd bool
}

type HtmlTag string

var (
	body HtmlTag = "body"
	div HtmlTag = "div"
	ul HtmlTag = "ul"
	ol HtmlTag = "ol"
	li HtmlTag = "li"
	br HtmlTag = "br"
	header HtmlTag = "header"
	main HtmlTag = "main"
	footer HtmlTag = "footer"
	p HtmlTag = "p"
	h1 HtmlTag = "h1"
	h2 HtmlTag = "h2"
	h3 HtmlTag = "h3"
	h4 HtmlTag = "h4"
	h5 HtmlTag = "h5"
	h6 HtmlTag = "h6"
	a HtmlTag = "a"
)

func NewElement(tag HtmlTag) *Element {
	element := new(Element)
	element.Tag = Tag{Name: tag}
	return element
}

func NewSelfElement(tag HtmlTag, selfTerminate bool) *SelfElement {
	element := new(SelfElement)
	element.Tag = Tag{Name: tag, SelfEnd: selfTerminate}
	return element
}

func NewListElement(tag HtmlTag) *ListElement {
	element := new(ListElement)
	element.Tag = Tag{Name: tag}
	return element
}

func Div() *Element {
	return NewElement(div)
}

func UnorderedList() *ListElement {
	return NewListElement(ul)
}

func OrderedList() *ListElement {
	return NewListElement(ol)
}

func Br() string {
	return NewSelfElement(br, false).Value()
}

func Header() *Element {
	return NewElement(header)
}

func Main() *Element {
	return NewElement(main)
}

func Footer() *Element {
	return NewElement(footer)
}

func P() *Element {
	return NewElement(p)
}

func H1() *Element {
	return NewElement(h1)
}

func H2() *Element {
	return NewElement(h2)
}

func H3() *Element {
	return NewElement(h3)
}
func H4() *Element {
	return NewElement(h4)
}

func H5() *Element {
	return NewElement(h5)
}

func H6() *Element {
	return NewElement(h6)
}

func A(href string) *Element {
	aElement := NewElement(a)
	aElement.Attributes_ = []Attr{Attr{"href", href}}
	return aElement
}