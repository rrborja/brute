package client

import (
	"strings"
	time_ "time"
	"fmt"
)

type Tag struct {
	Name HtmlTag
	Attributes []TagAttr
	SelfEnd bool
}

type HtmlTag string

var (
	selfValue HtmlTag = ""
	body HtmlTag = "body"
	div HtmlTag = "div"
	ul HtmlTag = "ul"
	ol HtmlTag = "ol"
	li HtmlTag = "li"
	br HtmlTag = "br"
	img HtmlTag = "img"
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
	em HtmlTag = "em"
	article HtmlTag = "article"
	strong HtmlTag = "strong"
	section HtmlTag = "section"
	code HtmlTag = "code"
	samp HtmlTag = "samp"
	kbd HtmlTag = "kbd"
	var_ HtmlTag = "var"
	nav HtmlTag = "nav"
	aside HtmlTag = "aside"
	figure HtmlTag = "figure"
	figcaption HtmlTag = "figcaption"
	mark HtmlTag = "mark"
	summary HtmlTag = "summary"
	details HtmlTag = "details"
	address HtmlTag = "address"
	abbr HtmlTag = "abbr"
	time HtmlTag = "time"
	form HtmlTag = "form"
	input HtmlTag = "input"
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

func NewFormElement() *FormElement {
	element := new(FormElement)
	element.Tag = Tag{Name: form}
	return element
}

func Value(value interface{}) string {
	element := new(Element)
	element.Tag = Tag{Name: selfValue}
	return element.Value(value)
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

func Break() string {
	return NewSelfElement(br, false).Value()
}

func Image(src string, alt ...string) string {
	element := NewSelfElement(img, false)
	if len(alt) > 0 {
		element.Attributes_ = []Attr {
			{name: "alt", value: strings.Join(alt, "")},
		}
	}
	return element.Value()
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

func Paragraph() *Element {
	return NewElement(p)
}

func Strong() *Element {
	return NewElement(strong)
}

func Emphasis() *Element {
	return NewElement(em)
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

func Article() *Element {
	return NewElement(article)
}

func Section() *Element {
	return NewElement(section)
}

func Code() *Element {
	return NewElement(code)
}

func SampleOutput() *Element {
	return NewElement(samp)
}

func KeyboardInput() *Element {
	return NewElement(kbd)
}

func Variable() *Element {
	return NewElement(var_)
}

func Navigation() *Element {
	return NewElement(nav)
}

func Aside() *Element {
	return NewElement(aside)
}

func Figure() *Element {
	return NewElement(figure)
}

func FigureCaption() *Element {
	return NewElement(figcaption)
}

func Mark() *Element {
	return NewElement(mark)
}

func Summary() *Element {
	return NewElement(summary)
}

func Details(open ...bool) *Element {
	element := NewElement(details)
	if len(open) > 0 && open[0] {
		element.Attributes_ = []Attr{
			{"open", true},
		}
	}
	return element
}

func Time(datetime *time_.Time, layout ...string) *Element {
	if datetime == nil {
		return NewElement(time)
	}
	finalLayout := "2006-01-02 15:04:05.999"
	if len(layout) > 0 {
		finalLayout = strings.Join(layout, "")
	}
	element := NewElement(time)
	element.Attributes_ = []Attr{
		{"datetime", datetime.Format(finalLayout)},
	}
	return element
}

func Address() *Element {
	return NewElement(address)
}

func Abbreviation(title string) *Element {
	element := NewElement(abbr)
	element.Attributes_ = []Attr{
		{"title", title},
	}
	return element
}

func Comment(comment string) string {
	return Value(fmt.Sprintf("<!--%s//-->", comment))
}

func Form() *FormElement {
	return NewFormElement()
}

func Submit(value string, classes ...string) *SubmitElement {
	button := NewElement(HtmlTag("button"))

	if len(classes) > 0 {
		button.Attributes_ = append(button.Attributes_, Attr{
			name: "class",
			value: strings.Join(classes, " "),
		})
	}
	button.Content = value

	e := SubmitElement(*button)
	return &e
}

func Input() *Element {
	return NewElement(input)
}

func inputElement(type_ string, name string, value ...string) string {
	e := NewSelfElement(input, false)
	e.Attributes_ = []Attr{
		{"type", type_},
		{"name", name},
	}
	if len(value) > 0 {
		e.Attributes_ = append(e.Attributes_, Attr{
			"value", strings.Join(value, ""),
		})
	}
	return e.Value()
}

func TextField(name string, value ...string) string {
	return inputElement("text", name, value...)
}

func PasswordField(name string, value ...string) string {
	return inputElement("password", name, value...)
}

func Radio(name string, value string) string {
	return inputElement("radio", name, value)
}

func Checkbox(name string, value string) string {
	return inputElement("checkbox", name, value)
}

func ColorPicker(name string, value string) string {
	return inputElement("color", name, value)
}

// values[0] is min
// values[1] is max
// values[2] is value
func minMaxInputElement(type_ string, name string, values ...string) string {
	e := NewSelfElement(input, false)
	e.Attributes_ = []Attr{
		{"type", type_},
		{"name", name},
	}
	if len(values) >= 1 && values[0] != "" {
		e.Attributes_ = append(e.Attributes_, Attr{
			"min", values[0],
		})
	}
	if len(values) >= 2 && values[1] != "" {
		e.Attributes_ = append(e.Attributes_, Attr{
			"max", values[1],
		})
	}
	if len(values) >= 3 && values[2] != "" {
		e.Attributes_ = append(e.Attributes_, Attr{
			"value", values[2],
		})
	}
	return e.Value()
}

func DatePicker(name string, values ...string) string {
	return minMaxInputElement("date", name, values...)
}

func DateTimePicker(name string, value string) string {
	return inputElement("datetime-local", name, value)
}

func EmailField(name string, value string) string {
	return inputElement("email", name, value)
}

func MonthDatePicker(name string, value string) string {
	return inputElement("month", name, value)
}

func NumberPicker(name string, values ...string) string {
	return minMaxInputElement("number", name, values...)
}

func RangePicker(name string, values ...string) string {
	return minMaxInputElement("range", name, values...)
}

func SearchField(name string, value string) string {
	return inputElement("search", name, value)
}

func TimeField(name string, value string) string {
	return inputElement("time", name, value)
}

func UrlField(name string, value string) string {
	return inputElement("url", name, value)
}

func WeekField(name string, value string) string {
	return inputElement("week", name, value)
}

// TODO: Make global function like Enabled(func()) and Disabled(func()) to auto mark fields enabled or disabled respectively
// TODO: Lacking chain method design for Input Elements other attributes like steps
// TODO: Substitute HTML elements with alternatives according to compatible "User-Agent"